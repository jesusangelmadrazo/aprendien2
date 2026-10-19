import { useEffect, useRef, useState } from "react";
import { getMyDocumentStatus, submitDocument } from "../../api/documents";
import type { DocumentStatusOverview, DocumentType } from "../../types";

const LABELS: Record<DocumentType, string> = {
  license: "Libreta de conducir",
  identity: "Cédula de identidad",
  vehicle_registration: "Propiedad del auto",
  insurance: "Seguro del vehículo",
};

const STATUS_LABEL: Record<string, string> = {
  pending_review: "En revisión",
  approved: "Aprobado",
  rejected: "Rechazado",
};

const NEEDS_EXPIRY: Record<DocumentType, boolean> = {
  license: true,
  identity: true,
  vehicle_registration: false,
  insurance: true,
};

const DOC_ORDER: DocumentType[] = ["license", "identity", "vehicle_registration", "insurance"];

function DocumentCard({
  type,
  overview,
  onSubmitted,
}: {
  type: DocumentType;
  overview: DocumentStatusOverview;
  onSubmitted: () => void;
}) {
  const fileRef = useRef<HTMLInputElement>(null);
  const [expiryDate, setExpiryDate] = useState("");
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const doc = overview.documents[type];
  const canSubmit = !doc || doc.Status === "rejected";

  async function handleUpload() {
    const file = fileRef.current?.files?.[0];
    if (!file) return;
    if (NEEDS_EXPIRY[type] && !expiryDate) {
      setError("La fecha de vencimiento es obligatoria para este documento");
      return;
    }
    setError(null);
    setUploading(true);
    try {
      await submitDocument(type, file, expiryDate || undefined);
      onSubmitted();
    } catch (err: any) {
      setError(err.response?.data?.error ?? "No se pudo subir el documento");
    } finally {
      setUploading(false);
    }
  }

  return (
    <div className="exam-card">
      <div className="exam-card-header">
        <h3>{LABELS[type]}</h3>
        {doc && (
          <span className={"exam-status-pill" + (doc.Status === "approved" ? " unlocked" : "")}>
            {STATUS_LABEL[doc.Status]}
          </span>
        )}
      </div>

      {doc && (
        <div className={"exam-submission exam-submission-" + doc.Status}>
          {doc.ExpiryDate && <p>Vence: {new Date(doc.ExpiryDate).toLocaleDateString("es-UY")}</p>}
          {doc.Status === "rejected" && doc.RejectionReason && <p className="hint">Motivo: {doc.RejectionReason}</p>}
          <a href={doc.PhotoURL} target="_blank" rel="noreferrer">
            Ver foto enviada
          </a>
        </div>
      )}

      {canSubmit && (
        <div className="exam-upload">
          {NEEDS_EXPIRY[type] && (
            <label>
              <span className="field-label">Vencimiento</span>
              <input type="date" value={expiryDate} onChange={(e) => setExpiryDate(e.target.value)} />
            </label>
          )}
          <input type="file" accept="image/*" ref={fileRef} />
          <button className="btn-small" disabled={uploading} onClick={handleUpload}>
            {uploading ? "Subiendo..." : doc ? "Volver a subir" : "Subir documento"}
          </button>
          {error && <p className="form-error">{error}</p>}
        </div>
      )}
    </div>
  );
}

export default function DocumentsPage() {
  const [overview, setOverview] = useState<DocumentStatusOverview | null>(null);
  const [loading, setLoading] = useState(true);

  function refresh() {
    getMyDocumentStatus()
      .then(setOverview)
      .finally(() => setLoading(false));
  }

  useEffect(refresh, []);

  return (
    <div className="page">
      <h1>Documentos</h1>
      <p className="hint">
        Necesitamos verificar tu libreta, tu cédula, la propiedad del vehículo y el seguro antes de que puedas
        ponerte online y que los estudiantes te reserven. Un admin revisa cada uno.
      </p>

      {!loading && overview && (
        <div className={"status-banner" + (overview.verified ? " status-ok" : "")}>
          {overview.verified ? "Estás verificado — ya podés ponerte online." : "Todavía te falta al menos un documento para quedar verificado."}
        </div>
      )}

      {loading && <p>Cargando...</p>}
      {overview &&
        DOC_ORDER.map((type) => <DocumentCard key={type} type={type} overview={overview} onSubmitted={refresh} />)}
    </div>
  );
}
