import { NavLink, Outlet } from "react-router-dom";

const TABS = [
  { to: "/instructor", label: "Inicio", end: true },
  { to: "/instructor/documentos", label: "Documentos" },
  { to: "/instructor/reservas", label: "Reservas" },
  { to: "/instructor/clases", label: "Clases dadas" },
  { to: "/instructor/wallet", label: "Wallet" },
  { to: "/instructor/cobro", label: "Medio de cobro" },
];

export default function InstructorLayout() {
  return (
    <div>
      <nav className="sub-nav">
        {TABS.map((tab) => (
          <NavLink
            key={tab.to}
            to={tab.to}
            end={tab.end}
            className={({ isActive }) => "sub-nav-link" + (isActive ? " active" : "")}
          >
            {tab.label}
          </NavLink>
        ))}
      </nav>
      <Outlet />
    </div>
  );
}
