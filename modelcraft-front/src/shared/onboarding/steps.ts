import type { OnboardingStepId } from './storage'

/** A sub-step that records completion (driven by mutation onCompleted) */
export interface OnboardingTrackedStep {
  kind: 'tracked'
  id: OnboardingStepId
  label: string
  type: 'action' | 'manual'
  optional?: boolean
  route: (params: { orgName: string; projectSlug: string | null }) => string | null
}

/** A sub-step that is pure navigation — no completion tracking, always clickable */
export interface OnboardingNavStep {
  kind: 'nav'
  id: string
  label: string
  route: (params: { orgName: string; projectSlug: string | null }) => string
}

export type OnboardingSubStep = OnboardingTrackedStep | OnboardingNavStep

export interface OnboardingGroup {
  id: string
  label: string
  optional?: boolean
  steps: OnboardingSubStep[]
}

export const ONBOARDING_GROUPS: OnboardingGroup[] = [
  {
    id: 'setup_project',
    label: '创建项目',
    steps: [
      {
        kind: 'nav',
        id: 'goto_project',
        label: '前往项目列表',
        route: ({ orgName }) => `/org/${orgName}/workspace`,
      },
      {
        kind: 'tracked',
        id: 'create_project',
        label: '创建项目',
        type: 'action',
        route: ({ orgName }) => `/org/${orgName}/workspace`,
      },
    ],
  },
  {
    id: 'design_model',
    label: '创建模型',
    steps: [
      {
        kind: 'nav',
        id: 'goto_model_editor',
        label: '前往模型编辑',
        route: ({ orgName, projectSlug }) =>
          projectSlug
            ? `/org/${orgName}/project/${projectSlug}/model-editor`
            : `/org/${orgName}/workspace`,
      },
      {
        kind: 'tracked',
        id: 'select_database',
        label: '选择数据库',
        type: 'action',
        route: ({ orgName, projectSlug }) =>
          projectSlug
            ? `/org/${orgName}/project/${projectSlug}/model-editor`
            : `/org/${orgName}/workspace`,
      },
      {
        kind: 'tracked',
        id: 'create_model',
        label: '创建模型',
        type: 'action',
        route: ({ orgName, projectSlug }) =>
          projectSlug
            ? `/org/${orgName}/project/${projectSlug}/model-editor`
            : `/org/${orgName}/workspace`,
      },
      {
        kind: 'tracked',
        id: 'insert_column',
        label: '插入列',
        type: 'action',
        route: ({ orgName, projectSlug }) =>
          projectSlug
            ? `/org/${orgName}/project/${projectSlug}/model-editor`
            : `/org/${orgName}/workspace`,
      },
      {
        kind: 'tracked',
        id: 'insert_data',
        label: '添加第一条数据',
        type: 'action',
        route: ({ orgName, projectSlug }) =>
          projectSlug
            ? `/org/${orgName}/project/${projectSlug}/model-editor`
            : `/org/${orgName}/workspace`,
      },
    ],
  },
  {
    id: 'configure_permissions',
    label: '创建权限及角色',
    optional: true,
    steps: [
      {
        kind: 'tracked',
        id: 'create_permission',
        label: '创建权限点',
        type: 'action',
        optional: true,
        route: ({ orgName, projectSlug }) =>
          projectSlug
            ? `/org/${orgName}/project/${projectSlug}/roles?tab=permissions`
            : `/org/${orgName}/workspace`,
      },
      {
        kind: 'tracked',
        id: 'create_bundle',
        label: '创建权限包',
        type: 'action',
        optional: true,
        route: ({ orgName, projectSlug }) =>
          projectSlug
            ? `/org/${orgName}/project/${projectSlug}/roles?tab=bundles`
            : `/org/${orgName}/workspace`,
      },
      {
        kind: 'tracked',
        id: 'create_role',
        label: '创建角色',
        type: 'action',
        optional: true,
        route: ({ orgName, projectSlug }) =>
          projectSlug
            ? `/org/${orgName}/project/${projectSlug}/roles`
            : `/org/${orgName}/workspace`,
      },
    ],
  },
  {
    id: 'add_users',
    label: '创建用户并体验',
    steps: [
      {
        kind: 'tracked',
        id: 'add_end_user',
        label: '添加终端用户',
        type: 'action',
        route: ({ orgName }) => `/org/${orgName}/end-users`,
      },
      {
        kind: 'tracked',
        id: 'assign_role',
        label: '分配角色',
        type: 'action',
        route: ({ orgName }) => `/org/${orgName}/end-users`,
      },
      {
        kind: 'tracked',
        id: 'end_user_login',
        label: '终端用户体验',
        type: 'manual',
        route: () => null,
      },
    ],
  },
]

/** Flat list of tracked steps only — used for completion counting (excludes optional) */
export const ONBOARDING_STEPS: OnboardingTrackedStep[] = ONBOARDING_GROUPS
  .flatMap((g) => g.steps)
  .filter((s): s is OnboardingTrackedStep => s.kind === 'tracked' && !s.optional)

/** All tracked steps including optional — for full state tracking */
export const ALL_TRACKED_STEPS: OnboardingTrackedStep[] = ONBOARDING_GROUPS
  .flatMap((g) => g.steps)
  .filter((s): s is OnboardingTrackedStep => s.kind === 'tracked')
