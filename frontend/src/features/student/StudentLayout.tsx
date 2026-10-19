import { NavLink, Outlet } from "react-router-dom";

const TABS = [
  { to: "/student", label: "Inicio", end: true },
  { to: "/student/reservas", label: "Reservas" },
  { to: "/student/clases", label: "Clases" },
  { to: "/student/pagos", label: "Pagos" },
  { to: "/student/examenes", label: "Exámenes" },
];

export default function StudentLayout() {
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
