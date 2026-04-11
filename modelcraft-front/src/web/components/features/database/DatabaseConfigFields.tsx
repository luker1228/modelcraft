"use client"

import { useState } from "react"
import { Eye, EyeOff, Globe, User, Lock } from "lucide-react"
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@web/components/ui/form"
import { Input } from "@web/components/ui/input"
import type { FieldValues, UseFormReturn } from "react-hook-form"

// ── Types ─────────────────────────────────────────────────────────────────────

export interface DatabaseConfigFieldsProps<T extends FieldValues = FieldValues> {
  /** 表单实例 */
  form: UseFormReturn<T>
  /** 字段前缀，如 "clusterInput.connectionInfo" */
  fieldPrefix?: string
  /** 是否显示密码切换按钮 */
  showPasswordToggle?: boolean
  /** 是否显示图标 */
  showIcons?: boolean
  /** 是否显示必填标记 */
  showRequired?: boolean
  /** 主机地址字段名，默认 "host" */
  hostFieldName?: string
  /** 端口字段名，默认 "port" */
  portFieldName?: string
  /** 用户名字段名，默认 "username" */
  usernameFieldName?: string
  /** 密码字段名，默认 "password" */
  passwordFieldName?: string
  /** 服务端密码占位符，当字段值等于此值时以 placeholder 样式展示 */
  encryptedByServerPlaceholder?: string
}

// 辅助函数：包装 FormField 为 form-row 布局
function FormRow({
  label,
  description,
  showIcons = false,
  showRequired = false,
  icon: Icon,
  children,
}: {
  label: string
  description?: string
  showIcons?: boolean
  showRequired?: boolean
  icon?: React.ComponentType<{ className?: string }>
  children: React.ReactNode
}) {
  return (
    <div className="flex items-start justify-between gap-4 border-b border-gray-100 py-3.5 last:border-0">
      <div className="w-[38%] flex-none">
        <div className="flex items-center gap-1.5">
          {showIcons && Icon && (
            <Icon className="size-3.5 flex-shrink-0 text-muted-foreground" />
          )}
          <span className="text-sm font-medium text-foreground">
            {label} {showRequired && <span className="text-xs text-red-500">*</span>}
          </span>
        </div>
        {description && <p className="mt-1 text-xs text-muted-foreground">{description}</p>}
      </div>
      <div className="flex-1">{children}</div>
    </div>
  )
}

// ── Component ─────────────────────────────────────────────────────────────────

/**
 * 数据库连接配置字段组件
 * 包含：主机地址、端口、用户名、密码
 *
 * 当密码值等于 encryptedByServerPlaceholder 时，输入框以 placeholder 样式展示
 * "已设置密码，输入新密码可修改"，而不是将占位符字符串直接显示在输入框中。
 * 用户聚焦密码框时，自动清空为空字符串进入编辑模式；失去焦点且未输入内容时
 * 恢复为占位符，表示保持原有密码不变。
 */
export function DatabaseConfigFields<T extends FieldValues = FieldValues>({
  form,
  fieldPrefix = "",
  showPasswordToggle = true,
  showIcons = true,
  showRequired = true,
  hostFieldName = "host",
  portFieldName = "port",
  usernameFieldName = "username",
  passwordFieldName = "password",
  encryptedByServerPlaceholder,
}: DatabaseConfigFieldsProps<T>) {
  const [showPassword, setShowPassword] = useState(false)

  // 构建完整字段名
  const getFieldPath = (fieldName: string) => {
    return fieldPrefix ? `${fieldPrefix}.${fieldName}` : fieldName
  }

  const passwordPath = getFieldPath(passwordFieldName)

  return (
    <>
      {/* 主机地址和端口 - 同行 */}
      <div className="flex items-start justify-between gap-4 border-b border-gray-100 py-3.5">
        <div className="w-[38%] flex-none">
          <div className="flex items-center gap-1.5">
            {showIcons && (
              <Globe className="size-3.5 flex-shrink-0 text-muted-foreground" />
            )}
            <span className="text-sm font-medium text-foreground">
              主机地址 {showRequired && <span className="text-xs text-red-500">*</span>}
            </span>
          </div>
        </div>
        <div className="grid flex-1 grid-cols-3 gap-3">
          <div className="col-span-2">
            <FormField
              control={form.control}
              name={getFieldPath(hostFieldName) as never}
              render={({ field }) => (
                <FormItem>
                  <FormControl>
                    <Input placeholder="localhost 或 db.example.com" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
          <FormField
            control={form.control}
            name={getFieldPath(portFieldName) as never}
            render={({ field }) => (
              <FormItem>
                <FormControl>
                  <Input
                    type="number"
                    placeholder="3306"
                    className="font-mono"
                    {...field}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
        </div>
      </div>

      {/* 用户名 */}
      <div className="flex items-start justify-between gap-4 border-b border-gray-100 py-3.5">
        <div className="w-[38%] flex-none">
          <div className="flex items-center gap-1.5">
            {showIcons && (
              <User className="size-3.5 flex-shrink-0 text-muted-foreground" />
            )}
            <span className="text-sm font-medium text-foreground">
              用户名 {showRequired && <span className="text-xs text-red-500">*</span>}
            </span>
          </div>
        </div>
        <div className="flex-1">
          <FormField
            control={form.control}
            name={getFieldPath(usernameFieldName) as never}
            render={({ field }) => (
              <FormItem>
                <FormControl>
                  <Input placeholder="root" className="font-mono" {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
        </div>
      </div>

      {/* 密码 */}
      <FormField
        control={form.control}
        name={passwordPath as never}
        render={({ field }) => {
          const isServerPlaceholder =
            encryptedByServerPlaceholder !== undefined &&
            field.value === encryptedByServerPlaceholder

          return (
            <div className="flex items-start justify-between gap-4 border-b border-gray-100 py-3.5 last:border-0">
              <div className="w-[38%] flex-none">
                <div className="flex items-center gap-1.5">
                  {showIcons && (
                    <Lock className="size-3.5 flex-shrink-0 text-muted-foreground" />
                  )}
                  <span className="text-sm font-medium text-foreground">
                    密码 {showRequired && <span className="text-xs text-red-500">*</span>}
                  </span>
                </div>
              </div>
              <div className="flex-1">
                <FormItem>
                  <FormControl>
                    <div className="relative">
                      <Input
                        type={showPassword ? "text" : "password"}
                        placeholder={isServerPlaceholder ? "已设置密码，输入新密码可修改" : "数据库密码"}
                        className={`${showPasswordToggle ? "pr-10" : ""} ${isServerPlaceholder ? "italic placeholder:text-muted-foreground/70" : ""}`}
                        {...field}
                        value={isServerPlaceholder ? "" : (field.value as string)}
                        onFocus={(e) => {
                          if (isServerPlaceholder) {
                            field.onChange("")
                          }
                          field.onBlur()
                          e.target.focus()
                        }}
                        onBlur={(e) => {
                          if (
                            encryptedByServerPlaceholder !== undefined &&
                            e.target.value === ""
                          ) {
                            field.onChange(encryptedByServerPlaceholder)
                          }
                          field.onBlur()
                        }}
                        onChange={(e) => {
                          field.onChange(e.target.value)
                        }}
                      />
                      {showPasswordToggle && (
                        <button
                          type="button"
                          className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                          onClick={() => setShowPassword(!showPassword)}
                        >
                          {showPassword ? (
                            <EyeOff className="size-4" />
                          ) : (
                            <Eye className="size-4" />
                          )}
                        </button>
                      )}
                    </div>
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </div>
            </div>
          )
        }}
      />
    </>
  )
}
