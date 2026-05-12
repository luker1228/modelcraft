import type { OnboardingStepId } from './storage'

export interface OnboardingSubStep {
  id: OnboardingStepId
  label: string
  description: string
  route: (params: { orgName: string; projectSlug: string | null }) => string | null
}

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
        id: 'create_project',
        label: '创建项目',
        description: '在组织下创建第一个项目',
        route: ({ orgName }) => `/org/${orgName}/workspace`,
      },
    ],
  },
  {
    id: 'design_model',
    label: '创建模型',
    steps: [
      {
        id: 'create_model',
        label: '创建模型',
        description: '在项目中定义第一个数据模型',
        route: ({ orgName, projectSlug }) =>
          projectSlug ? `/org/${orgName}/project/${projectSlug}/model-editor` : null,
      },
      {
        id: 'add_field',
        label: '添加字段',
        description: '为模型添加至少一个字段',
        route: ({ orgName, projectSlug }) =>
          projectSlug ? `/org/${orgName}/project/${projectSlug}/model-editor` : null,
      },
    ],
  },
  {
    id: 'configure_permissions',
    label: '创建权限及角色',
    steps: [
      {
        id: 'apply_preset',
        label: '应用权限预设',
        description: '为模型选择终端用户权限预设策略',
        route: ({ orgName, projectSlug }) =>
          projectSlug ? `/org/${orgName}/project/${projectSlug}/rbac/permissions` : null,
      },
      {
        id: 'create_role',
        label: '创建角色',
        description: '创建一个终端用户角色',
        route: ({ orgName, projectSlug }) =>
          projectSlug ? `/org/${orgName}/project/${projectSlug}/rbac/roles` : null,
      },
    ],
  },
  {
    id: 'add_users',
    label: '创建用户并体验',
    steps: [
      {
        id: 'add_end_user',
        label: '添加终端用户',
        description: '在组织中创建第一个终端用户账号',
        route: ({ orgName }) => `/org/${orgName}/end-users`,
      },
      {
        id: 'assign_role',
        label: '分配角色',
        description: '将角色授予终端用户',
        route: ({ orgName }) => `/org/${orgName}/end-users`,
      },
      {
        id: 'end_user_login',
        label: '终端用户体验',
        description: '以终端用户身份登录，验证整条链路',
        route: () => null,
      },
    ],
  },
]

/** Flat list of all sub-steps in order — used by storage & context for completion tracking */
export const ONBOARDING_STEPS: OnboardingSubStep[] = ONBOARDING_GROUPS.flatMap((g) => g.steps)
