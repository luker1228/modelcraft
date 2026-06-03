/**
 * Copy text to clipboard with fallback for non-secure contexts (HTTP).
 *
 * `navigator.clipboard` is only available in Secure Contexts (HTTPS / localhost).
 * On plain HTTP, it is `undefined` and calling `.writeText()` throws a TypeError.
 * This helper falls back to the legacy `document.execCommand('copy')` API.
 */
export function copyToClipboard(text: string): void {
  if (navigator.clipboard) {
    void navigator.clipboard.writeText(text)
    return
  }

  // Fallback for non-secure contexts
  const el = document.createElement('textarea')
  el.value = text
  el.style.position = 'fixed'
  el.style.opacity = '0'
  document.body.appendChild(el)
  el.focus()
  el.select()
  document.execCommand('copy')
  document.body.removeChild(el)
}

/**
 * Copy text to clipboard and invoke `onSuccess` callback when done.
 * Handles both the async Clipboard API and the synchronous execCommand fallback.
 */
export function copyToClipboardWithCallback(
  text: string,
  onSuccess: () => void
): void {
  if (navigator.clipboard) {
    void navigator.clipboard.writeText(text).then(onSuccess)
    return
  }

  // Fallback for non-secure contexts
  const el = document.createElement('textarea')
  el.value = text
  el.style.position = 'fixed'
  el.style.opacity = '0'
  document.body.appendChild(el)
  el.focus()
  el.select()
  document.execCommand('copy')
  document.body.removeChild(el)
  onSuccess()
}
