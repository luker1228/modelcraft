export type FKDirection = 'NORMAL' | 'REVERSE'

export interface LogicalForeignKey {
  id: string
  pairId: string
  direction: FKDirection
  modelId: string
  modelName?: string
  refModelId: string
  refModelName?: string
  sourceFields: string[]
  targetFields: string[]
}

export interface CreateLogicalForeignKeyInput {
  modelId: string
  refModelId: string
  sourceFields: string[]
  targetFields: string[]
}

export interface FKColumnsNotFoundError {
  message: string
}

export interface FKFieldCountMismatchError {
  message: string
}

export interface FKNotFoundError {
  message: string
}

export interface FKPairHasRelateFieldsError {
  message: string
}

export type CreateLogicalForeignKeyResult =
  | (LogicalForeignKey & { __typename: 'LogicalForeignKey' })
  | (FKColumnsNotFoundError & { __typename: 'FKColumnsNotFoundError' })
  | (FKFieldCountMismatchError & { __typename: 'FKFieldCountMismatchError' })

export type DeleteLogicalForeignKeyResult =
  | ({ pairId: string } & { __typename: 'DeleteLogicalForeignKeySuccess' })
  | (FKNotFoundError & { __typename: 'FKNotFoundError' })
  | (FKPairHasRelateFieldsError & { __typename: 'FKPairHasRelateFieldsError' })
