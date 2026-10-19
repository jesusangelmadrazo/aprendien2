import { useEffect, useState } from "react";
import { getPendingExams, reviewExam } from "../../api/exams";
import type { PendingExam } from "../../types";

const LABELS: Record<string, string> = { theory: "Teórico", practice: "Práctico" };

export default function ExamsReviewPage() {
  const [pending, setPending] = useState<PendingExam[]>([]);
  const [loading, setLoading] = useState(true);
  const [busyId, setBusyId] = useState<string | null>(null);

  function refresh() {
    setLoading(true);
    getPendingExams()
      .then(setPending)
      .finally(() => setLoading(false));
  }

  useEffect(refresh, []);

  async function handleReview(id: string, approve: boolean) {
    let reason = "";
    if (!approve) {
      reason = prompt("Motivo del rechazo:") ?? "";
    }
    setBusyId(id);
    try {
      await reviewExam(id, approve, reason);
      refresh();
    } finally {
      setBusyId(null);
    }
  }

  return (
    <div className="page">
      <h1>Exámenes pendientes de revisión</h1>
      {loading && <p>Cargando...</p>}
      {!loading && pending.length === 0 && <p className="hint">No hay exámenes esperando revisión.</p>}

      <ul className="list-card">
        {pending.map((e) => (
          <li key={e.ID} className="review-exam-row">
            <div>
              <div className="list-row-title">
                {e.student_name} — {LABELS[e.Type] ?? e.Type}
              </div>
              <div className="hint">Enviado {new Date(e.SubmittedAt).toLocaleString("es-UY")}</div>
              <a href={e.PhotoURL} target="_blank" rel="noreferrer">
                Ver foto
              </a>
            </div>
            <div className="review-exam-actions">
              <button className="btn-secondary btn-small" disabled={busyId === e.ID} onClick={() => handleReview(e.ID, false)}>
                Rechazar
              </button>
              <button className="btn-small" disabled={busyId === e.ID} onClick={() => handleReview(e.ID, true)}>
                Aprobar
              </button>
            </div>
          </li>
        ))}
      </ul>
    </div>
  );
}
