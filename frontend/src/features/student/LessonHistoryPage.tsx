import { useEffect, useState } from "react";
import { listMyLessons } from "../../api/lessons";
import { getAllInstructors } from "../../api/instructors";
import type { LessonRequest } from "../../types";

const DATE_FORMATTER = new Intl.DateTimeFormat("es-UY", { day: "numeric", month: "short", year: "numeric", hour: "2-digit", minute: "2-digit" });

export default function LessonHistoryPage() {
  const [lessons, setLessons] = useState<LessonRequest[]>([]);
  const [instructorNames, setInstructorNames] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    listMyLessons("completed")
      .then(setLessons)
      .finally(() => setLoading(false));
    getAllInstructors().then((list) => {
      const map: Record<string, string> = {};
      list.forEach((i) => (map[i.UserID] = i.FullName));
      setInstructorNames(map);
    });
  }, []);

  const totalMinutes = lessons.reduce((sum, l) => sum + l.DurationMinutes, 0);

  return (
    <div className="page">
      <h1>Clases tomadas</h1>
      {!loading && lessons.length > 0 && (
        <p className="hint">
          {lessons.length} clases · {(totalMinutes / 60).toFixed(1)} horas en total
        </p>
      )}
      {loading && <p>Cargando...</p>}
      {!loading && lessons.length === 0 && <p className="hint">Todavía no completaste ninguna clase.</p>}

      <ul className="list-card">
        {lessons.map((l) => (
          <li key={l.ID} className="list-row">
            <div>
              <div className="list-row-title">{l.InstructorID ? instructorNames[l.InstructorID] ?? "Instructor" : "Instructor"}</div>
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
