import type { ModelEnumDomainError, ModelEnumErrorCode } from '@/types'

export interface RawGraphQLErrorLike {
  __typename?: string
  code?: string
  message?: string
  suggestion?: string
}

const KNOWN_ERROR_CODES: readonly ModelEnumErrorCode[] = [
  'FIELD_ENUM_SOURCE_CONFLICT',
  'FIELD_FORMAT_IMMUTABLE',
  'FIELD_REFERENCE_IN_USE',
  'UNKNOWN',
]

function isModelEnumErrorCode(value: string): value is ModelEnumErrorCode {
  return KNOWN_ERROR_CODES.includes(value as ModelEnumErrorCode)
}

function normalizeCode(code: string | undefined): ModelEnumErrorCode | undefined {
  if (!code) {
    return undefined
  }

  const normalized = code.trim().toUpperCase()
  return isModelEnumErrorCode(normalized) ? normalized : undefined
}

function normalizeMessage(message: string | undefined, fallback: string): string {
  const normalized = message?.trim()
  return normalized && normalized.length > 0 ? normalized : fallback
}

function normalizeSuggestion(suggestion: string | undefined): string | undefined {
  const normalized = suggestion?.trim()
  return normalized && normalized.length > 0 ? normalized : undefined
}

export function mapModelEnumError(raw: RawGraphQLErrorLike | null | undefined): ModelEnumDomainError | null {
  if (!raw) {
    return null
  }

  const typename = raw.__typename?.trim()
  const code = normalizeCode(raw.code)
  const suggestion = normalizeSuggestion(raw.suggestion)

  if (typename === 'InvalidModelInput') {
    return {
      type: 'InvalidModelInput',
      message: normalizeMessage(raw.message, '模型参数非法。'),
      suggestion,
    }
  }

  if (typename === 'InvalidFieldInput') {
    return {
      type: 'InvalidFieldInput',
      message: normalizeMessage(raw.message, '字段参数非法。'),
      suggestion,
    }
  }

  if (typename === 'FieldEnumSourceConflict' || code === 'FIELD_ENUM_SOURCE_CONFLICT') {
    return {
      type: 'FieldEnumSourceConflict',
      code: 'FIELD_ENUM_SOURCE_CONFLICT',
      message: normalizeMessage(raw.message, '该 source 字段已被占用。'),
      suggestion,
    }
  }

  if (typename === 'FieldFormatImmutable' || code === 'FIELD_FORMAT_IMMUTABLE') {
    return {
      type: 'FieldFormatImmutable',
      code: 'FIELD_FORMAT_IMMUTABLE',
      message: normalizeMessage(raw.message, '字段 format 不可编辑。'),
      suggestion,
    }
  }

  if (typename === 'FieldReferenceInUse' || code === 'FIELD_REFERENCE_IN_USE') {
    return {
      type: 'FieldReferenceInUse',
      code: 'FIELD_REFERENCE_IN_USE',
      message: normalizeMessage(raw.message, '字段仍被引用，无法完成当前操作。'),
      suggestion,
    }
  }

  return {
    type: 'Unknown',
    code: code ?? 'UNKNOWN',
    message: normalizeMessage(raw.message, '发生未知错误，请稍后重试。'),
    suggestion,
  }
}
