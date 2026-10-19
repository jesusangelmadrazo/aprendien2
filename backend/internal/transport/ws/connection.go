package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/coder/websocket"
)

// OfferResponder is implemented by the matching service; kept as a narrow
// interface here so this package doesn't depend on the service package.
type OfferResponder interface {
	RespondToOffer(offerID string, accepted bool) error
}

// LocationUpdater is implemented by the instructor service.
type LocationUpdater interface {
	UpdateLocation(ctx context.Context, userID string, lat, lon float64) error
}

// ServeWS upgrades the request, registers the connection under userID, and
// pumps incoming client messages until the socket closes.
func ServeWS(hub *Hub, matcher OfferResponder, instructors LocationUpdater, w http.ResponseWriter, r *http.Request, userID string) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"}, // MVP: single trusted frontend origin is enforced by CORS on the REST API
	})
	if err != nil {
		return
	}
	defer conn.CloseNow()

	hub.Register(userID, conn)
	defer hub.Unregister(userID, conn)

	ctx := r.Context()
	for {
		var msg ClientMessage
		if err := readJSON(ctx, conn, &msg); err != nil {
			return
		}

		switch msg.Type {
		case "location_update":
			var payload LocationUpdatePayload
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				continue
			}
			if err := instructors.UpdateLocation(ctx, userID, payload.Lat, payload.Lon); err != nil {
				slog.Warn("failed to update location", "user_id", userID, "error", err)
			}

		case "lesson_offer_response":
			var payload LessonOfferResponsePayload
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				continue
			}
			if err := matcher.RespondToOffer(payload.OfferID, payload.Accepted); err != nil {
				slog.Info("offer response ignored", "offer_id", payload.OfferID, "error", err)
			}
		}
	}
}

func readJSON(ctx context.Context, conn *websocket.Conn, v interface{}) error {
	_, data, err := conn.Read(ctx)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}
