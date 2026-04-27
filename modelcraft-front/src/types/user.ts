export type OrganizationStatus = 'ACTIVE' | 'SUSPENDED' | 'DELETED'
export type MembershipStatus = 'ACTIVE' | 'SUSPENDED' | 'INVITED'

export interface Organization {
  id: string
  name: string
  displayName?: string
  ownerID: string
  status: OrganizationStatus
  createdAt: string
  updatedAt: string
}

export interface Role {
  id: string
  name: string
  description?: string
  permissions: string[]
  isSystem: boolean
  createdAt: string
  updatedAt: string
}

export interface OrganizationMember {
  id: string
  userID: string
  userName: string
  orgID: string
  role: Role
  status: MembershipStatus
  joinedAt?: string
  createdAt: string
}

export interface CurrentUser {
  id: string
  externalID: string
  email: string
  name: string
  organization?: Organization
  role?: Role
  permissions: string[]
}

export interface CreateRoleInput {
  name: string
  description?: string
  permissions: string[]
}

export interface UpdateOrganizationInput {
  displayName?: string
}

export interface PermissionRole {
  id: number
  name: string
  description?: string
  isSystem: boolean
  orgName: string
  createdAt: string
  updatedAt: string
}

export interface PermissionDef {
  obj: string
  act: string
}

export interface UserRoleAssignment {
  id: number
  userId: string
  roleId: number
  orgName: string
  createdAt: string
}

export interface CreateCustomRoleInput {
  name: string
  description?: string
  orgName: string
}

export interface UpdateRoleInput {
  name?: string
  description?: string
}
