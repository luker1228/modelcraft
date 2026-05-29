'use client'

import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Loader2 } from 'lucide-react'
import { loginFormSchema, type LoginFormValues } from '@/shared/validation/auth'
import { useLogin } from '@/web/hooks/auth/use-auth-form'
import { AuthLayout } from '@/web/components/features/auth/auth-layout'
import { Button } from '@web/components/ui/button'
import { Input } from '@web/components/ui/input'
import { PasswordInput } from '@/web/components/common/password-input'
import { Tabs, TabsList, TabsTrigger } from '@web/components/ui/tabs'
import {
  Form,
  FormField,
  FormItem,
  FormLabel,
  FormControl,
  FormMessage,
} from '@web/components/ui/form'
import type { IdentifierType } from '@/types/auth'

export default function TenantLoginPage() {
  const { login, isLoading, error, identifierType, setIdentifierType } =
    useLogin()

  const form = useForm<LoginFormValues>({
    resolver: zodResolver(loginFormSchema),
    defaultValues: { identifier: '', identifierType: 'USERNAME', password: '' },
  })

  // 切换登录类型时清空输入并更新 form 值
  const handleTabChange = (value: string) => {
    const type = value as IdentifierType
    setIdentifierType(type)
    form.setValue('identifierType', type)
    form.setValue('identifier', '')
    form.clearErrors('identifier')
  }

  const handleSubmit = form.handleSubmit(async (values) => {
    await login(values)
  })

  return (
    <AuthLayout
      title="欢迎回来，管理员"
      subtitle="登录管理控制台"
      backLink={{ href: '/', label: '返回登录选择' }}
    >
      <Form {...form}>
        <form onSubmit={handleSubmit} className="flex flex-col gap-5">
          {/* Server error banner */}
          {error && (
            <div className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">
              {error}
            </div>
          )}

          {/* 登录方式切换 */}
          <Tabs
            value={identifierType}
            onValueChange={handleTabChange}
            className="w-full"
          >
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="USERNAME">用户名登录</TabsTrigger>
              <TabsTrigger value="PHONE">手机号登录</TabsTrigger>
            </TabsList>
          </Tabs>

          <FormField
            control={form.control}
            name="identifier"
            render={({ field }) => (
              <FormItem>
                <FormLabel>
                  {identifierType === 'PHONE' ? '手机号' : '用户名'}
                </FormLabel>
                <FormControl>
                  <Input
                    placeholder={
                      identifierType === 'PHONE'
                        ? '请输入手机号'
                        : '请输入用户名'
                    }
                    autoComplete={identifierType === 'PHONE' ? 'tel' : 'username'}
                    {...field}
                  />
                </FormControl>
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
                  <PasswordInput placeholder="请输入密码" autoComplete="current-password" {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <Button
            type="submit"
            className="mt-1 h-10 w-full"
            disabled={isLoading}
          >
            {isLoading && <Loader2 className="mr-2 size-4 animate-spin" />}
            登录
          </Button>
        </form>
      </Form>
    </AuthLayout>
  )
}
