import { useState, type FormEvent } from "react";
import { Link, Navigate, useNavigate } from "react-router-dom";

import Button from "../components/ui/Button";
import Input from "../components/ui/Input";
import { useAuth } from "../hooks/useAuth";
import { ApiError } from "../services/api";
import { passwordStrength } from "../utils/passwordStrength";

const barColors = ["", "bg-bad", "bg-cpu", "bg-mem", "bg-ok"];

const steps = [
  ["Create your account", "Takes under a minute."],
  ["Add a server", "You get a personal agent key."],
  ["Start the agent", "Live metrics show up in seconds."],
];

export default function Register() {
  const { register, isAuthenticated } = useAuth();
  const navigate = useNavigate();

  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const strength = passwordStrength(password);

  if (isAuthenticated) return <Navigate to="/dashboard" replace />;

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);

    if (password.length < 8) {
      setError("Password must be at least 8 characters.");
      return;
    }

    setLoading(true);
    try {
      await register({ name, email, password });
      navigate("/dashboard", { replace: true });
    } catch (err) {
      setError(
        err instanceof ApiError && err.status === 409
          ? "An account with this email already exists."
          : "Could not create your account. Check your connection and try again."
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
          <i className="h-2.5 w-2.5 rounded-full bg-ok shadow-[0_0_0_4px_color-mix(in_srgb,var(--ok)_25%,transparent)]" />
          CloudPulse
        </div>

        <div className="max-md:hidden">
          <h2 className="max-w-[15ch] text-[34px] font-bold leading-tight tracking-tight">
            Up and running in three steps
          </h2>
          <ol className="mt-6 flex flex-col gap-4">
            {steps.map(([title, text], i) => (
              <li key={title} className="flex items-start gap-3.5">
                <b className="grid h-7 w-7 flex-none place-items-center rounded-full border border-line text-[13px] font-bold text-cpu">
                  {i + 1}
                </b>
                <div>
                  <strong className="block font-semibold">{title}</strong>
                  <span className="text-sm text-muted">{text}</span>
                </div>
              </li>
            ))}
          </ol>
        </div>

        <div
          aria-hidden
          className="overflow-x-auto whitespace-nowrap rounded-xl border border-line bg-bg px-4 py-3.5 font-mono text-[13px] font-medium text-mem max-md:hidden"
        >
          $ ./cloudpulse-agent --key cp_live_8f3a…
        </div>
      </aside>

      {/* Form */}
      <main className="flex items-center justify-center p-6 md:p-10">
        <form onSubmit={onSubmit} className="w-full max-w-sm">
          <h1 className="text-[28px] font-bold tracking-tight">Create your account</h1>
          <p className="mb-7 mt-1 text-muted">Free to start. No credit card needed.</p>

          {error && (
            <div
              role="alert"
              className="mb-4 rounded-lg border border-bad bg-[color-mix(in_srgb,var(--bad)_12%,transparent)] px-3 py-2.5 text-sm"
            >
              {error}
            </div>
          )}

          <Input
            label="Name"
            autoComplete="name"
            placeholder="Ana Rahman"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
          />
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
            autoComplete="new-password"
            placeholder="At least 8 characters"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />

          {/* Strength meter sits under the password field */}
          <div className="-mt-2 mb-5">
            <div className="grid grid-cols-4 gap-1.5" aria-hidden>
              {[1, 2, 3, 4].map((n) => (
                <i
                  key={n}
                  className={`h-1 rounded-sm transition-colors motion-reduce:transition-none ${
                    strength.score >= n ? barColors[strength.score] : "bg-line"
                  }`}
                />
              ))}
            </div>
            <p className="mt-1.5 min-h-5 text-[13px] text-muted" aria-live="polite">
              {strength.hint}
            </p>
          </div>

          <Button type="submit" disabled={loading} className="w-full">
            {loading ? "Creating account..." : "Create account"}
          </Button>

          <p className="mt-5 text-center text-muted">
            Already have an account?{" "}
            <Link to="/login" className="font-semibold text-ink">
              Sign in
            </Link>
          </p>
        </form>
      </main>
    </div>
  );
}