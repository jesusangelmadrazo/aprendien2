import { useEffect, useState } from "react";
import { getMyPayoutMethod, setMyPayoutMethod } from "../../api/payouts";
import type { AccountType } from "../../types";

export default function PayoutMethodPage() {
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [bankName, setBankName] = useState("");
  const [accountType, setAccountType] = useState<AccountType>("caja_ahorro");
  const [accountNumber, setAccountNumber] = useState("");
  const [holderName, setHolderName] = useState("");
  const [holderDocument, setHolderDocument] = useState("");

  useEffect(() => {
    getMyPayoutMethod()
      .then((m) => {
        setBankName(m.BankName);
        setAccountType(m.AccountType);
        setAccountNumber(m.AccountNumber);
        setHolderName(m.AccountHolderName);
        setHolderDocument(m.AccountHolderDocument);
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setSaving(true);
    setSaved(false);
    try {
      await setMyPayoutMethod({
        bank_name: bankName,
        account_type: accountType,
        account_number: accountNumber,
        account_holder_name: holderName,
        account_holder_document: holderDocument,
      });
      setSaved(true);
    } catch (err: any) {
      setError(err.response?.data?.error ?? "No se pudo guardar el medio de cobro");
    } finally {
      setSaving(false);
    }
  }

  if (loading) return <div className="page"><p>Cargando...</p></div>;

  return (
    <div className="page">
      <h1>Medio de cobro</h1>
      <p className="hint">A esta cuenta te transferimos semanalmente lo que vas generando con tus clases.</p>

      <form onSubmit={handleSubmit} className="auth-form">
        <label>
          Banco
          <input value={bankName} onChange={(e) => setBankName(e.target.value)} required />
        </label>
        <label>
          Tipo de cuenta
          <select value={accountType} onChange={(e) => setAccountType(e.target.value as AccountType)}>
            <option value="caja_ahorro">Caja de ahorro</option>
            <option value="cuenta_corriente">Cuenta corriente</option>
          </select>
        </label>
        <label>
          Número de cuenta
          <input value={accountNumber} onChange={(e) => setAccountNumber(e.target.value)} required />
        </label>
        <label>
          Titular de la cuenta
          <input value={holderName} onChange={(e) => setHolderName(e.target.value)} required />
        </label>
        <label>
          Cédula del titular
          <input value={holderDocument} onChange={(e) => setHolderDocument(e.target.value)} required />
        </label>
        {error && <p className="form-error">{error}</p>}
        {saved && <p className="hint">Guardado.</p>}
        <button type="submit" disabled={saving}>
          {saving ? "Guardando..." : "Guardar"}
        </button>
      </form>
    </div>
  );
}
