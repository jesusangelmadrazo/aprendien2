import { useEffect, useState } from "react";
import { getPendingDocuments, reviewDocument } from "../../api/documents";
import type { PendingDocument } from "../../types";

const LABELS: Record<string, string> = {
  license: "Libreta",
  identity: "Cédula",
  vehicle_registration: "Propiedad del auto",
  insurance: "Seguro",
};

export default function DocumentsReviewPage() {
  const [pending, setPending] = useState<PendingDocument[]>([]);
  const [loading, setLoading] = useState(true);
  const [busyId, setBusyId] = useState<string | null>(null);

  function refresh() {
    setLoading(true);
    getPendingDocuments()
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
      await reviewDocument(id, approve, reason);
      refresh();
    } finally {
      setBusyId(null);
    }
  }

  return (
    <div className="page">
      <h1>Documentos pendientes de revisión</h1>
      {loading && <p>Cargando...</p>}
      {!loading && pending.length === 0 && <p className="hint">No hay documentos esperando revisión.</p>}

      <ul className="list-card">
        {pending.map((d) => (
          <li key={d.ID} className="review-exam-row">
            <div>
              <div className="list-row-title">
                {d.instructor_name} — {LABELS[d.Type] ?? d.Type}
              </div>
              <div className="hint">
                Enviado {new Date(d.SubmittedAt).toLocaleString("es-UY")}
                {d.ExpiryDate && <> · Vence {new Date(d.ExpiryDate).toLocaleDateString("es-UY")}</>}
              </div>
              <a href={d.PhotoURL} target="_blank" rel="noreferrer">
                Ver foto
              </a>
            </div>
            <div className="review-exam-actions">
              <button className="btn-secondary btn-small" disabled={busyId === d.ID} onClick={() => handleReview(d.ID, false)}>
                Rechazar
              </button>
              <button className="btn-small" disabled={busyId === d.ID} onClick={() => handleReview(d.ID, true)}>
                Aprobar
              </button>
            </div>
          </li>
        ))}
      </ul>
    </div>
  );
}
