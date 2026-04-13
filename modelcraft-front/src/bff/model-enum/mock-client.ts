import type {
  EnumRelationOption,
  EnumSourceOption,
  ModelEnumDomainError,
  ModelEnumErrorCode,
  ModelEnumErrorType,
} from '@/types'
import type {
  CreateEnumFieldCommand,
  CreateEnumLabelFieldCommand,
  CreateFieldEnumRelationCommand,
  FieldEnumRelationListResult,
  ModelEnumActionResult,
  ModelEnumContextQuery,
  ModelEnumContextResult,
  UpdateFieldMetaCommand,
} from './types'

interface FieldMetaSnapshot {
  title?: string
  description?: string
  validationConfig?: UpdateFieldMetaCommand['validationConfig']
}

interface MockModelEnumState {
  enumSources: EnumSourceOption[]
  relations: EnumRelationOption[]
  fieldFormats: Record<string, 'ENUM' | 'ENUM_LABEL' | 'OTHER'>
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
  return createDomainError('InvalidModelInput', message, {
    suggestion: '请确认组织、项目和模型参数均已提供。',
  })
}

function fieldInputError(message: string): ModelEnumDomainError {
  return createDomainError('InvalidFieldInput', message, {
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

function cloneRelations(relations: EnumRelationOption[]): EnumRelationOption[] {
  return relations.map((relation) => ({ ...relation }))
}

function createMockRelationId(sourceFieldName: string): string {
  const safeSource = sourceFieldName.replace(/\s+/g, '-').toLowerCase()
  const randomSuffix = Math.random().toString(36).slice(2, 8)
  return `rel-${safeSource}-${randomSuffix}`
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
    relations: [],
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

export async function mockQueryModelEnumContext(
  query: ModelEnumContextQuery,
): Promise<ModelEnumContextResult> {
  const contextError = validateContext(query)
  if (contextError) {
    return {
      enumSources: [],
      relations: [],
      error: contextError,
    }
  }

  const state = getOrCreateState(query)

  return {
    enumSources: cloneSources(state.enumSources),
    relations: cloneRelations(state.relations),
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

export async function mockCreateEnumLabelField(
  command: CreateEnumLabelFieldCommand,
): Promise<ModelEnumActionResult> {
  const contextError = validateContext(command)
  if (contextError) {
    return actionFailed(contextError)
  }

  if (isBlank(command.name) || isBlank(command.title) || isBlank(command.sourceFieldName)) {
    return actionFailed(fieldInputError('ENUM_LABEL 字段参数不完整。'))
  }

  if (isBlank(command.enumRelationId)) {
    return actionFailed(fieldInputError('ENUM_LABEL 字段创建必须传入 enumRelationId。'))
  }

  const state = getOrCreateState(command)
  const source = state.enumSources.find((item) => item.fieldName === command.sourceFieldName)

  if (!source) {
    return actionFailed(fieldInputError(`未找到 source 字段 ${command.sourceFieldName}。`))
  }

  if (source.occupied) {
    return actionFailed(
      createDomainError('FieldEnumSourceConflict', '该 source 已绑定 ENUM_LABEL。', {
        code: 'FIELD_ENUM_SOURCE_CONFLICT',
        suggestion: '请更换 source 字段或解绑后重试。',
      }),
    )
  }

  const relation = state.relations.find((item) => item.id === command.enumRelationId)
  if (!relation) {
    return actionFailed(fieldInputError(`未找到 enumRelationId=${command.enumRelationId} 对应关系。`))
  }

  if (relation.sourceFieldName !== command.sourceFieldName) {
    return actionFailed(fieldInputError('enumRelationId 与 sourceFieldName 不匹配。'))
  }

  source.occupied = true
  state.fieldFormats[command.name] = 'ENUM_LABEL'
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
    state.enumSources.some((source) => source.fieldName === command.fieldName) ||
    state.relations.some((relation) => relation.labelFieldName === command.fieldName)

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

export async function mockListFieldEnumRelations(
  query: ModelEnumContextQuery,
): Promise<FieldEnumRelationListResult> {
  const contextError = validateContext(query)
  if (contextError) {
    return {
      relations: [],
      error: contextError,
    }
  }

  const state = getOrCreateState(query)

  return {
    relations: cloneRelations(state.relations),
    error: null,
  }
}

export async function mockCreateFieldEnumRelation(
  command: CreateFieldEnumRelationCommand,
): Promise<ModelEnumActionResult> {
  const contextError = validateContext(command)
  if (contextError) {
    return actionFailed(contextError)
  }

  if (isBlank(command.sourceFieldName) || isBlank(command.enumName) || isBlank(command.labelFieldName)) {
    return actionFailed(fieldInputError('创建 enum relation 参数不完整。'))
  }

  const state = getOrCreateState(command)
  const source = state.enumSources.find((item) => item.fieldName === command.sourceFieldName)

  if (!source) {
    return actionFailed(fieldInputError(`未找到 source 字段 ${command.sourceFieldName}。`))
  }

  if (source.enumName !== command.enumName) {
    return actionFailed(fieldInputError('enumName 与 sourceFieldName 的枚举定义不一致。'))
  }

  const sourceHasRelation = state.relations.some(
    (relation) => relation.sourceFieldName === command.sourceFieldName,
  )

  if (sourceHasRelation || source.occupied) {
    return actionFailed(
      createDomainError('FieldEnumSourceConflict', 'source 字段已存在关联关系。', {
        code: 'FIELD_ENUM_SOURCE_CONFLICT',
        suggestion: '每个 source 字段只能关联一个 ENUM_LABEL。',
      }),
    )
  }

  state.relations.push({
    id: createMockRelationId(command.sourceFieldName),
    sourceFieldName: command.sourceFieldName,
    enumName: command.enumName,
    labelFieldName: command.labelFieldName,
  })

  return actionSucceeded()
}
