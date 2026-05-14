import React, { useRef } from 'react'
import { Button } from '@web/components/ui/button'
import { cn } from '@/shared/utils'

export interface WhereJsonEditorProps {
  value: string
  onChange: (value: string) => void
  onFormat: () => void
  onClear: () => void
  onApply: () => void
  isValid: boolean
}

/**
 * Plain textarea JSON editor with validation status, format, clear, and apply actions.
 *
 * Exposes a ref-based `insertAtCursor` method so parent components (FilterPanel)
 * can programmatically insert text at the current cursor position — used by
 * FieldSchemaPanel when a field is clicked.
 *
 * AI note: to programmatically set the filter, call `onChange` with the full
 * where JSON string, then call `onApply`. No ref manipulation needed from AI.
 */
export interface WhereJsonEditorRef {
  insertAtCursor: (snippet: string) => void
}

export const WhereJsonEditor = React.forwardRef<WhereJsonEditorRef, WhereJsonEditorProps>(
  function WhereJsonEditor({ value, onChange, onFormat, onClear, onApply, isValid }, ref) {
    const textareaRef = useRef<HTMLTextAreaElement>(null)

    React.useImperativeHandle(ref, () => ({
      insertAtCursor(snippet: string) {
        const el = textareaRef.current
        if (!el) return
        const start = el.selectionStart
        const end = el.selectionEnd
        const newValue = value.slice(0, start) + snippet + value.slice(end)
        onChange(newValue)
        // Restore cursor after the inserted snippet
        requestAnimationFrame(() => {
          el.selectionStart = start + snippet.length
          el.selectionEnd = start + snippet.length
          el.focus()
        })
      },
    }))

    const isEmpty = !value.trim()
    const showError = !isEmpty && !isValid

    return (
      <div className="flex flex-1 flex-col gap-2">
        {/* Header row */}
        <div className="flex items-center justify-between">
          <span className="text-xs font-medium text-foreground">Where JSON</span>
          <div className="flex gap-1.5">
            <Button
              variant="outline"
              size="sm"
              className="h-6 px-2 text-xs"
              onClick={onFormat}
              disabled={isEmpty || !isValid}
            >
              格式化
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="h-6 px-2 text-xs"
              onClick={onClear}
              disabled={isEmpty}
            >
              清空
            </Button>
          </div>
        </div>

        {/* Textarea */}
        {/* Code-editor aesthetic — intentionally dark so JSON is visually distinct
            from the surrounding light UI. Using Catppuccin Mocha palette as a one-off exception. */}
        <textarea
          ref={textareaRef}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder={'{\n  "AND": [\n    { "fieldName": { "contains": "value" } }\n  ]\n}'}
          spellCheck={false}
          className={cn(
            'min-h-[120px] w-full resize-y rounded-md border bg-[#1e1e2e] p-3 font-mono text-[11px] leading-relaxed text-[#cdd6f4] placeholder:text-[#6c7086] focus:outline-none focus:ring-1',
            showError ? 'border-destructive focus:ring-destructive' : 'border-border focus:ring-ring'
          )}
        />

        {/* Footer row */}
        <div className="flex items-center justify-between">
          <span
            className={cn(
              'text-[11px]',
              isEmpty
                ? 'text-muted-foreground'
                : isValid
                  ? 'text-green-600'
                  : 'text-destructive'
            )}
          >
            {isEmpty ? '输入 where 条件后点击应用' : isValid ? '✓ 有效 JSON' : '✗ JSON 格式错误'}
          </span>
          <Button
            size="sm"
            className="h-7 px-3 text-xs"
            onClick={onApply}
            disabled={!isValid && !isEmpty}
          >
            应用筛选
          </Button>
        </div>
      </div>
    )
  }
)
