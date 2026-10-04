import { useState, type FormEvent } from "react";
import { Link, Navigate, useLocation, useNavigate } from "react-router-dom";

import Button from "../components/ui/Button";
import Input from "../components/ui/Input";
import { useAuth } from "../hooks/useAuth";
import { ApiError } from "../services/api";

export default function Login() {
  const { login, isAuthenticated } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const from = (location.state as { from?: { pathname: string } } | null)?.from?.pathname ?? "/dashboard";

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  if (isAuthenticated) return <Navigate to={from} replace />;

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      await login({ email, password });
      navigate(from, { replace: true });
    } catch (err) {
      setError(
        err instanceof ApiError && err.status === 401
          ? "Email or password is incorrect. Check both and try again."
          : "Could not sign in. Check your connection and try again."
      );
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="grid min-h-screen md:grid-cols-[1.1fr_1fr]">
      {/* Brand panel */}
      <aside className="flex flex-col justify-between gap-8 border-b border-line bg-panel p-6 md:border-b-0 md:border-r md:p-12">
        <div className="flex items-center gap-2 text-xl font-bold">
          <i className="h-2.5 w-2.5 rounded-full bg-ok ring-4 ring-ok/25" />
          CloudPulse
        </div>

        <div className="max-md:hidden">
          <h2 className="max-w-[15ch] text-[34px] font-bold leading-tight tracking-tight">
            See your servers the second they change
          </h2>
          <p className="mt-3 max-w-[38ch] text-muted">
            Connect a server, install the agent, and watch CPU, memory and uptime live.
          </p>
        </div>

        <div>
          <svg viewBox="0 0 600 120" preserveAspectRatio="none" className="h-[70px] w-full md:h-[120px]" aria-hidden>
            <path
              d="M0 80 L70 80 L95 80 L115 30 L140 105 L160 80 L260 80 L285 80 L305 20 L335 108 L355 80 L460 80 L485 80 L505 36 L530 100 L550 80 L600 80"
              fill="none"
              stroke="var(--cpu)"
              strokeWidth="3"
              strokeLinejoin="round"
              strokeLinecap="round"
              vectorEffect="non-scaling-stroke"
            />
          </svg>
          <div className="mt-5 flex flex-wrap gap-7 max-md:hidden">
            {[
              ["Uptime", "99.98%"],
              ["Servers watched", "1,240"],
              ["Alerts sent today", "86"],
            ].map(([label, value]) => (
              <div key={label}>
                <small className="block text-sm text-muted">{label}</small>
                <strong className="text-xl font-bold">{value}</strong>
              </div>
            ))}
          </div>
        </div>
      </aside>

      {/* Form */}
      <main className="flex items-center justify-center p-6 md:p-10">
        <form onSubmit={onSubmit} className="w-full max-w-sm">
          <h1 className="text-[28px] font-bold tracking-tight">Sign in</h1>
          <p className="mb-7 mt-1 text-muted">
            Welcome back. Enter your details to open your dashboard.
          </p>

          {error && (
            <div
              role="alert"
              className="mb-4 rounded-lg border border-bad bg-bad/10 px-3 py-2.5 text-sm"
            >
              {error}
            </div>
          )}

          <Input
            label="Email"
            type="email"
            autoComplete="email"
            placeholder="you@company.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
          <Input
            label="Password"
            type="password"
            autoComplete="current-password"
            placeholder="At least 8 characters"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />

          <Button type="submit" disabled={loading} className="mt-1.5 w-full">
            {loading ? "Signing in..." : "Sign in"}
          </Button>

          <p className="mt-5 text-center text-muted">
            New to CloudPulse?{" "}
            <Link to="/register" className="font-semibold text-ink">
              Create an account
            </Link>
          </p>
        </form>
      </main>
    </div>
  );
}