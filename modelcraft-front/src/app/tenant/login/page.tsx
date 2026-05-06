// src/app/tenant/login/page.tsx
// 租户端登录页。复用 /login 的登录表单逻辑，路由迁移至 /tenant/login。
import type { Metadata } from 'next'
import LoginPage from '@/app/login/page'

export const metadata: Metadata = {
  title: '租户登录 — ModelCraft',
  robots: {
    index: false,
    follow: false,
  },
}

// 复用 LoginPage 组件（username/password 表单，调用 /api/auth/login BFF）
export default LoginPage
