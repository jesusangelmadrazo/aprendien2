import { NavLink, Outlet } from "react-router-dom";

const TABS = [
  { to: "/admin", label: "Exámenes", end: true },
  { to: "/admin/documentos", label: "Documentos" },
  { to: "/admin/payouts", label: "Liquidaciones" },
];

export default function AdminLayout() {
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
