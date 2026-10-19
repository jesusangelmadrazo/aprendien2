import { useEffect, useState } from "react";
import { getPendingBalances, createPayout } from "../../api/payouts";
import type { InstructorBalance } from "../../types";

export default function PayoutsPage() {
  const [balances, setBalances] = useState<InstructorBalance[]>([]);
  const [loading, setLoading] = useState(true);
  const [busyId, setBusyId] = useState<string | null>(null);

  function refresh() {
    setLoading(true);
    getPendingBalances()
      .then(setBalances)
      .finally(() => setLoading(false));
  }

  useEffect(refresh, []);

  async function handlePay(instructorId: string) {
    if (!confirm("¿Confirmás que ya le transferiste el dinero a este instructor?")) return;
    setBusyId(instructorId);
    try {
      await createPayout(instructorId);
      refresh();
    } finally {
      setBusyId(null);
    }
  }

  return (
    <div className="page">
      <h1>Liquidaciones pendientes</h1>
      <p className="hint">
        Transferí el dinero por fuera de la app (según el medio de cobro que cargó cada instructor) y después marcalo
        acá como pagado.
      </p>

      {loading && <p>Cargando...</p>}
      {!loading && balances.length === 0 && <p className="hint">No hay saldos pendientes.</p>}

      <ul className="list-card">
        {balances.map((b) => (
          <li key={b.instructor_id} className="list-row">
            <div>
              <div className="list-row-title">{b.instructor_name}</div>
              <div className="hint">${b.amount_uyu} UYU pendientes</div>
            </div>
            <button className="btn-small" disabled={busyId === b.instructor_id} onClick={() => handlePay(b.instructor_id)}>
              {busyId === b.instructor_id ? "..." : "Marcar como pagado"}
            </button>
          </li>
        ))}
      </ul>
    </div>
  );
}
