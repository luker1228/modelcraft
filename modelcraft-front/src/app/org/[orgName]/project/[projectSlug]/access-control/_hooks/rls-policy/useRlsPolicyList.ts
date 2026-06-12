/* eslint-disable @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-member-access */

import { useQuery } from '@apollo/client'
import { useProjectScopedClient } from '@api-client/apollo/develop-client'
import { GET_RLS_POLICIES } from '@/api-client/rls-policy'
import type { RlsPolicy } from '@/generated/graphql'

interface UseRlsPolicyListProps {
  projectSlug: string
  modelId: string | null
}

interface UseRlsPolicyListReturn {
  policies: RlsPolicy[]
  loading: boolean
  error: Error | undefined
}

export function useRlsPolicyList({ projectSlug, modelId }: UseRlsPolicyListProps): UseRlsPolicyListReturn {
  const client = useProjectScopedClient(projectSlug)

  const { data, loading, error } = useQuery(GET_RLS_POLICIES, {
    client,
    variables: { modelId },
    skip: !modelId,
  })

  const policies: RlsPolicy[] = data?.rlsPolicies ?? []

  return { policies, loading, error }
}
