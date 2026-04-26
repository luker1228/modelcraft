import type {
  EnumSourceOption,
  ModelEnumDomainError,
  ModelEnumErrorCode,
  ModelEnumErrorType,
} from '@/types'
import type {
  CreateEnumFieldCommand,
  ModelEnumActionResult,
  ModelEnumContextQuery,
  UpdateFieldMetaCommand,
} from './types'

interface FieldMetaSnapshot {
  title?: string
  description?: string
  validationConfig?: UpdateFieldMetaCommand['validationConfig']
}

interface MockModelEnumState {
  enumSources: EnumSourceOption[]
  fieldFormats: Record<string, 'ENUM' | 'OTHER'>
  fieldMeta: Record<string, FieldMetaSnapshot>
}

const mockStore = new Map<string, MockModelEnumState>()

function getStoreKey({ orgName, projectSlug, modelId }: ModelEnumContextQuery): string {
  return `${orgName}::${projectSlug}::${modelId}`
}

function isBlank(value: string | undefined): boolean {
  return !value || value.trim().length === 0
}

function createDomainError(
  type: ModelEnumErrorType,
  message: string,
  options?: {
    code?: ModelEnumErrorCode
    suggestion?: string
  },
): ModelEnumDomainError {
  return {
    type,
    code: options?.code,
    message,
    suggestion: options?.suggestion,
  }
}

function modelInputError(message: string): ModelEnumDomainError {
  return createDomainError('InvalidInput', message, {
    suggestion: '请确认组织、项目和模型参数均已提供。',
  })
}

function fieldInputError(message: string): ModelEnumDomainError {
  return createDomainError('InvalidInput', message, {
    suggestion: '请检查字段参数是否完整且合法。',
  })
}

function validateContext(query: ModelEnumContextQuery): ModelEnumDomainError | null {
  if (isBlank(query.orgName) || isBlank(query.projectSlug) || isBlank(query.modelId)) {
    return modelInputError('组织、项目或模型参数无效。')
  }

  return null
}

function cloneSources(sources: EnumSourceOption[]): EnumSourceOption[] {
  return sources.map((source) => ({ ...source }))
}

function getOrCreateState(query: ModelEnumContextQuery): MockModelEnumState {
  const key = getStoreKey(query)
  const existed = mockStore.get(key)

  if (existed) {
    return existed
  }

  const seededState: MockModelEnumState = {
    enumSources: [
      {
        fieldName: 'status',
        title: '状态',
        enumName: 'StatusEnum',
        occupied: false,
      },
    ],
    fieldFormats: {
      status: 'ENUM',
    },
    fieldMeta: {},
  }

  mockStore.set(key, seededState)
  return seededState
}

function actionFailed(error: ModelEnumDomainError): ModelEnumActionResult {
  return {
    success: false,
    error,
  }
}

function actionSucceeded(): ModelEnumActionResult {
  return {
    success: true,
    error: null,
  }
}

export async function mockCreateEnumField(command: CreateEnumFieldCommand): Promise<ModelEnumActionResult> {
  const contextError = validateContext(command)
  if (contextError) {
    return actionFailed(contextError)
  }

  if (isBlank(command.name) || isBlank(command.title)) {
    return actionFailed(fieldInputError('字段 name/title 不能为空。'))
  }

  if (isBlank(command.relateEnumName)) {
    return actionFailed(fieldInputError('ENUM 字段创建必须传入 relateEnumName。'))
  }

  const state = getOrCreateState(command)
  const sourceExists = state.enumSources.some((source) => source.fieldName === command.name)

  if (sourceExists) {
    return actionFailed(fieldInputError(`字段 ${command.name} 已存在。`))
  }

  state.enumSources.push({
    fieldName: command.name,
    title: command.title,
    enumName: command.relateEnumName,
    occupied: false,
  })
  state.fieldFormats[command.name] = 'ENUM'
  state.fieldMeta[command.name] = {
    title: command.title,
    description: command.description,
  }

  return actionSucceeded()
}

export async function mockUpdateFieldMeta(command: UpdateFieldMetaCommand): Promise<ModelEnumActionResult> {
  const contextError = validateContext(command)
  if (contextError) {
    return actionFailed(contextError)
  }

  if (isBlank(command.fieldName)) {
    return actionFailed(fieldInputError('fieldName 不能为空。'))
  }

  const rawCommand = command as unknown as Record<string, unknown>
  if (typeof rawCommand.format !== 'undefined') {
    return actionFailed(
      createDomainError('FieldFormatImmutable', '字段 format 不可编辑。', {
        code: 'FIELD_FORMAT_IMMUTABLE',
        suggestion: '请仅更新 title/description/validationConfig。',
      }),
    )
  }

  const state = getOrCreateState(command)
  const exists =
    Boolean(state.fieldFormats[command.fieldName]) ||
    state.enumSources.some((source) => source.fieldName === command.fieldName)

  if (!exists) {
    return actionFailed(fieldInputError(`字段 ${command.fieldName} 不存在。`))
  }

  state.fieldMeta[command.fieldName] = {
    title: command.title,
    description: command.description,
    validationConfig: command.validationConfig,
  }

  return actionSucceeded()
}
