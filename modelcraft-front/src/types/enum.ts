export interface EnumOption {
  id: string
  code: string
  label: string
  order: number
  description?: string
}

export interface EnumDefinition {
  id: string
  projectSlug: string
  name: string
  displayName: string
  description?: string
  options: EnumOption[]
  isMultiSelect: boolean
  createdAt: string
  updatedAt: string
}

export interface EnumOptionInput {
  code: string
  label: string
  order: number
  description?: string
}

export interface CreateEnumInput {
  projectSlug: string
  name: string
  displayName: string
  description?: string
  options: EnumOptionInput[]
}

export interface UpdateEnumInput {
  displayName?: string
  description?: string
  options?: EnumOptionInput[]
}
