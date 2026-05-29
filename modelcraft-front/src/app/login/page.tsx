import { redirect } from 'next/navigation'
import { TENANT_LOGIN_PATH } from '@shared/constants/routes'

export default function LoginPage() {
  redirect(TENANT_LOGIN_PATH)
}
