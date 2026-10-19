import { useEffect, useState } from "react";
import { listMyLessons, cancelLesson, startLesson, completeLesson } from "../../api/lessons";
import { getUserName } from "../../api/users";
import type { LessonRequest } from "../../types";

const UPCOMING_STATUSES = new Set(["accepted", "instructor_enroute", "in_progress"]);
const DATE_FORMATTER = new Intl.DateTimeFormat("es-UY", { weekday: "short", day: "numeric", month: "short", hour: "2-digit", minute: "2-digit" });

const STATUS_LABEL: Record<string, string> = {
  accepted: "Confirmada",
  instructor_enroute: "En camino",
  in_progress: "En curso",
};

export default function InstructorReservationsPage() {
  const [lessons, setLessons] = useState<LessonRequest[]>([]);
  const [studentNames, setStudentNames] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(true);
  const [busyId, setBusyId] = useState<string | null>(null);

  function refresh() {
    setLoading(true);
    listMyLessons()
      .then((all) => setLessons(all.filter((l) => UPCOMING_STATUSES.has(l.Status))))
      .finally(() => setLoading(false));
  }

  useEffect(refresh, []);

  useEffect(() => {
    const missing = [...new Set(lessons.map((l) => l.StudentID))].filter((id) => !(id in studentNames));
    if (missing.length === 0) return;
    Promise.all(missing.map((id) => getUserName(id).then((name) => [id, name] as const))).then((pairs) => {
      setStudentNames((prev) => ({ ...prev, ...Object.fromEntries(pairs) }));
    });
  }, [lessons]);

  async function handleAction(action: (id: string) => Promise<unknown>, id: string) {
    setBusyId(id);
    try {
      await action(id);
      refresh();
    } finally {
      setBusyId(null);
    }
  }

  return (
    <div className="page">
      <h1>Mis reservas</h1>
      {loading && <p>Cargando...</p>}
      {!loading && lessons.length === 0 && <p className="hint">No tenés reservas próximas.</p>}

      <ul className="list-card">
        {lessons.map((l) => (
          <li key={l.ID} className="list-row">
            <div>
              <div className="list-row-title">
                {l.ScheduledAt ? DATE_FORMATTER.format(new Date(l.ScheduledAt)) : "Pedido en tiempo real"}
              </div>
              <div className="hint">
                {studentNames[l.StudentID] ?? "Estudiante"} · {STATUS_LABEL[l.Status] ?? l.Status} · ${l.EstimatedPriceUYU} UYU
              </div>
            </div>
            <div className="review-exam-actions">
              {l.Status === "accepted" && (
                <button className="btn-small" disabled={busyId === l.ID} onClick={() => handleAction(startLesson, l.ID)}>
                  Empezar clase
                </button>
              )}
              {l.Status === "in_progress" && (
                <button className="btn-small" disabled={busyId === l.ID} onClick={() => handleAction(completeLesson, l.ID)}>
                  Completar
                </button>
              )}
              <button className="btn-secondary btn-small" disabled={busyId === l.ID} onClick={() => handleAction((id) => cancelLesson(id, "Cancelada por el instructor"), l.ID)}>
                Cancelar
              </button>
            </div>
          </li>
        ))}
      </ul>
    </div>
  );
}
