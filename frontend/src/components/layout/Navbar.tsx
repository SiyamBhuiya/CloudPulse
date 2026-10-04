import { useAuth } from "../../hooks/useAuth";

export default function Navbar() {
  const { user, logout } = useAuth();

  return (
    <header className="flex items-center justify-end gap-4 border-b border-line px-6 py-3">
      <span className="text-sm text-muted">{user?.email}</span>
      <button onClick={logout} className="text-sm font-semibold hover:underline">
        Sign out
      </button>
    </header>
  );
}