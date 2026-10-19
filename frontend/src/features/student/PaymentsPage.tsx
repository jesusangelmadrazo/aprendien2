import { useEffect, useState } from "react";
import { getMyPayments } from "../../api/payments";
import type { Payment } from "../../types";

const DATE_FORMATTER = new Intl.DateTimeFormat("es-UY", { day: "numeric", month: "short", year: "numeric" });

const STATUS_LABEL: Record<string, string> = {
  pending: "Pendiente",
  processing: "Procesando",
  approved: "Aprobado",
  rejected: "Rechazado",
  refunded: "Reembolsado",
};

export default function PaymentsPage() {
  const [payments, setPayments] = useState<Payment[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getMyPayments()
      .then(setPayments)
      .finally(() => setLoading(false));
  }, []);

  const totalPaid = payments.filter((p) => p.Status === "approved").reduce((sum, p) => sum + p.AmountUYU, 0);

  return (
    <div className="page">
      <h1>Mis pagos</h1>
      {!loading && payments.length > 0 && <p className="hint">Total pagado: ${totalPaid} UYU</p>}
      {loading && <p>Cargando...</p>}
      {!loading && payments.length === 0 && <p className="hint">Todavía no hiciste ningún pago.</p>}

      <ul className="list-card">
        {payments.map((p) => (
          <li key={p.ID} className="list-row">
            <div>
              <div className="list-row-title">${p.AmountUYU} UYU</div>
              <div className="hint">
                {DATE_FORMATTER.format(new Date(p.CreatedAt))} · instructor ${p.InstructorPriceUYU} + fee ${p.PlatformFeeUYU} ·{" "}
                {STATUS_LABEL[p.Status] ?? p.Status}
              </div>
            </div>
          </li>
        ))}
      </ul>
    </div>
  );
}
