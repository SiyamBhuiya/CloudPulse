export interface Strength {
  score: 0 | 1 | 2 | 3 | 4;
  hint: string;
}

const hints = [
  "Use 8+ characters with a mix of letters and numbers.",
  "Too weak. Add more characters.",
  "Fair. Add numbers or symbols.",
  "Good. One more step for a strong password.",
  "Strong password.",
];

export function passwordStrength(pw: string): Strength {
  let score = 0;
  if (pw.length) {
    score = 1;
    if (pw.length >= 8) score++;
    if (/[A-Za-z]/.test(pw) && /\d/.test(pw)) score++;
    if (pw.length >= 12 && /[^A-Za-z0-9]/.test(pw)) score++;
  }
  const s = score as Strength["score"];
  return { score: s, hint: hints[s] };
}