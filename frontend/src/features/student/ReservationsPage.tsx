import { useEffect, useState } from "react";
import { listMyLessons, cancelLesson } from "../../api/lessons";
import { getAllInstructors } from "../../api/instructors";
import type { LessonRequest } from "../../types";

const UPCOMING_STATUSES = new Set(["accepted", "instructor_enroute", "in_progress", "searching", "requested"]);
const DATE_FORMATTER = new Intl.DateTimeFormat("es-UY", { weekday: "short", day: "numeric", month: "short", hour: "2-digit", minute: "2-digit" });

const STATUS_LABEL: Record<string, string> = {
  requested: "Solicitada",
  searching: "Buscando instructor",
  accepted: "Confirmada",
  instructor_enroute: "Instructor en camino",
  in_progress: "En curso",
};

export default function ReservationsPage() {
  const [lessons, setLessons] = useState<LessonRequest[]>([]);
  const [instructorNames, setInstructorNames] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  function refresh() {
    setLoading(true);
    listMyLessons()
      .then((all) => setLessons(all.filter((l) => UPCOMING_STATUSES.has(l.Status))))
      .catch(() => setError("No se pudieron cargar las reservas"))
      .finally(() => setLoading(false));
  }

  useEffect(refresh, []);

  useEffect(() => {
    getAllInstructors().then((list) => {
      const map: Record<string, string> = {};
      list.forEach((i) => (map[i.UserID] = i.FullName));
      setInstructorNames(map);
    });
  }, []);

  async function handleCancel(id: string) {
    if (!confirm("¿Seguro que querés cancelar esta reserva?")) return;
    await cancelLesson(id, "Cancelada por el estudiante");
    refresh();
  }

  return (
    <div className="page">
      <h1>Mis reservas</h1>
      {loading && <p>Cargando...</p>}
      {error && <p className="form-error">{error}</p>}
      {!loading && lessons.length === 0 && <p className="hint">No tenés reservas próximas.</p>}

      <ul className="list-card">
        {lessons.map((l) => (
          <li key={l.ID} className="list-row">
            <div>
              <div className="list-row-title">
                {l.ScheduledAt ? DATE_FORMATTER.format(new Date(l.ScheduledAt)) : "Pedido en tiempo real"}
              </div>
              <div className="hint">
                {l.InstructorID ? instructorNames[l.InstructorID] ?? "Instructor asignado" : "Buscando instructor"} ·{" "}
                {STATUS_LABEL[l.Status] ?? l.Status} · ${l.EstimatedPriceUYU} UYU
              </div>
            </div>
            <button className="btn-secondary btn-small" onClick={() => handleCancel(l.ID)}>
              Cancelar
            </button>
          </li>
        ))}
      </ul>
    </div>
  );
}
