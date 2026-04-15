'use client'

import React from 'react'
import {
  createEnumField,
  createEnumLabelField,
  createFieldEnumRelation,
  listFieldEnumRelations,
  queryModelEnumContext,
  updateFieldMeta,
} from '@bff/model-enum/public'
import { useProjectScopedClient } from '@bff/apollo/public'
import type { RawGraphQLErrorLike } from '@/shared/errors/model-enum-error-mapper'
import { mapModelEnumError } from '@/shared/errors/model-enum-error-mapper'
import type {
  CreateEnumFieldFormValues,
  CreateEnumLabelFieldFormValues,
  EnumRelationOption,
  EnumSourceOption,
  ModelEnumDomainError,
  UpdateFieldMetaFormValues,
} from '@/types'

interface UseModelEnumContextState {
  sourceOptions: EnumSourceOption[]
  relationOptions: EnumRelationOption[]
  error: ModelEnumDomainError | null
}

function isModelEnumDomainError(value: unknown): value is ModelEnumDomainError {
  if (!value || typeof value !== 'object') {
    return false
  }

  const candidate = value as Record<string, unknown>
  return typeof candidate.type === 'string' && typeof candidate.message === 'string'
}

function createUnknownError(message: string): ModelEnumDomainError {
  return {
    type: 'Unknown',
    message,
  }
}

function normalizeUnknownError(error: unknown, fallbackMessage: string): ModelEnumDomainError {
  if (error instanceof Error && error.message.trim().length > 0) {
    return createUnknownError(error.message)
  }

  return createUnknownError(fallbackMessage)
}

function normalizeDomainError(error: unknown): ModelEnumDomainError | null {
  if (!error) {
    return null
  }

  if (isModelEnumDomainError(error)) {
    return error
  }

  if (typeof error === 'string' && error.trim().length > 0) {
    return createUnknownError(error)
  }

  if (typeof error === 'object') {
    return mapModelEnumError(error as RawGraphQLErrorLike)
  }

  return createUnknownError('发生未知错误，请稍后重试。')
}

async function runModelEnumAction(
  action: () => Promise<{ success: boolean; error: unknown }>,
  fallbackMessage: string,
): Promise<ModelEnumDomainError | null> {
  const result = await action()
  const mappedError = normalizeDomainError(result.error)

  if (!result.success) {
    return mappedError ?? createUnknownError(fallbackMessage)
  }

  return mappedError
}

export interface UseModelEnumContextParams {
  orgName: string
  projectSlug: string
  modelId: string
}

export interface UseModelEnumContextReturn {
  sourceOptions: EnumSourceOption[]
  relationOptions: EnumRelationOption[]
  loading: boolean
  error: ModelEnumDomainError | null
  refetch: () => Promise<void>
}

export function useModelEnumContext(params: UseModelEnumContextParams): UseModelEnumContextReturn {
  const { orgName, projectSlug, modelId } = params
  const projectClient = useProjectScopedClient(projectSlug)
  const [sourceOptions, setSourceOptions] = React.useState<EnumSourceOption[]>([])
  const [relationOptions, setRelationOptions] = React.useState<EnumRelationOption[]>([])
  const [loading, setLoading] = React.useState(false)
  const [error, setError] = React.useState<ModelEnumDomainError | null>(null)

  const runContextRequest = React.useCallback(async (): Promise<UseModelEnumContextState> => {
    const [contextResult, relationResult] = await Promise.all([
      queryModelEnumContext({ orgName, projectSlug, modelId }, projectClient),
      listFieldEnumRelations({ orgName, projectSlug, modelId }),
    ])

    const contextError = normalizeDomainError(contextResult.error)
    const relationError = normalizeDomainError(relationResult.error)

    return {
      sourceOptions: contextResult.enumSources,
      relationOptions: relationError ? contextResult.relations : relationResult.relations,
      error: contextError ?? relationError,
    }
  }, [modelId, orgName, projectSlug, projectClient])

  const refetch = React.useCallback(async () => {
    setLoading(true)
    setError(null)

    try {
      const nextState = await runContextRequest()
      setSourceOptions(nextState.sourceOptions)
      setRelationOptions(nextState.relationOptions)
      setError(nextState.error)
    } catch (refetchError) {
      setError(normalizeUnknownError(refetchError, '加载字段枚举上下文失败，请稍后重试。'))
    } finally {
      setLoading(false)
    }
  }, [runContextRequest])

  React.useEffect(() => {
    let alive = true

    const bootstrap = async () => {
      setLoading(true)
      setError(null)

      try {
        const nextState = await runContextRequest()

        if (!alive) {
          return
        }

        setSourceOptions(nextState.sourceOptions)
        setRelationOptions(nextState.relationOptions)
        setError(nextState.error)
      } catch (bootstrapError) {
        if (!alive) {
          return
        }

        setError(normalizeUnknownError(bootstrapError, '加载字段枚举上下文失败，请稍后重试。'))
      } finally {
        if (alive) {
          setLoading(false)
        }
      }
    }

    void bootstrap()

    return () => {
      alive = false
    }
  }, [runContextRequest])

  return {
    sourceOptions,
    relationOptions,
    loading,
    error,
    refetch,
  }
}

export interface UseCreateEnumFieldReturn {
  mutate: (values: CreateEnumFieldFormValues) => Promise<void>
  loading: boolean
  error: ModelEnumDomainError | null
}

export function useCreateEnumField(params: UseModelEnumContextParams): UseCreateEnumFieldReturn {
  const { orgName, projectSlug, modelId } = params
  const [loading, setLoading] = React.useState(false)
  const [error, setError] = React.useState<ModelEnumDomainError | null>(null)

  const mutate = React.useCallback(
    async (values: CreateEnumFieldFormValues) => {
      setLoading(true)
      setError(null)

      try {
        const actionError = await runModelEnumAction(
          () =>
            createEnumField({
              orgName,
              projectSlug,
              modelId,
              name: values.name,
              title: values.title,
              description: values.description,
              relateEnumName: values.relateEnumName,
            }),
          '创建 ENUM 字段失败。',
        )

        setError(actionError)
      } catch (mutationError) {
        setError(normalizeUnknownError(mutationError, '创建 ENUM 字段失败，请稍后重试。'))
      } finally {
        setLoading(false)
      }
    },
    [modelId, orgName, projectSlug],
  )

  return {
    mutate,
    loading,
    error,
  }
}

export interface UseCreateEnumLabelFieldReturn {
  mutate: (values: CreateEnumLabelFieldFormValues) => Promise<void>
  loading: boolean
  error: ModelEnumDomainError | null
}

export function useCreateEnumLabelField(params: UseModelEnumContextParams): UseCreateEnumLabelFieldReturn {
  const { orgName, projectSlug, modelId } = params
  const [loading, setLoading] = React.useState(false)
  const [error, setError] = React.useState<ModelEnumDomainError | null>(null)

  const mutate = React.useCallback(
    async (values: CreateEnumLabelFieldFormValues) => {
      setLoading(true)
      setError(null)

      try {
        const enumRelationID = values.enumRelationId?.trim()
        if (!enumRelationID) {
          setError({
            type: 'InvalidInput',
            message: '创建 ENUM_LABEL 前必须先创建 relation。',
          })
          return
        }

        const actionError = await runModelEnumAction(
          () =>
            createEnumLabelField({
              orgName,
              projectSlug,
              modelId,
              name: values.name,
              title: values.title,
              description: values.description,
              sourceFieldName: values.sourceFieldName,
              enumRelationId: enumRelationID,
            }),
          '创建 ENUM_LABEL 字段失败。',
        )

        setError(actionError)
      } catch (mutationError) {
        setError(normalizeUnknownError(mutationError, '创建 ENUM_LABEL 字段失败，请稍后重试。'))
      } finally {
        setLoading(false)
      }
    },
    [modelId, orgName, projectSlug],
  )

  return {
    mutate,
    loading,
    error,
  }
}

export interface UseUpdateFieldMetaReturn {
  mutate: (fieldName: string, values: UpdateFieldMetaFormValues) => Promise<void>
  loading: boolean
  error: ModelEnumDomainError | null
}

export function useUpdateFieldMeta(params: UseModelEnumContextParams): UseUpdateFieldMetaReturn {
  const { orgName, projectSlug, modelId } = params
  const [loading, setLoading] = React.useState(false)
  const [error, setError] = React.useState<ModelEnumDomainError | null>(null)

  const mutate = React.useCallback(
    async (fieldName: string, values: UpdateFieldMetaFormValues) => {
      setLoading(true)
      setError(null)

      try {
        const actionError = await runModelEnumAction(
          () =>
            updateFieldMeta({
              orgName,
              projectSlug,
              modelId,
              fieldName,
              title: values.title,
              description: values.description,
              validationConfig: values.validationConfig,
            }),
          '更新字段元信息失败。',
        )

        setError(actionError)
      } catch (mutationError) {
        setError(normalizeUnknownError(mutationError, '更新字段元信息失败，请稍后重试。'))
      } finally {
        setLoading(false)
      }
    },
    [modelId, orgName, projectSlug],
  )

  return {
    mutate,
    loading,
    error,
  }
}

export interface UseCreateFieldEnumRelationReturn {
  mutate: (sourceFieldName: string, enumName: string, labelFieldName: string) => Promise<void>
  loading: boolean
  error: ModelEnumDomainError | null
}

export function useCreateFieldEnumRelation(
  params: UseModelEnumContextParams,
): UseCreateFieldEnumRelationReturn {
  const { orgName, projectSlug, modelId } = params
  const [loading, setLoading] = React.useState(false)
  const [error, setError] = React.useState<ModelEnumDomainError | null>(null)

  const mutate = React.useCallback(
    async (sourceFieldName: string, enumName: string, labelFieldName: string) => {
      setLoading(true)
      setError(null)

      try {
        const actionError = await runModelEnumAction(
          () =>
            createFieldEnumRelation({
              orgName,
              projectSlug,
              modelId,
              sourceFieldName,
              enumName,
              labelFieldName,
            }),
          '创建字段枚举关联失败。',
        )

        setError(actionError)
      } catch (mutationError) {
        setError(normalizeUnknownError(mutationError, '创建字段枚举关联失败，请稍后重试。'))
      } finally {
        setLoading(false)
      }
    },
    [modelId, orgName, projectSlug],
  )

  return {
    mutate,
    loading,
    error,
  }
}
