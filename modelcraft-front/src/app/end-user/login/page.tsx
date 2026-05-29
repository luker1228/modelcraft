'use client'

import { useSearchParams } from 'next/navigation'
import { Loader2 } from 'lucide-react'
import { Form } from '@web/components/ui/form'
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@web/components/ui/form'
import { AuthLayout } from '@/web/components/features/auth/auth-layout'
import { PasswordInput } from '@/web/components/common/password-input'
import { Button } from '@web/components/ui/button'
import { Input } from '@web/components/ui/input'
import { useEndUserGlobalLoginForm } from '@web/hooks/end-user-auth-v2/useEndUserGlobalLoginForm'

export default function EndUserLoginPage() {
  const searchParams = useSearchParams()
  const redirectTo = searchParams.get('redirect')
  const { form, onSubmit, isLoading, error } = useEndUserGlobalLoginForm(redirectTo)

  return (
    <AuthLayout
      title="欢迎回来"
      subtitle="输入用户名和密码后自动进入所属组织"
      backLink={{ href: '/', label: '返回登录选择' }}
    >
      <Form {...form}>
        <form onSubmit={onSubmit} className="flex flex-col gap-5">
          {error && (
            <div className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">
              {error}
            </div>
          )}

          <FormField
            control={form.control}
            name="username"
            render={({ field }) => (
              <FormItem>
                <FormLabel>用户名</FormLabel>
                <FormControl>
                  <Input
                    placeholder="请输入用户名"
                    autoComplete="username"
                    disabled={isLoading}
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
                  <PasswordInput
                    placeholder="请输入密码"
                    autoComplete="current-password"
                    disabled={isLoading}
                    {...field}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <p className="text-sm text-muted-foreground">
            无需预先选择组织，登录成功后系统会自动跳转到您的工作区。
          </p>

          <Button type="submit" className="w-full" disabled={isLoading}>
            {isLoading && <Loader2 className="mr-2 size-4 animate-spin" />}
            用户登录
          </Button>
        </form>
      </Form>
    </AuthLayout>
  )
}
