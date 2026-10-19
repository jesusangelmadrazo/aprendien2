import { useEffect, useRef, useState } from "react";
import { getMyExamStatus, submitExam } from "../../api/exams";
import type { ExamStatusOverview, ExamType, ExamTypeStatus } from "../../types";

const LABELS: Record<ExamType, string> = { theory: "Examen teórico", practice: "Examen práctico" };
const STATUS_LABEL: Record<string, string> = {
  pending_review: "En revisión",
  approved: "Aprobado",
  rejected: "Rechazado",
};

function ExamCard({ type, status, onSubmitted }: { type: ExamType; status: ExamTypeStatus; onSubmitted: () => void }) {
  const fileRef = useRef<HTMLInputElement>(null);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleUpload() {
    const file = fileRef.current?.files?.[0];
    if (!file) return;
    setError(null);
    setUploading(true);
    try {
      await submitExam(type, file);
      onSubmitted();
    } catch (err: any) {
      setError(err.response?.data?.error ?? "No se pudo subir la foto");
    } finally {
      setUploading(false);
    }
  }

  const submission = status.submission;
  const canSubmit = status.unlocked && (!submission || submission.Status === "rejected");
  const progressLabel = type === "theory" ? `${status.progress}/${status.threshold} clases` : `${(status.progress / 60).toFixed(1)}/${(status.threshold / 60).toFixed(0)} horas`;

  return (
    <div className="exam-card">
      <div className="exam-card-header">
        <h3>{LABELS[type]}</h3>
        <span className={"exam-status-pill" + (status.unlocked ? " unlocked" : "")}>
          {status.unlocked ? "Habilitado" : "Bloqueado"}
        </span>
      </div>
      <p className="hint">{progressLabel}</p>

      {submission && (
        <div className={"exam-submission exam-submission-" + submission.Status}>
          <p>
            Estado: <strong>{STATUS_LABEL[submission.Status] ?? submission.Status}</strong>
          </p>
          {submission.Status === "rejected" && submission.RejectionReason && (
            <p className="hint">Motivo: {submission.RejectionReason}</p>
          )}
          <a href={submission.PhotoURL} target="_blank" rel="noreferrer">
            Ver foto enviada
          </a>
        </div>
      )}

      {!status.unlocked && <p className="hint">Todavía no llegás al mínimo para rendir este examen.</p>}

      {canSubmit && (
        <div className="exam-upload">
          <input type="file" accept="image/*" ref={fileRef} />
          <button className="btn-small" disabled={uploading} onClick={handleUpload}>
            {uploading ? "Subiendo..." : "Subir foto del resultado"}
          </button>
          {error && <p className="form-error">{error}</p>}
        </div>
      )}
    </div>
  );
}

export default function ExamsPage() {
  const [status, setStatus] = useState<ExamStatusOverview | null>(null);
  const [loading, setLoading] = useState(true);

  function refresh() {
    getMyExamStatus()
      .then(setStatus)
      .finally(() => setLoading(false));
  }

  useEffect(refresh, []);

  return (
    <div className="page">
      <h1>Exámenes</h1>
      <p className="hint">
        Subí una foto del resultado que te entregan al rendir — un admin de la plataforma la revisa antes de darla
        por válida, porque ese papel es el que te da la licencia.
      </p>

      {loading && <p>Cargando...</p>}
      {status && (
        <>
          <ExamCard type="theory" status={status.theory} onSubmitted={refresh} />
          <ExamCard type="practice" status={status.practice} onSubmitted={refresh} />
        </>
      )}
    </div>
  );
}
