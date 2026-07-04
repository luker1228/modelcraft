import type { OnboardingStepId } from './storage'

/** A sub-step that records completion — always type: 'manual', user confirms */
export interface OnboardingTrackedStep {
  kind: 'tracked'
  id: OnboardingStepId
  label: string
  type: 'manual'
  route: (params: { orgName: string; projectSlug: string | null }) => string | null
}

/** A sub-step that is pure navigation — no completion tracking */
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
        route: ({ orgName }) => `/org/${orgName}/dashboard`,
      },
      {
        kind: 'nav',
        id: 'nav_create_project',
        label: '新建项目',
        route: ({ orgName }) => `/org/${orgName}/dashboard`,
      },
      {
        kind: 'tracked',
        id: 'confirm_project',
        label: '确认已完成',
        type: 'manual',
        route: () => null,
      },
    ],
  },
  {
    id: 'design_model',
    label: '导入模型',
    steps: [
      {
        kind: 'nav',
        id: 'goto_model_editor',
        label: '进入项目',
        route: ({ orgName, projectSlug }) =>
          projectSlug
            ? `/org/${orgName}/project/${projectSlug}/model-editor`
            : `/org/${orgName}/dashboard`,
      },
      {
        kind: 'nav',
        id: 'select_database_nav',
        label: '选择数据库',
        route: ({ orgName, projectSlug }) =>
          projectSlug
            ? `/org/${orgName}/project/${projectSlug}/model-editor`
            : `/org/${orgName}/dashboard`,
      },
      {
        kind: 'nav',
        id: 'nav_import_model',
        label: '点击导入模型',
        route: ({ orgName, projectSlug }) =>
          projectSlug
            ? `/org/${orgName}/project/${projectSlug}/model-editor`
            : `/org/${orgName}/dashboard`,
      },
      {
        kind: 'tracked',
        id: 'confirm_model',
        label: '确认已完成',
        type: 'manual',
        route: () => null,
      },
    ],
  },
  {
    id: 'add_users',
    label: '创建终端用户',
    steps: [
      {
        kind: 'nav',
        id: 'goto_end_users',
        label: '前往终端用户',
        route: ({ orgName }) => `/org/${orgName}/end-users`,
      },
      {
        kind: 'nav',
        id: 'nav_add_end_user',
        label: '新建终端用户',
        route: ({ orgName }) => `/org/${orgName}/end-users`,
      },
      {
        kind: 'tracked',
        id: 'confirm_end_user',
        label: '确认已完成',
        type: 'manual',
        route: () => null,
      },
    ],
  },
  {
    id: 'assign_permissions',
    label: '分配权限',
    steps: [
      {
        kind: 'nav',
        id: 'goto_end_user_access',
        label: '进入项目，前往终端用户授权',
        route: ({ orgName, projectSlug }) =>
          projectSlug
            ? `/org/${orgName}/project/${projectSlug}/end-user-access`
            : `/org/${orgName}/dashboard`,
      },
      {
        kind: 'nav',
        id: 'nav_assign_role',
        label: '为用户分配角色',
        route: ({ orgName, projectSlug }) =>
          projectSlug
            ? `/org/${orgName}/project/${projectSlug}/end-user-access`
            : `/org/${orgName}/dashboard`,
      },
      {
        kind: 'tracked',
        id: 'confirm_permission',
        label: '确认已完成',
        type: 'manual',
        route: () => null,
      },
    ],
  },
  {
    id: 'experience_login',
    label: '体验终端用户登录',
    steps: [
      {
        kind: 'tracked',
        id: 'end_user_login',
        label: '终端用户登录体验',
        type: 'manual',
        route: () => null,
      },
    ],
  },
]

/** Flat list of all tracked steps — used for completion counting */
export const ONBOARDING_STEPS: OnboardingTrackedStep[] = ONBOARDING_GROUPS
  .flatMap((g) => g.steps)
  .filter((s): s is OnboardingTrackedStep => s.kind === 'tracked')

/** Alias for compatibility */
export const ALL_TRACKED_STEPS = ONBOARDING_STEPS
