import { z } from 'zod'
import { isReservedUserName } from '@/shared/constants/reserved-usernames'
import type { IdentifierType } from '@/types/auth'

/** 手机号格式：11 位国内号码 */
export const phoneNumberSchema = z
  .string()
  .regex(/^1[3-9]\d{9}$/, '请输入有效的 11 位手机号')

/** 密码：最少 8 位 */
export const passwordSchema = z.string().min(8, '密码至少需要 8 位')

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

/** 登录标识符 schema（支持手机号或用户名） */
export const identifierSchema = z.string().min(1, '请输入手机号或用户名')

/** 登录表单 */
export const loginFormSchema = z
  .object({
    identifier: identifierSchema,
    identifierType: z.enum(['PHONE', 'USERNAME'] as const),
    password: z.string().min(1, '请输入密码'),
  })
  .superRefine((data, ctx) => {
    const errorMessage = validateIdentifier(data.identifier, data.identifierType)
    if (errorMessage) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: errorMessage,
        path: ['identifier'],
      })
    }
  })

/** 注册表单 */
export const registerFormSchema = z
  .object({
    phone: phoneNumberSchema,
    userName: userNameSchema,
    password: passwordSchema,
    confirmPassword: z.string().min(1, '请确认密码'),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: '两次输入的密码不一致',
    path: ['confirmPassword'],
  })

export type LoginFormValues = z.infer<typeof loginFormSchema>
export type RegisterFormValues = z.infer<typeof registerFormSchema>

/**
 * 根据标识符类型校验输入值
 * @returns 错误消息，如果校验通过则返回 null
 */
export function validateIdentifier(
  identifier: string,
  type: IdentifierType
): string | null {
  if (type === 'PHONE') {
    const result = phoneNumberSchema.safeParse(identifier)
    if (!result.success) {
      return result.error.errors[0]?.message ?? '手机号格式不正确'
    }
  }
  // USERNAME 类型不做前端强校验（允许已注册的各种格式），交给后端验证
  return null
}
