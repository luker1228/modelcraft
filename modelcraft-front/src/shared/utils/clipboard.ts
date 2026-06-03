/**
 * Copy text to clipboard with fallback for non-secure contexts (HTTP).
 *
 * Strategy:
 * 1. Try navigator.clipboard (HTTPS / localhost only)
 * 2. On reject or unavailable, fall back to execCommand (works on HTTP)
 */

function execCommandCopy(text: string): boolean {
  const el = document.createElement('textarea')
  el.value = text
  el.style.position = 'fixed'
  el.style.top = '0'
  el.style.left = '0'
  el.style.opacity = '0'
  document.body.appendChild(el)
  el.focus()
  el.select()
  const success = document.execCommand('copy')
  document.body.removeChild(el)
  return success
}

export function copyToClipboard(text: string): void {
  if (navigator.clipboard) {
    void navigator.clipboard.writeText(text).catch(() => {
      execCommandCopy(text)
    })
  } else {
    execCommandCopy(text)
  }
}

export function copyToClipboardWithCallback(
  text: string,
  onSuccess: () => void
): void {
  if (navigator.clipboard) {
    navigator.clipboard.writeText(text).then(onSuccess).catch(() => {
      // Clipboard API rejected (e.g. HTTP context) — fall back to execCommand
      if (execCommandCopy(text)) onSuccess()
    })
  } else {
    if (execCommandCopy(text)) onSuccess()
  }
}

