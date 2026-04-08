import { z } from 'zod'

/** 手机号格式：11 位国内号码 */
export const phoneNumberSchema = z
  .string()
  .regex(/^1[3-9]\d{9}$/, '请输入有效的 11 位手机号')

/** 密码：最少 8 位 */
export const passwordSchema = z
  .string()
  .min(8, '密码至少需要 8 位')

/** 登录表单 */
export const loginFormSchema = z.object({
  phone: phoneNumberSchema,
  password: z.string().min(1, '请输入密码'),
})

/** 注册表单 */
export const registerFormSchema = z.object({
  phone: phoneNumberSchema,
  password: passwordSchema,
  confirmPassword: z.string().min(1, '请确认密码'),
}).refine((data) => data.password === data.confirmPassword, {
  message: '两次输入的密码不一致',
  path: ['confirmPassword'],
})

export type LoginFormValues = z.infer<typeof loginFormSchema>
export type RegisterFormValues = z.infer<typeof registerFormSchema>
