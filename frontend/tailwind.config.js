/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        bg: "var(--bg)",
        panel: "var(--panel)",
        line: "var(--line)",
        ink: "var(--text)",
        muted: "var(--muted)",
        cpu: "var(--cpu)",
        mem: "var(--mem)",
        net: "var(--net)",
        ok: "var(--ok)",
        bad: "var(--bad)",
      },
      fontFamily: {
        sans: ['"Schibsted Grotesk"', "system-ui", "sans-serif"],
      },
    },
  },
  plugins: [],
};