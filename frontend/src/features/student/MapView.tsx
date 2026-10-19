import { MapContainer, TileLayer, Marker, Popup, useMapEvents } from "react-leaflet";
import L from "leaflet";
import "leaflet/dist/leaflet.css";
import type { NearbyInstructor } from "../../types";

// Vite doesn't resolve Leaflet's bundled marker image paths on its own;
// point the default icon at the same assets served from a CDN instead.
delete (L.Icon.Default.prototype as unknown as { _getIconUrl?: unknown })._getIconUrl;
L.Icon.Default.mergeOptions({
  iconRetinaUrl: "https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon-2x.png",
  iconUrl: "https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon.png",
  shadowUrl: "https://unpkg.com/leaflet@1.9.4/dist/images/marker-shadow.png",
});

interface Props {
  center: { lat: number; lon: number };
  instructors: NearbyInstructor[];
  pickup?: { lat: number; lon: number } | null;
  onPickupChange?: (lat: number, lon: number) => void;
}

function ClickHandler({ onPickupChange }: { onPickupChange?: (lat: number, lon: number) => void }) {
  useMapEvents({
    click(e) {
      onPickupChange?.(e.latlng.lat, e.latlng.lng);
    },
  });
  return null;
}

export default function MapView({ center, instructors, pickup, onPickupChange }: Props) {
  return (
    <MapContainer
      center={[center.lat, center.lon]}
      zoom={14}
      style={{ height: 360, width: "100%", borderRadius: 10 }}
    >
      <TileLayer
        attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
        url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
      />
      <ClickHandler onPickupChange={onPickupChange} />
      {instructors.map((i) => (
        <Marker key={i.UserID} position={[i.Lat, i.Lon]}>
          <Popup>
            {i.FullName} — {i.VehicleMake} {i.VehicleModel}
            <br />
            ${i.HourlyRateUYU} UYU/hora · {i.DistanceKm.toFixed(1)} km
          </Popup>
        </Marker>
      ))}
      {pickup && (
        <Marker position={[pickup.lat, pickup.lon]}>
          <Popup>Punto de encuentro</Popup>
        </Marker>
      )}
    </MapContainer>
  );
}
