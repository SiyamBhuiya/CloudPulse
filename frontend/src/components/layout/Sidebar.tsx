import { NavLink } from "react-router-dom";

const links = [
  { to: "/dashboard", label: "Dashboard" },
  { to: "/servers", label: "Servers" },
  { to: "/alerts", label: "Alerts" },
  { to: "/health-checks", label: "Health checks" },
];

export default function Sidebar() {
  return (
    <nav
      aria-label="Main"
      className="flex w-56 shrink-0 flex-col gap-1 border-r border-line p-4 max-md:hidden"
    >
      <div className="mx-2 mb-6 flex items-center gap-2 text-xl font-bold">
        <i className="h-2.5 w-2.5 rounded-full bg-ok ring-4 ring-ok/25" />
        CloudPulse
      </div>
      {links.map((l) => (
        <NavLink
          key={l.to}
          to={l.to}
          className={({ isActive }) =>
            `rounded-lg px-3 py-2 font-medium ${
              isActive
                ? "bg-panel text-ink shadow-[inset_3px_0_0_var(--cpu)]"
                : "text-muted hover:text-ink"
            }`
          }
        >
          {l.label}
        </NavLink>
      ))}
    </nav>
  );
}