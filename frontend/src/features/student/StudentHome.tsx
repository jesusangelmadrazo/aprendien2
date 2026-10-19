import { useEffect, useState } from "react";
import { useAuthStore } from "../../store/authStore";
import { useLessonStore } from "../../store/lessonStore";
import { getNearbyInstructors, getAllInstructors } from "../../api/instructors";
import { createLesson } from "../../api/lessons";
import MapView from "./MapView";
import SlotSearch from "./SlotSearch";
import type { InstructorSummary, NearbyInstructor } from "../../types";

export default function StudentHome() {
  const user = useAuthStore((s) => s.user);
  const currentLesson = useLessonStore((s) => s.currentLesson);
  const clearCurrentLesson = useLessonStore((s) => s.clearCurrentLesson);

  const [center, setCenter] = useState<{ lat: number; lon: number } | null>(null);
  const [pickup, setPickup] = useState<{ lat: number; lon: number } | null>(null);
  const [instructors, setInstructors] = useState<NearbyInstructor[]>([]);
  const [allInstructors, setAllInstructors] = useState<InstructorSummary[]>([]);
  const [requesting, setRequesting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    navigator.geolocation?.getCurrentPosition(
      (pos) => {
        const here = { lat: pos.coords.latitude, lon: pos.coords.longitude };
        setCenter(here);
        setPickup(here);
      },
      () => setCenter({ lat: -34.9011, lon: -56.1645 }) // fallback: Montevideo centro
    );
  }, []);

  useEffect(() => {
    if (!center) return;
    const refresh = () => getNearbyInstructors(center.lat, center.lon).then(setInstructors).catch(() => {});
    refresh();
    const interval = setInterval(refresh, 5000);
    return () => clearInterval(interval);
  }, [center]);

  // Independent of the live map: an instructor doesn't need to be online
  // right now for a student to book a future slot with them.
  useEffect(() => {
    getAllInstructors().then(setAllInstructors).catch(() => {});
  }, []);

  async function handleRequestLesson() {
    if (!pickup) return;
    setError(null);
    setRequesting(true);
    try {
      await createLesson({ pickup_lat: pickup.lat, pickup_lon: pickup.lon, duration_minutes: 60 });
      useLessonStore.getState().setCurrentLesson({
        lessonRequestId: "pending",
        status: "searching",
      });
    } catch (err: any) {
      setError(err.response?.data?.error ?? "No se pudo pedir la clase");
    } finally {
      setRequesting(false);
    }
  }

  return (
    <div className="page">
      <h1>Hola, {user?.full_name}</h1>

      <section className="on-demand-card">
        <h2>Pedir clase ahora</h2>
        <p className="hint">
          Buscamos al instructor online más cercano en este momento. El mapa te sirve acá para ajustar dónde te
          encontrás con él — el día de la clase, además, vas a poder seguirlo llegar en vivo.
        </p>

        {center && (
          <MapView
            center={center}
            instructors={instructors}
            pickup={pickup}
            onPickupChange={(lat, lon) => setPickup({ lat, lon })}
          />
        )}

        {currentLesson?.status === "searching" && <p className="status-banner">Buscando instructor cerca tuyo…</p>}
        {currentLesson?.status === "accepted" && (
          <div className="status-banner status-ok">
            <p>
              ¡Instructor asignado! <strong>{currentLesson.instructorName}</strong> — {currentLesson.vehicle}
            </p>
            <button onClick={clearCurrentLesson}>Ok</button>
          </div>
        )}
        {currentLesson?.status === "no_instructor_found" && (
          <div className="status-banner status-error">
            <p>No encontramos instructores disponibles cerca. Probá de nuevo en un rato.</p>
            <button onClick={clearCurrentLesson}>Ok</button>
          </div>
        )}

        {error && <p className="form-error">{error}</p>}

        {(!currentLesson || currentLesson.status === "no_instructor_found") && (
          <button onClick={handleRequestLesson} disabled={requesting || !pickup}>
            {requesting ? "Pidiendo..." : "Pedir clase ahora"}
          </button>
        )}
      </section>

      <SlotSearch instructors={allInstructors} pickup={pickup} />

      <section className="requirements-card">
        <h2>Requisitos para sacar la libreta (categoría A, primera vez) en Montevideo</h2>
        <ul>
          <li>Cédula de identidad o pasaporte uruguayo vigente</li>
          <li>Certificado médico de una institución habilitada</li>
          <li><strong>15 horas de clases prácticas</strong> antes de rendir el examen práctico</li>
          <li>Agendar cita: agenda electrónica, WhatsApp 099 019 500 o teléfono 1950 4000</li>
          <li>Trámite en Av. 18 de Julio 1360 — costo aproximado $2.360</li>
        </ul>
        <p className="hint">Fuente: tramites.montevideo.gub.uy</p>
      </section>
    </div>
  );
}
