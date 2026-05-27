'use client'

import NextLink from 'next/link'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Loader2 } from 'lucide-react'
import { registerFormSchema, type RegisterFormValues } from '@/shared/validation/auth'
import { useRegister } from '@/web/hooks/auth/use-auth-form'
import { TENANT_LOGIN_PATH } from '@shared/constants/routes'
import { AuthLayout } from '@/web/components/features/auth/auth-layout'
import { Button } from '@web/components/ui/button'
import { Input } from '@web/components/ui/input'
import { PasswordInput } from '@/web/components/common/password-input'
import {
  Form,
  FormField,
  FormItem,
  FormLabel,
  FormControl,
  FormMessage,
  FormDescription,
} from '@web/components/ui/form'

export default function RegisterPage() {
  const { register, isLoading, error } = useRegister()

  const form = useForm<RegisterFormValues>({
    resolver: zodResolver(registerFormSchema),
    defaultValues: { phone: '', userName: '', password: '', confirmPassword: '' },
  })

  const handleSubmit = form.handleSubmit(async (values) => {
    await register(values)
  })

  return (
    <AuthLayout title="创建账号" subtitle="注册新账号">
      <Form {...form}>
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          {/* Server error banner */}
          {error && (
            <div className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">
              {error}
            </div>
          )}

          <FormField
            control={form.control}
            name="phone"
            render={({ field }) => (
              <FormItem>
                <FormLabel>手机号</FormLabel>
                <FormControl>
                  <Input placeholder="请输入手机号" autoComplete="tel" {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="userName"
            render={({ field }) => (
              <FormItem>
                <FormLabel>用户名</FormLabel>
                <FormControl>
                  <Input
                    placeholder="字母/数字/_-，不能数字开头"
                    autoComplete="username"
                    {...field}
                  />
                </FormControl>
                <FormDescription className="text-xs">
                  3-32 位，注册后不可修改，将作为登录凭证
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="password"
            render={({ field }) => (
              <FormItem>
                <FormLabel>密码</FormLabel>
                <FormControl>
                  <PasswordInput
                    placeholder="至少 8 位，需包含字母和数字"
                    autoComplete="new-password"
                    {...field}
                  />
                </FormControl>
                <FormDescription className="text-xs">
                  密码至少 8 位，且必须包含至少一个字母和一个数字
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="confirmPassword"
            render={({ field }) => (
              <FormItem>
                <FormLabel>确认密码</FormLabel>
                <FormControl>
                  <PasswordInput placeholder="请再次输入密码" autoComplete="new-password" {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <Button type="submit" className="mt-1 h-10 w-full" disabled={isLoading}>
            {isLoading && <Loader2 className="mr-2 size-4 animate-spin" />}
            注册
          </Button>

          <p className="text-center text-sm text-muted-foreground">
            已有账号？{' '}
            <NextLink href={TENANT_LOGIN_PATH} className="font-medium text-primary hover:underline">
              立即登录
            </NextLink>
          </p>
        </form>
      </Form>
    </AuthLayout>
  )
}
