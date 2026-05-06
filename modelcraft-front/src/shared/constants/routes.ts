export const TENANT_LOGIN_PATH = '/tenant/login'
export const TENANT_REGISTER_PATH = '/register'

export function getEndUserLoginPath(orgName: string): string {
  return `/end-user/${orgName}/login`
}
