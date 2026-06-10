import { z } from 'zod'
import { isReservedUserName } from '@/shared/constants/reserved-usernames'

/** 手机号格式：11 位国内号码 */
export const phoneNumberSchema = z
  .string()
  .regex(/^1[3-9]\d{9}$/, '请输入有效的 11 位手机号')

/** 密码：最少 8 位，且必须包含字母和数字 */
export const passwordSchema = z
  .string()
  .min(8, '密码至少需要 8 位')
  .regex(/[A-Za-z]/, '密码必须至少包含一个字母')
  .regex(/\d/, '密码必须至少包含一个数字')

/**
 * 用户名规则：
 * - 英文字母/数字/_/-
 * - 不能以数字开头
 * - 长度 3–32 位
 * - 不能是保留字
 */
export const userNameSchema = z
  .string()
  .min(3, '用户名至少需要 3 个字符')
  .max(32, '用户名最多 32 个字符')
  .regex(
    /^[a-zA-Z_-][a-zA-Z0-9_-]{2,31}$/,
    '用户名只能包含字母、数字、下划线和连字符，且不能以数字开头'
  )
  .refine((val) => !isReservedUserName(val), {
    message: '该用户名为系统保留字，请换一个',
  })

/** 组织名 slug：6–24 位，小写字母/数字/下划线，且必须以字母开头 */
export const orgNameSchema = z
  .string()
  .min(6, '组织名至少 6 个字符')
  .max(24, '组织名最多 24 个字符')
  .regex(/^[a-z][a-z0-9_]*$/, '只能包含小写字母、数字和下划线，且必须以字母开头')

/** 组织展示名称：1–64 个字符 */
export const orgDisplayNameSchema = z
  .string()
  .min(1, '请输入组织名称')
  .max(64, '组织名称最多 64 个字符')

/** 登录表单 — 仅手机号 */
export const loginFormSchema = z.object({
  phone: phoneNumberSchema,
  password: z.string().min(1, '请输入密码'),
})

/** 注册表单 */
export const registerFormSchema = z
  .object({
    phone: phoneNumberSchema,
    userName: userNameSchema,
    orgDisplayName: orgDisplayNameSchema,
    orgName: orgNameSchema,
    password: passwordSchema,
    confirmPassword: z.string().min(1, '请确认密码'),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: '两次输入的密码不一致',
    path: ['confirmPassword'],
  })

export type LoginFormValues = z.infer<typeof loginFormSchema>
export type RegisterFormValues = z.infer<typeof registerFormSchema>
