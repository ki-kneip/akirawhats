import type { ButtonHTMLAttributes } from "react"

type Variant = "primary" | "secondary" | "danger" | "ghost"

const variants: Record<Variant, string> = {
  primary: "bg-green-600 text-white hover:bg-green-700 focus-visible:ring-green-500",
  secondary: "bg-white text-zinc-900 border border-zinc-300 hover:bg-zinc-50 focus-visible:ring-zinc-400",
  danger: "bg-red-600 text-white hover:bg-red-700 focus-visible:ring-red-500",
  ghost: "text-zinc-600 hover:bg-zinc-100 focus-visible:ring-zinc-400",
}

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant
  size?: "sm" | "md"
}

export function Button({
  variant = "primary",
  size = "md",
  className = "",
  children,
  ...props
}: ButtonProps) {
  const sizeClass = size === "sm" ? "px-3 py-1.5 text-sm" : "px-4 py-2 text-sm"
  return (
    <button
      className={`inline-flex items-center justify-center gap-2 rounded-lg font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed ${variants[variant]} ${sizeClass} ${className}`}
      {...props}
    >
      {children}
    </button>
  )
}
