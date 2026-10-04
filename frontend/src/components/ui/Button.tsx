import type { ButtonHTMLAttributes } from "react";

export default function Button({
  className = "",
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement>) {
  return (
    <button
      {...props}
      className={`rounded-lg bg-cpu px-4 py-2.5 font-semibold text-[#10222b] transition
        hover:brightness-105 disabled:cursor-not-allowed disabled:opacity-60
        focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-net
        ${className}`}
    />
  );
}