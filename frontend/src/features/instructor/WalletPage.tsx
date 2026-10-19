import { useEffect, useState } from "react";
import { getMyWallet } from "../../api/payouts";
import type { WalletSummary } from "../../types";

const DATE_FORMATTER = new Intl.DateTimeFormat("es-UY", { day: "numeric", month: "short", year: "numeric" });

export default function WalletPage() {
  const [wallet, setWallet] = useState<WalletSummary | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getMyWallet()
      .then(setWallet)
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="page">
      <h1>Wallet</h1>
      {loading && <p>Cargando...</p>}

      {wallet && (
        <>
          <div className="wallet-summary">
            <div className="wallet-tile">
              <span className="hint">Por cobrar</span>
              <strong>${wallet.pending_uyu} UYU</strong>
            </div>
            <div className="wallet-tile">
              <span className="hint">Cobrado en total</span>
              <strong>${wallet.total_paid_uyu} UYU</strong>
            </div>
          </div>
          <p className="hint">
            El saldo "por cobrar" se liquida semanalmente por transferencia bancaria a tu medio de cobro registrado.
          </p>

          <h2>Historial de liquidaciones</h2>
          {(!wallet.payouts || wallet.payouts.length === 0) && <p className="hint">Todavía no recibiste ninguna liquidación.</p>}
          <ul className="list-card">
            {wallet.payouts?.map((p) => (
              <li key={p.ID} className="list-row">
                <div>
                  <div className="list-row-title">${p.AmountUYU} UYU</div>
                  <div className="hint">
                    Período {DATE_FORMATTER.format(new Date(p.PeriodStart))} — {DATE_FORMATTER.format(new Date(p.PeriodEnd))}
                  </div>
                </div>
              </li>
            ))}
          </ul>
        </>
      )}
    </div>
  );
}
