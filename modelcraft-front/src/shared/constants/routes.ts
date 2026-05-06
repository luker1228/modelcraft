export const TENANT_LOGIN_PATH = '/login'
export const TENANT_REGISTER_PATH = '/register'

export function getEndUserLoginPath(orgName: string): string {
  return `/end-user/${orgName}/login`
}
