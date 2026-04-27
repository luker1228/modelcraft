export const routes = {
  home: '/',
  login: '/login',
  register: '/register',
  workspace: (orgName: string) => `/org/${orgName}/workspace`,
  project: (orgName: string, projectSlug: string) =>
    `/org/${orgName}/project/${projectSlug}`,
} as const

export const AUTH_STORAGE_STATE = 'e2e/.auth/user.json'
