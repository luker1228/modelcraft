'use client'

import * as React from 'react'
import { Copy, Check } from 'lucide-react'

import { Button } from '@/web/components/ui/button'
import { cn } from '@/shared/utils'
import type { JsonExpr } from '@/types/rls'
import type { PolicyJSONPreviewProps } from './types'

/**
 * JSON 语法高亮组件
 *
 * 简单的语法高亮，不使用外部库
 */
function HighlightedJSON({ value }: { value: JsonExpr | null }) {
  const jsonString = React.useMemo(() => {
    if (value === null) return 'null'
    return JSON.stringify(value, null, 2)
  }, [value])

  // 简单的语法高亮：将 JSON 字符串分段渲染
  const renderHighlighted = (str: string): React.ReactNode[] => {
    const parts: React.ReactNode[] = []
    let key = 0

    // 按行分割
    const lines = str.split('\n')

    lines.forEach((line, lineIndex) => {
      const lineParts: React.ReactNode[] = []
      let remaining = line

      // 匹配键名（双引号包围的字符串后跟冒号）
      const keyRegex = /"([^"]+)":/g
      let match
      let lastIndex = 0

      // eslint-disable-next-line no-cond-assign
      while ((match = keyRegex.exec(line)) !== null) {
        // 添加匹配前的内容
        if (match.index > lastIndex) {
          const before = line.slice(lastIndex, match.index)
          lineParts.push(
            <span key={`${key++}`} className="text-foreground">
              {before}
            </span>
          )
        }

        // 添加键名（紫色）
        lineParts.push(
          <span key={`${key++}`}>
            <span className="text-purple-600 dark:text-purple-400">&quot;{match[1]}&quot;</span>
            <span className="text-foreground">:</span>
          </span>
        )

        lastIndex = match.index + match[0].length
      }

      // 添加剩余内容
      if (lastIndex < line.length) {
        const remainingStr = line.slice(lastIndex)

        // 字符串值（绿色）
        const stringRegex = /"([^"]*)"/g
        const strParts: React.ReactNode[] = []
        let strMatch
        let strLastIndex = 0
        let strKey = 0

        // eslint-disable-next-line no-cond-assign
        while ((strMatch = stringRegex.exec(remainingStr)) !== null) {
          if (strMatch.index > strLastIndex) {
            strParts.push(
              <span key={`str-${strKey++}`} className="text-foreground">
                {remainingStr.slice(strLastIndex, strMatch.index)}
              </span>
            )
          }

          strParts.push(
            <span key={`str-${strKey++}`} className="text-green-600 dark:text-green-400">
              &quot;{strMatch[1]}&quot;
            </span>
          )

          strLastIndex = strMatch.index + strMatch[0].length
        }

        if (strLastIndex < remainingStr.length) {
          const finalStr = remainingStr.slice(strLastIndex)

          // 数字（蓝色）
          const numParts: React.ReactNode[] = []
          const numRegex = /-?\d+(?:\.\d+)?/g
          let numMatch
          let numLastIndex = 0
          let numKey = 0

          // eslint-disable-next-line no-cond-assign
          while ((numMatch = numRegex.exec(finalStr)) !== null) {
            if (numMatch.index > numLastIndex) {
              const beforeNum = finalStr.slice(numLastIndex, numMatch.index)
              // 布尔值和 null（红色）
              if (/\b(true|false|null)\b/.test(beforeNum)) {
                numParts.push(
                  <span
                    key={`num-${numKey++}`}
                    className="text-red-600 dark:text-red-400"
                    dangerouslySetInnerHTML={{
                      __html: beforeNum.replace(
                        /\b(true|false|null)\b/g,
                        '<span class="text-red-600 dark:text-red-400">$1</span>'
                      ),
                    }}
                  />
                )
              } else {
                numParts.push(
                  <span key={`num-${numKey++}`} className="text-foreground">
                    {beforeNum}
                  </span>
                )
              }
            }

            numParts.push(
              <span key={`num-${numKey++}`} className="text-blue-600 dark:text-blue-400">
                {numMatch[0]}
              </span>
            )

            numLastIndex = numMatch.index + numMatch[0].length
          }

          if (numLastIndex < finalStr.length) {
            const finalPart = finalStr.slice(numLastIndex)
            numParts.push(
              <span key={`num-${numKey++}`} className="text-foreground">
                {finalPart}
              </span>
            )
          }

          if (numParts.length > 0) {
            strParts.push(...numParts)
          } else {
            // 处理布尔值和 null
            if (/\b(true|false|null)\b/.test(finalStr)) {
              strParts.push(
                <span
                  key={`str-${strKey++}`}
                  dangerouslySetInnerHTML={{
                    __html: finalStr.replace(
                      /\b(true|false|null)\b/g,
                      '<span class="text-red-600 dark:text-red-400">$1</span>'
                    ),
                  }}
                />
              )
            } else {
              strParts.push(
                <span key={`str-${strKey++}`} className="text-foreground">
                  {finalStr}
                </span>
              )
            }
          }
        }

        if (strParts.length > 0) {
          lineParts.push(...strParts)
        } else {
          lineParts.push(
            <span key={`${key++}`} className="text-foreground">
              {remainingStr}
            </span>
          )
        }
      }

      parts.push(
        <div key={lineIndex} className="font-mono text-sm leading-relaxed">
          {lineParts.length > 0 ? lineParts : line}
        </div>
      )
    })

    return parts
  }

  return (
    <pre className="overflow-x-auto p-4 font-mono text-sm">
      {renderHighlighted(jsonString)}
    </pre>
  )
}

/**
 * Policy JSON 预览组件
 *
 * 格式化展示 JSON 表达式，支持语法高亮和复制
 */
export function PolicyJSONPreview({ value, className }: PolicyJSONPreviewProps) {
  const [copied, setCopied] = React.useState(false)

  const jsonString = React.useMemo(() => {
    if (value === null) return 'null'
    return JSON.stringify(value, null, 2)
  }, [value])

  const handleCopy = React.useCallback(async () => {
    try {
      await navigator.clipboard.writeText(jsonString)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      // 复制失败时静默处理
    }
  }, [jsonString])

  return (
    <div className={cn('relative rounded-md border bg-muted/30', className)}>
      {/* 复制按钮 */}
      <Button
        variant="ghost"
        size="sm"
        className="absolute right-2 top-2 h-8 gap-1.5 text-muted-foreground hover:text-foreground"
        onClick={handleCopy}
      >
        {copied ? (
          <>
            <Check className="size-3.5 text-green-500" />
            <span className="text-xs">已复制</span>
          </>
        ) : (
          <>
            <Copy className="size-3.5" />
            <span className="text-xs">复制</span>
          </>
        )}
      </Button>

      {/* JSON 内容 */}
      <div className="max-h-[400px] overflow-auto">
        <HighlightedJSON value={value} />
      </div>
    </div>
  )
}
