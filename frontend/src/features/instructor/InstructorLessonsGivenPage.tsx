import { useEffect, useState } from "react";
import { listMyLessons } from "../../api/lessons";
import { getUserName } from "../../api/users";
import type { LessonRequest } from "../../types";

const DATE_FORMATTER = new Intl.DateTimeFormat("es-UY", { day: "numeric", month: "short", year: "numeric", hour: "2-digit", minute: "2-digit" });

export default function InstructorLessonsGivenPage() {
  const [lessons, setLessons] = useState<LessonRequest[]>([]);
  const [studentNames, setStudentNames] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    listMyLessons("completed")
      .then(setLessons)
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    const ids = [...new Set(lessons.map((l) => l.StudentID))];
    if (ids.length === 0) return;
    Promise.all(ids.map((id) => getUserName(id).then((name) => [id, name] as const))).then((pairs) => {
      setStudentNames(Object.fromEntries(pairs));
    });
  }, [lessons]);

  const totalMinutes = lessons.reduce((sum, l) => sum + l.DurationMinutes, 0);
  const totalEarned = lessons.reduce((sum, l) => sum + (l.FinalPriceUYU ?? l.EstimatedPriceUYU), 0);

  return (
    <div className="page">
      <h1>Clases dadas</h1>
      {!loading && lessons.length > 0 && (
        <p className="hint">
          {lessons.length} clases · {(totalMinutes / 60).toFixed(1)} horas · ${totalEarned} UYU generados
        </p>
      )}
      {loading && <p>Cargando...</p>}
      {!loading && lessons.length === 0 && <p className="hint">Todavía no diste ninguna clase.</p>}

      <ul className="list-card">
        {lessons.map((l) => (
          <li key={l.ID} className="list-row">
            <div>
              <div className="list-row-title">{studentNames[l.StudentID] ?? "Estudiante"}</div>
              <div className="hint">
                {DATE_FORMATTER.format(new Date(l.ScheduledAt ?? l.RequestedAt))} · {l.DurationMinutes} min
              </div>
            </div>
            <div className="row-price">${l.FinalPriceUYU ?? l.EstimatedPriceUYU} UYU</div>
          </li>
        ))}
      </ul>
    </div>
  );
}
