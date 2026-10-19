package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	liveKey  = "instructors:live"
	metaTTL  = 30 * time.Second
)

type GeoRepo struct {
	client *redis.Client
}

func NewGeoRepo(client *redis.Client) *GeoRepo {
	return &GeoRepo{client: client}
}

// SetLocation records an instructor's current position for live matching.
func (g *GeoRepo) SetLocation(ctx context.Context, instructorID string, lat, lon float64) error {
	if err := g.client.GeoAdd(ctx, liveKey, &redis.GeoLocation{
		Name:      instructorID,
		Latitude:  lat,
		Longitude: lon,
	}).Err(); err != nil {
		return err
	}
	return g.client.Set(ctx, onlineKey(instructorID), "1", metaTTL).Err()
}

// RemoveInstructor takes an instructor out of live search (goes offline/busy).
func (g *GeoRepo) RemoveInstructor(ctx context.Context, instructorID string) error {
	return g.client.ZRem(ctx, liveKey, instructorID).Err()
}

type NearbyResult struct {
	InstructorID string
	DistanceKm   float64
	Lat, Lon     float64
}

// SearchNearby returns online instructors within radiusKm, closest first.
// The GEO sorted set has no per-member TTL, so a client that disconnects
// without explicitly going offline (crash, closed laptop) would otherwise
// stay matchable forever; this filters candidates against their short-lived
// "online" flag (refreshed on every SetLocation call) to catch that case.
// It over-fetches from Redis so that filtering out stale entries still
// leaves up to `count` real candidates.
func (g *GeoRepo) SearchNearby(ctx context.Context, lat, lon, radiusKm float64, count int) ([]NearbyResult, error) {
	res, err := g.client.GeoSearchLocation(ctx, liveKey, &redis.GeoSearchLocationQuery{
		GeoSearchQuery: redis.GeoSearchQuery{
			Longitude:  lon,
			Latitude:   lat,
			Radius:     radiusKm,
			RadiusUnit: "km",
			Sort:       "ASC",
			Count:      count * 3,
		},
		WithCoord: true,
		WithDist:  true,
	}).Result()
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, nil
	}

	pipe := g.client.Pipeline()
	existsCmds := make([]*redis.IntCmd, len(res))
	for i, loc := range res {
		existsCmds[i] = pipe.Exists(ctx, onlineKey(loc.Name))
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}

	out := make([]NearbyResult, 0, count)
	for i, loc := range res {
		if len(out) == count {
			break
		}
		if existsCmds[i].Val() == 0 {
			g.client.ZRem(ctx, liveKey, loc.Name) // stale entry, evict opportunistically
			continue
		}
		out = append(out, NearbyResult{
			InstructorID: loc.Name,
			DistanceKm:   loc.Dist,
			Lat:          loc.Latitude,
			Lon:          loc.Longitude,
		})
	}
	return out, nil
}

func onlineKey(instructorID string) string {
	return "instructor:" + instructorID + ":online"
}
