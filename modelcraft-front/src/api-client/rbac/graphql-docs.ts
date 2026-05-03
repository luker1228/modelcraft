import { gql } from '@apollo/client'

// ── Queries ────────────────────────────────────────────────────────────────────

export const GET_END_USER_PERMISSIONS = gql`
  query GetEndUserPermissions {
    endUserPermissions(input: {}) {
      edges {
        node {
          id
          modelId
          action
          rowScope
          preset
          displayName
          description
          columnPolicy {
            defaultMode
            rules {
              fieldName
              mode
              maskPattern
            }
          }
          createdAt
          updatedAt
        }
      }
      pageInfo {
        hasNextPage
        endCursor
      }
      totalCount
    }
  }
`

export const GET_END_USER_BUNDLES = gql`
  query GetEndUserBundles {
    endUserPermissionBundles(input: {}) {
      edges {
        node {
          id
          slug
          name
          description
          createdAt
          updatedAt
          permissions {
            sortOrder
            permission {
              id
              modelId
              action
              rowScope
              preset
              columnPolicy {
                defaultMode
                rules {
                  fieldName
                  mode
                  maskPattern
                }
              }
            }
          }
        }
      }
      pageInfo {
        hasNextPage
        endCursor
      }
      totalCount
    }
  }
`

export const GET_END_USER_BUNDLE = gql`
  query GetEndUserBundle($id: ID!) {
    endUserPermissionBundle(id: $id) {
      id
      slug
      name
      description
      createdAt
      updatedAt
      currentVersion
      snapshots {
        version
        createdAt
        createdBy
        restoredFrom
        items {
          modelId
          grantType
          preset
          customPermissionId
          sortOrder
        }
        permissions {
          sortOrder
          permissionId
        }
      }
      dataPermissionItems {
        id
        bundleId
        modelId
        grantType
        preset
        customPermissionId
        customPermission {
          id
          modelId
          action
          rowScope
          displayName
          description
          columnPolicy {
            defaultMode
            rules {
              fieldName
              mode
              maskPattern
            }
          }
        }
        modelName
        databaseName
        modelTitle
        sortOrder
        createdAt
        updatedAt
      }
      permissions {
        sortOrder
        permission {
          id
          modelId
          action
          rowScope
          preset
          displayName
          description
          columnPolicy {
            defaultMode
            rules {
              fieldName
              mode
              maskPattern
            }
          }
        }
      }
    }
  }
`

export const GET_END_USER_ROLES = gql`
  query GetEndUserRoles {
    endUserRoles(input: { includeImplicit: true }) {
      edges {
        node {
          id
          name
          description
          isImplicit
          createdAt
          updatedAt
          permissionBundles {
            bundle {
              id
              name
              description
            }
          }
        }
      }
      pageInfo {
        hasNextPage
        endCursor
      }
      totalCount
    }
  }
`

export const GET_END_USER_ROLE = gql`
  query GetEndUserRole($id: ID!) {
    endUserRole(id: $id) {
      id
      name
      description
      isImplicit
      createdAt
      updatedAt
      permissionBundles {
        bundle {
          id
          name
          description
          permissions {
            sortOrder
            permission {
              id
              modelId
              action
              rowScope
              preset
              displayName
            }
          }
        }
      }
    }
  }
`

export const GET_END_USER_EFFECTIVE_PERMISSIONS = gql`
  query GetEndUserEffectivePermissions($endUserId: ID!, $modelId: ID!) {
    effectivePermissions(input: { endUserId: $endUserId, modelId: $modelId }) {
      effectivePermissions {
        endUserId
        modelId
        grants {
          action
          rowScope
          columnPolicy {
            defaultMode
            rules {
              fieldName
              mode
              maskPattern
            }
          }
        }
      }
      error {
        __typename
        ... on ResourceNotFound {
          message
          resourceType
        }
      }
    }
  }
`

export const LIST_PROJECT_END_USER_ROLE_USERS = gql`
  query ListProjectEndUserRoleUsers($input: ListProjectEndUserRoleUsersInput) {
    listProjectEndUserRoleUsers(input: $input) {
      connection {
        nodes {
          endUser {
            id
            username
            isForbidden
          }
          role {
            id
            name
            description
          }
          assignedAt
        }
        pageInfo {
          hasNextPage
          endCursor
        }
        totalCount
      }
      error {
        __typename
        ... on InvalidInput {
          message
        }
        ... on ResourceNotFound {
          message
          resourceType
        }
      }
    }
  }
`

export const GET_END_USER_ROLE_ASSIGNMENTS = gql`
  query GetEndUserRoleAssignments($endUserId: ID!) {
    endUserRoleAssignments(endUserId: $endUserId) {
      endUserId
      assignedAt
      role {
        id
        name
        description
        isImplicit
      }
    }
  }
`

export const GET_END_USER_BUNDLE_ASSIGNMENTS = gql`
  query GetEndUserBundleAssignments($endUserId: ID!) {
    endUserBundleAssignments(endUserId: $endUserId) {
      endUserId
      assignedAt
      bundle {
        id
        name
        description
        permissions {
          sortOrder
          permission {
            id
            modelId
            action
            rowScope
          }
        }
      }
    }
  }
`

// ── Mutations ──────────────────────────────────────────────────────────────────

export const CREATE_END_USER_PERMISSION = gql`
  mutation CreateEndUserPermission($input: CreateEndUserPermissionInput!) {
    createEndUserPermission(input: $input) {
      permission {
        id
        modelId
        action
        rowScope
        displayName
        description
        columnPolicy {
          defaultMode
          rules { fieldName mode maskPattern }
        }
        createdAt
        updatedAt
      }
      error {
        __typename
        ... on InvalidInput {
          message
        }
        ... on ResourceNotFound {
          message
          resourceType
        }
        ... on RowScopeFieldMissing {
          message
          missingField
          requiredByRowScope
        }
      }
    }
  }
`

export const DELETE_END_USER_PERMISSION = gql`
  mutation DeleteEndUserPermission($id: ID!) {
    deleteEndUserPermission(id: $id) {
      success
      error {
        __typename
        ... on ResourceNotFound {
          message
          resourceType
        }
        ... on EndUserPermissionInUse {
          message
        }
      }
    }
  }
`

export const CREATE_END_USER_BUNDLE = gql`
  mutation CreateEndUserBundle($input: CreateEndUserPermissionBundleInput!) {
    createEndUserPermissionBundle(input: $input) {
      bundle {
        id
        slug
        name
        description
        permissions {
          sortOrder
          permission {
            id
            modelId
            action
            rowScope
          }
        }
        createdAt
        updatedAt
      }
      error {
        __typename
        ... on InvalidInput {
          message
        }
        ... on EndUserPermissionBundleAlreadyExists {
          message
        }
        ... on ResourceNotFound {
          message
          resourceType
        }
      }
    }
  }
`

export const UPDATE_END_USER_BUNDLE = gql`
  mutation UpdateEndUserBundle($id: ID!, $input: UpdateEndUserPermissionBundleInput!) {
    updateEndUserPermissionBundle(id: $id, input: $input) {
      bundle {
        id
        slug
        name
        description
        updatedAt
      }
      error {
        __typename
        ... on ResourceNotFound {
          message
          resourceType
        }
        ... on EndUserPermissionBundleAlreadyExists {
          message
        }
        ... on InvalidInput {
          message
        }
      }
    }
  }
`

export const DELETE_END_USER_BUNDLE = gql`
  mutation DeleteEndUserBundle($id: ID!) {
    deleteEndUserPermissionBundle(id: $id) {
      success
      error {
        __typename
        ... on ResourceNotFound {
          message
          resourceType
        }
        ... on EndUserPermissionBundleInUse {
          message
        }
      }
    }
  }
`

export const ADD_PERMISSION_TO_BUNDLE = gql`
  mutation AddPermissionToBundle($input: AddEndUserPermissionToBundleInput!) {
    addEndUserPermissionToBundle(input: $input) {
      bundle {
        id
        slug
        name
        permissions {
          sortOrder
          permission {
            id
            modelId
            action
            rowScope
          }
        }
      }
      error {
        __typename
        ... on ResourceNotFound {
          message
          resourceType
        }
        ... on InvalidInput {
          message
        }
      }
    }
  }
`

export const REMOVE_PERMISSION_FROM_BUNDLE = gql`
  mutation RemovePermissionFromBundle($input: RemoveEndUserPermissionFromBundleInput!) {
    removeEndUserPermissionFromBundle(input: $input) {
      bundle {
        id
        slug
        name
        permissions {
          sortOrder
          permission {
            id
            modelId
            action
            rowScope
          }
        }
      }
      error {
        __typename
        ... on ResourceNotFound {
          message
          resourceType
        }
      }
    }
  }
`

export const CREATE_END_USER_ROLE = gql`
  mutation CreateEndUserRole($input: CreateEndUserRoleInput!) {
    createEndUserRole(input: $input) {
      role {
        id
        name
        description
        isImplicit
        permissionBundles {
          bundle {
            id
            name
          }
        }
        createdAt
        updatedAt
      }
      error {
        __typename
        ... on InvalidInput {
          message
        }
        ... on EndUserRoleAlreadyExists {
          message
        }
        ... on ResourceNotFound {
          message
          resourceType
        }
      }
    }
  }
`

export const DELETE_END_USER_ROLE = gql`
  mutation DeleteEndUserRole($id: ID!) {
    deleteEndUserRole(id: $id) {
      success
      error {
        __typename
        ... on ResourceNotFound {
          message
          resourceType
        }
        ... on EndUserImplicitRoleCannotBeModified {
          message
        }
      }
    }
  }
`

export const UPDATE_END_USER_ROLE = gql`
  mutation UpdateEndUserRole($id: ID!, $input: UpdateEndUserRoleInput!) {
    updateEndUserRole(id: $id, input: $input) {
      role {
        id
        name
        description
        isImplicit
        updatedAt
      }
      error {
        __typename
        ... on ResourceNotFound {
          message
          resourceType
        }
        ... on EndUserImplicitRoleCannotBeModified {
          message
        }
      }
    }
  }
`

export const ASSIGN_BUNDLE_TO_ROLE = gql`
  mutation AssignBundleToRole($input: AssignBundleToEndUserRoleInput!) {
    assignBundleToEndUserRole(input: $input) {
      role {
        id
        name
        permissionBundles {
          bundle {
            id
            name
            description
          }
        }
      }
      error {
        __typename
        ... on ResourceNotFound {
          message
          resourceType
        }
      }
    }
  }
`

export const REVOKE_BUNDLE_FROM_ROLE = gql`
  mutation RevokeBundleFromRole($input: RevokeBundleFromEndUserRoleInput!) {
    revokeBundleFromEndUserRole(input: $input) {
      role {
        id
        name
        permissionBundles {
          bundle {
            id
            name
            description
          }
        }
      }
      error {
        __typename
        ... on ResourceNotFound {
          message
          resourceType
        }
      }
    }
  }
`

export const ASSIGN_END_USER_ROLE_TO_USER = gql`
  mutation AssignEndUserRole($input: AssignEndUserRoleInput!) {
    assignEndUserRole(input: $input) {
      endUserId
      role {
        id
        name
      }
      error {
        __typename
        ... on ResourceNotFound {
          message
          resourceType
        }
        ... on EndUserCannotAssignImplicitRole {
          message
        }
        ... on UserRoleAlreadyAssigned {
          message
        }
      }
    }
  }
`

export const REVOKE_END_USER_ROLE_FROM_USER = gql`
  mutation RevokeEndUserRole($input: RevokeEndUserRoleInput!) {
    revokeEndUserRole(input: $input) {
      success
      error {
        __typename
        ... on ResourceNotFound {
          message
          resourceType
        }
      }
    }
  }
`

export const ASSIGN_BUNDLE_TO_END_USER = gql`
  mutation AssignBundleToEndUser($input: AssignBundleToEndUserInput!) {
    assignBundleToEndUser(input: $input) {
      endUserId
      bundle {
        id
        name
      }
      error {
        __typename
        ... on ResourceNotFound {
          message
          resourceType
        }
        ... on UserBundleAlreadyAssigned {
          message
        }
      }
    }
  }
`

export const REVOKE_BUNDLE_FROM_END_USER = gql`
  mutation RevokeBundleFromEndUser($input: RevokeBundleFromEndUserInput!) {
    revokeBundleFromEndUser(input: $input) {
      success
      error {
        __typename
        ... on ResourceNotFound {
          message
          resourceType
        }
      }
    }
  }
`

export const APPLY_END_USER_PRESET_POLICY = gql`
  mutation ApplyEndUserPresetPolicy($input: ApplyEndUserPresetPolicyInput!) {
    applyEndUserPresetPolicy(input: $input) {
      permissions {
        id
        modelId
        preset
        displayName
        description
        columnPolicy {
          defaultMode
          rules {
            fieldName
            mode
            maskPattern
          }
        }
        createdAt
        updatedAt
      }
      error {
        __typename
        ... on ResourceNotFound {
          message
          resourceType
        }
        ... on PresetRequiresOwnerField {
          message
        }
      }
    }
  }
`

export const RESTORE_END_USER_BUNDLE = gql`
  mutation RestoreEndUserBundle($input: RestoreEndUserPermissionBundleInput!) {
    restoreEndUserPermissionBundle(input: $input) {
      bundle {
        id
        slug
        currentVersion
        snapshots {
          version
          createdAt
          createdBy
          restoredFrom
          items {
            modelId
            grantType
            preset
            customPermissionId
            sortOrder
          }
          permissions {
            sortOrder
            permissionId
          }
        }
        dataPermissionItems {
          id
          bundleId
          modelId
          grantType
          preset
          customPermissionId
          customPermission {
            id
            modelId
            action
            rowScope
            displayName
            description
          }
          sortOrder
          createdAt
          updatedAt
        }
        permissions {
          sortOrder
          permission {
            id
            modelId
            action
            rowScope
            displayName
            description
            columnPolicy {
              defaultMode
              rules {
                fieldName
                mode
                maskPattern
              }
            }
          }
        }
      }
      newVersion
      error {
        __typename
        ... on ResourceNotFound {
          message
          resourceType
        }
      }
    }
  }
`

export const BIND_PRESET_ITEM_TO_BUNDLE = gql`
  mutation BindPresetItemToBundle($input: BindPresetItemToBundleInput!) {
    bindPresetItemToBundle(input: $input) {
      bundle {
        id
        slug
        dataPermissionItems {
          id
          bundleId
          modelId
          grantType
          preset
          sortOrder
          createdAt
          updatedAt
        }
        currentVersion
      }
      error {
        __typename
        ... on ResourceNotFound {
          message
          resourceType
        }
        ... on PresetRequiresOwnerField {
          message
          preset
        }
        ... on InvalidInput {
          message
        }
      }
    }
  }
`

export const BIND_CUSTOM_ITEM_TO_BUNDLE = gql`
  mutation BindCustomItemToBundle($input: BindCustomItemToBundleInput!) {
    bindCustomItemToBundle(input: $input) {
      bundle {
        id
        slug
        dataPermissionItems {
          id
          bundleId
          modelId
          grantType
          customPermissionId
          customPermission {
            id
            modelId
            action
            rowScope
            displayName
            description
          }
          sortOrder
          createdAt
          updatedAt
        }
        currentVersion
      }
      error {
        __typename
        ... on ResourceNotFound {
          message
          resourceType
        }
        ... on InvalidInput {
          message
        }
      }
    }
  }
`

export const REMOVE_DATA_PERMISSION_ITEM_FROM_BUNDLE = gql`
  mutation RemoveDataPermissionItemFromBundle($input: RemoveDataPermissionItemFromBundleInput!) {
    removeDataPermissionItemFromBundle(input: $input) {
      bundle {
        id
        slug
        dataPermissionItems {
          id
          bundleId
          modelId
          grantType
          preset
          customPermissionId
          sortOrder
        }
        currentVersion
      }
      error {
        __typename
        ... on ResourceNotFound {
          message
          resourceType
        }
      }
    }
  }
`

export const GET_VIRTUAL_PRESETS_BY_MODEL = gql`
  query GetVirtualPresetsByModel($modelId: ID!) {
    virtualPresetsByModel(modelId: $modelId)
  }
`
