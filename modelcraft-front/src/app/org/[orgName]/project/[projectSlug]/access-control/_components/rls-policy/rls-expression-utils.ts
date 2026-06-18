import type { RlsAction, RlsExprType } from '@/generated/graphql'

export type RlsExpressionKind = 'using' | 'check'
export interface RlsExpressionField {
  name: string
  title?: string | null
}

export interface RlsAuthVariable {
  name: string
  source?: string | null
  type?: string | null
}

export type RlsCompletionRoot = 'row' | 'input' | 'auth'

export interface RlsCompletionContext {
  root: RlsCompletionRoot
  query: string
  replaceStart: number
  replaceEnd: number
}

export interface RlsCompletionItem {
  key: string
  value: string
  label: string
  description?: string
}

interface RlsExpressionHelp {
  placeholder: string
  example: string
  availableFields: string[]
  rootLabel: 'row' | 'input'
}

type SyntaxResult =
  | { valid: true; empty: boolean }
  | { valid: false; empty: false; message: string }

const AUTH_FIELD_PRIORITY = [
  'owner_id',
  'userid',
  'username',
  'end_user_id',
  'creator_id',
  'created_by',
  'created_by_id',
]

export const FIXED_RLS_AUTH_VARIABLES: RlsAuthVariable[] = [
  { name: 'userid', source: 'X-MC-Auth-Userid-Str / X-MC-Auth-Userid-Int', type: 'string' },
  { name: 'username', source: 'X-MC-Auth-Username', type: 'string' },
  { name: 'roles', source: 'X-MC-Auth-Roles', type: 'string[]' },
]

export function validateRlsExpressionSyntax(value: string): SyntaxResult {
  const trimmed = value.trim()
  if (!trimmed) return { valid: true, empty: true }

  if (trimmed.startsWith('{')) {
    return {
      valid: false,
      empty: false,
      message: '请输入 CEL 表达式，例如 row.owner_id == auth.userid',
    }
  }

  if (trimmed === 'true' || trimmed === 'false') {
    return { valid: true, empty: false }
  }

  const hasAllowedRoot = /\b(row|input|auth)\.[A-Za-z_][A-Za-z0-9_]*/.test(trimmed)
  if (!hasAllowedRoot) {
    return {
      valid: false,
      empty: false,
      message: '表达式需要引用 row、input 或 auth',
    }
  }

  return { valid: true, empty: false }
}

export function getRlsExpressionType(action: RlsAction, kind: RlsExpressionKind): RlsExprType {
  if (kind === 'check') {
    return action === 'update' ? 'UPDATE_CHECK' : 'INSERT_CHECK'
  }

  if (action === 'update') return 'UPDATE_PREDICATE'
  if (action === 'delete') return 'DELETE_PREDICATE'
  return 'SELECT_PREDICATE'
}

export function shouldShowRlsUsingExpression(action: RlsAction): boolean {
  return action === 'read' || action === 'update' || action === 'delete'
}

export function shouldShowRlsCheckExpression(action: RlsAction): boolean {
  return action === 'create' || action === 'update'
}

export function buildRlsExpressionHelp(
  kind: RlsExpressionKind,
  fields: RlsExpressionField[],
): RlsExpressionHelp {
  const rootLabel = kind === 'using' ? 'row' : 'input'
  const availableFields = fields.map((field) => `${rootLabel}.${field.name}`)
  const authField = pickAuthField(fields)
  const example = authField ? `${rootLabel}.${authField.name} == auth.userid` : 'true'

  return {
    placeholder: `例如：${example}`,
    example,
    availableFields,
    rootLabel,
  }
}

export function extractRlsCompletionContext(
  value: string,
  cursorPosition: number,
): RlsCompletionContext | null {
  const beforeCursor = value.slice(0, cursorPosition)
  const match = /\b(row|input|auth)\.([A-Za-z0-9_]*)$/.exec(beforeCursor)
  if (!match || match.index === undefined) return null

  return {
    root: match[1] as RlsCompletionRoot,
    query: match[2] ?? '',
    replaceStart: match.index,
    replaceEnd: cursorPosition,
  }
}

export function buildRlsCompletionItems(input: {
  context: RlsCompletionContext
  rootLabel: 'row' | 'input'
  fields: RlsExpressionField[]
  authVariables: RlsAuthVariable[]
}): RlsCompletionItem[] {
  const { context, rootLabel, fields, authVariables } = input
  const normalizedQuery = context.query.toLowerCase()

  if (context.root === 'auth') {
    return authVariables
      .filter((variable) => variable.name.toLowerCase().includes(normalizedQuery))
      .map((variable) => ({
        key: `auth.${variable.name}`,
        value: `auth.${variable.name}`,
        label: `auth.${variable.name}`,
        description: [variable.type, variable.source].filter(Boolean).join(' · ') || undefined,
      }))
  }

  if (context.root !== rootLabel) return []

  return fields
    .filter((field) => field.name.toLowerCase().includes(normalizedQuery))
    .map((field) => ({
      key: `${rootLabel}.${field.name}`,
      value: `${rootLabel}.${field.name}`,
      label: `${rootLabel}.${field.name}`,
      description: field.title ?? undefined,
    }))
}

export function getRlsAvailableContexts(rootLabel: 'row' | 'input'): Array<{
  key: RlsCompletionRoot
  value: `${RlsCompletionRoot}.`
  description: string
}> {
  const primaryDescription =
    rootLabel === 'row' ? '当前记录行上下文（Using Filter）' : '本次写入输入上下文（Input Check）'

  return [
    {
      key: rootLabel,
      value: `${rootLabel}.`,
      description: primaryDescription,
    },
    {
      key: 'auth',
      value: 'auth.',
      description: '当前认证身份上下文',
    },
  ]
}

function pickAuthField(fields: RlsExpressionField[]): RlsExpressionField | undefined {
  const normalized = fields.map((field) => ({
    field,
    lowerName: field.name.toLowerCase(),
  }))

  for (const candidate of AUTH_FIELD_PRIORITY) {
    const matched = normalized.find(({ lowerName }) => lowerName === candidate)
    if (matched) return matched.field
  }

  return normalized.find(({ lowerName }) => lowerName.includes('owner') || lowerName.includes('user'))?.field
}
