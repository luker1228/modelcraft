import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatDateSafe(
  date: string | Date | null | undefined,
  options?: Intl.DateTimeFormatOptions,
  locale = "zh-CN",
  fallback = "-"
): string {
  if (date === null || date === undefined) {
    return fallback
  }

  if (typeof date === "string" && date.trim() === "") {
    return fallback
  }

  const parsedDate = date instanceof Date ? date : new Date(date)
  if (Number.isNaN(parsedDate.getTime())) {
    return fallback
  }

  return parsedDate.toLocaleDateString(locale, options)
}

export function formatDate(date: string | Date | null | undefined): string {
  return formatDateSafe(
    date,
    {
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
    },
    "zh-CN"
  )
}
