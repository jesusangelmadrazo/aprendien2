import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { createMyProfile, getMyProfile, updateMyStatus } from "../../api/instructors";
import type { InstructorProfile } from "../../types";
import { useAuthStore } from "../../store/authStore";
import { sendLocationUpdate, onSocketReconnect } from "../../ws/useWebSocket";
import AvailabilityEditor from "./AvailabilityEditor";

const LOCATION_HEARTBEAT_MS = 10_000;

export default function InstructorProfilePage() {
  const user = useAuthStore((s) => s.user);
  const [profile, setProfile] = useState<InstructorProfile | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [vehicleMake, setVehicleMake] = useState("");
  const [vehicleModel, setVehicleModel] = useState("");
  const [vehiclePlate, setVehiclePlate] = useState("");
  const [hourlyRate, setHourlyRate] = useState(800);

  useEffect(() => {
    getMyProfile()
      .then(setProfile)
      .catch(() => setProfile(null))
      .finally(() => setLoading(false));
  }, []);

  async function handleCreateProfile(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    try {
      const created = await createMyProfile({
        vehicle_make: vehicleMake,
        vehicle_model: vehicleModel,
        vehicle_plate: vehiclePlate,
        hourly_rate_uyu: hourlyRate,
      });
      setProfile(created);
    } catch (err: any) {
      setError(err.response?.data?.error ?? "No se pudo crear el perfil");
    }
  }

  // Keeps the instructor matchable: the backend's Redis "online" flag expires
  // after 30s, so a heartbeat well under that keeps them searchable while
  // this tab stays open, and lets them drop out of search automatically if
  // it closes without an explicit "go offline". Also fires right away
  // whenever the socket reconnects (network blip, laptop sleep) instead of
  // waiting up to LOCATION_HEARTBEAT_MS to become matchable again.
  useEffect(() => {
    if (profile?.Status !== "online") return;

    const sendHeartbeat = () => {
      navigator.geolocation?.getCurrentPosition((pos) => {
        sendLocationUpdate(pos.coords.latitude, pos.coords.longitude);
      });
    };

    sendHeartbeat();
    const interval = setInterval(sendHeartbeat, LOCATION_HEARTBEAT_MS);
    const unsubscribe = onSocketReconnect(sendHeartbeat);
    return () => {
      clearInterval(interval);
      unsubscribe();
    };
  }, [profile?.Status]);

  async function toggleOnline() {
    if (!profile) return;
    setError(null);
    const goingOnline = profile.Status !== "online";
    const coords = goingOnline ? await getBrowserLocation() : undefined;
    try {
      await updateMyStatus(goingOnline ? "online" : "offline", coords?.lat, coords?.lon);
      setProfile({ ...profile, Status: goingOnline ? "online" : "offline" });
    } catch (err: any) {
      setError(err.response?.data?.error ?? "No se pudo cambiar el estado");
    }
  }

  if (loading) return <p>Cargando...</p>;

  return (
    <div className="page">
      <h1>Hola, {user?.full_name}</h1>

      {!profile ? (
        <form onSubmit={handleCreateProfile} className="auth-form">
          <h2>Completá tu perfil de instructor</h2>
          <label>
            Marca del vehículo
            <input value={vehicleMake} onChange={(e) => setVehicleMake(e.target.value)} required />
          </label>
          <label>
            Modelo
            <input value={vehicleModel} onChange={(e) => setVehicleModel(e.target.value)} required />
          </label>
          <label>
            Matrícula
            <input value={vehiclePlate} onChange={(e) => setVehiclePlate(e.target.value)} required />
          </label>
          <label>
            Tarifa por hora (UYU)
            <input
              type="number"
              value={hourlyRate}
              onChange={(e) => setHourlyRate(Number(e.target.value))}
              min={1}
              required
            />
          </label>
          {error && <p className="form-error">{error}</p>}
          <button type="submit">Guardar perfil</button>
        </form>
      ) : (
        <div>
          <p>
            {profile.VehicleMake} {profile.VehicleModel} — {profile.VehiclePlate}
          </p>
          <p>Tarifa: ${profile.HourlyRateUYU} UYU/hora</p>
          <p>Estado: {profile.Status}</p>

          {profile.VerificationStatus !== "approved" && (
            <div className="status-banner status-error">
              <p>
                Tu cuenta todavía no fue verificada por un admin — subí tu libreta, cédula verde y seguro en{" "}
                <Link to="/instructor/documentos">Documentos</Link> para poder ponerte online.
              </p>
            </div>
          )}

          {error && <p className="form-error">{error}</p>}

          <button onClick={toggleOnline}>
            {profile.Status === "online" ? "Ponerme offline" : "Ponerme online"}
          </button>

          <AvailabilityEditor />
        </div>
      )}
    </div>
  );
}

function getBrowserLocation(): Promise<{ lat: number; lon: number } | undefined> {
  return new Promise((resolve) => {
    if (!navigator.geolocation) {
      resolve(undefined);
      return;
    }
    navigator.geolocation.getCurrentPosition(
      (pos) => resolve({ lat: pos.coords.latitude, lon: pos.coords.longitude }),
      () => resolve(undefined)
    );
  });
}
