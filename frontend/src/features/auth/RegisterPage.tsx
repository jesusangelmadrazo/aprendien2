import { useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import { register } from "../../api/auth";
import { useAuthStore } from "../../store/authStore";
import type { Role } from "../../types";

export default function RegisterPage() {
  const navigate = useNavigate();
  const setAuth = useAuthStore((s) => s.setAuth);

  const [fullName, setFullName] = useState("");
  const [email, setEmail] = useState("");
  const [phone, setPhone] = useState("");
  const [password, setPassword] = useState("");
  const [role, setRole] = useState<Role>("student");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      const { access_token, user } = await register({
        email,
        phone,
        password,
        full_name: fullName,
        role,
      });
      setAuth(access_token, user);
      navigate(role === "instructor" ? "/instructor" : "/student");
    } catch (err: any) {
      setError(err.response?.data?.error ?? "No se pudo completar el registro");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="auth-page">
      <h1>Crear cuenta</h1>
      <form onSubmit={handleSubmit} className="auth-form">
        <label>
          Nombre completo
          <input value={fullName} onChange={(e) => setFullName(e.target.value)} required />
        </label>
        <label>
          Email
          <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required />
        </label>
        <label>
          Teléfono
          <input value={phone} onChange={(e) => setPhone(e.target.value)} required />
        </label>
        <label>
          Contraseña
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            minLength={8}
            required
          />
        </label>
        <label>
          Quiero
          <select value={role} onChange={(e) => setRole(e.target.value as Role)}>
            <option value="student">Aprender a manejar</option>
            <option value="instructor">Dar clases de manejo</option>
          </select>
        </label>
        {error && <p className="form-error">{error}</p>}
        <button type="submit" disabled={loading}>
          {loading ? "Creando..." : "Crear cuenta"}
        </button>
      </form>
      <p>
        ¿Ya tenés cuenta? <Link to="/login">Iniciá sesión</Link>
      </p>
    </div>
  );
}
