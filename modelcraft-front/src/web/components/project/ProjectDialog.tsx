"use client"

import { useEffect } from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import * as z from "zod"
import { pinyin } from "pinyin-pro"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@web/components/ui/dialog"
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@web/components/ui/form"
import { Button } from "@web/components/ui/button"
import { Input } from "@web/components/ui/input"
import { Textarea } from "@web/components/ui/textarea"
import { Checkbox } from "@web/components/ui/checkbox"
import { Separator } from "@web/components/ui/separator"
import { DatabaseConfigFields } from "@web/components/database/DatabaseConfigFields"
import type { Project } from "@/types"

// ── Zod Schema ────────────────────────────────────────────────────────────────

const clusterConnectionSchema = z.object({
  host: z.string().min(1, "主机地址不能为空").max(255),
  port: z.coerce.number().int("端口必须为整数").min(1024).max(65535),
  username: z.string().min(1, "用户名不能为空").max(36),
  password: z.string().min(1, "密码不能为空").max(36),
})

const clusterInputSchema = z.object({
  title: z.string().min(1, "集群名称不能为空").max(100),
  description: z.string().max(500).optional(),
  connectionInfo: clusterConnectionSchema,
})

// Single source-of-truth schema. clusterInput is required on create, stripped on edit.
export const projectFormSchema = z.object({
  slug: z
    .string()
    .min(3, "项目标识至少需要3个字符")
    .max(50)
    .regex(/^[a-z][a-z0-9_]*$/, "必须以字母开头，只能包含小写字母、数字和下划线"),
  title: z.string().min(1, "项目名称不能为空").max(100),
  description: z.string().max(500).optional(),
  loginUrl: z.string().url("请输入有效的 URL").optional().or(z.literal("")),
  clusterInput: clusterInputSchema.optional(),
  skipConnectionTest: z.boolean().default(false),
})

export type ProjectFormValues = z.infer<typeof projectFormSchema>

// Runtime refinement: require clusterInput when creating
const createSchema = projectFormSchema.required({ clusterInput: true }).extend({
  clusterInput: clusterInputSchema,
})

const editSchema = projectFormSchema.omit({ clusterInput: true, skipConnectionTest: true })

// ── Props ─────────────────────────────────────────────────────────────────────

interface ProjectDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  project?: Project | null
  onSubmit: (data: ProjectFormValues) => void
  loading?: boolean
}

// ── Default values ────────────────────────────────────────────────────────────

const CREATE_DEFAULTS: ProjectFormValues = {
  slug: "",
  title: "",
  description: "",
  loginUrl: "",
  clusterInput: {
    title: "主数据库集群",
    description: "",
    connectionInfo: { host: "", port: 3306, username: "", password: "" },
  },
  skipConnectionTest: false,
}

// ── Helper: Generate project name from title ─────────────────────────────────

/**
 * 根据项目名称生成项目标识
 * 规则：中文转拼音，必须以字母开头，只允许小写字母、数字和下划线
 */
function generateProjectSlug(title: string): string {
  if (!title) return ""

  // 使用拼音库转换中文
  const pinyinStr = pinyin(title, {
    toneType: "none",
    type: "array",
  }).join("")

  // 转小写，移除空格和连字符，只保留小写字母、数字和下划线
  let result = pinyinStr
    .toLowerCase()
    .replace(/[\s\-]+/g, "_") // 空格和连字符转为下划线
    .replace(/[^a-z0-9_]/g, "") // 只保留小写字母、数字和下划线

  // 确保以字母开头
  if (result && !/^[a-z]/.test(result)) {
    result = "p" + result // 如果开头不是字母，添加前缀 "p"
  }

  return result
}

// ── Component ─────────────────────────────────────────────────────────────────

export function ProjectDialog({
  open,
  onOpenChange,
  project,
  onSubmit,
  loading = false,
}: ProjectDialogProps) {
  const isEditing = !!project

  const form = useForm<ProjectFormValues>({
    resolver: zodResolver(isEditing ? editSchema : createSchema),
    defaultValues: CREATE_DEFAULTS,
  })

  useEffect(() => {
    if (!open) return
    if (project) {
      form.reset({
        slug: project.slug,
        title: project.title,
        description: project.description ?? "",
        loginUrl: project.loginUrl ?? "",
      })
    } else {
      form.reset(CREATE_DEFAULTS)
    }
  }, [open, project, form])

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-[580px]">
        <DialogHeader>
          <DialogTitle>{isEditing ? "编辑项目" : "创建新项目"}</DialogTitle>
          <DialogDescription>
            {isEditing
              ? "修改项目的基本信息"
              : "填写项目信息和数据库集群配置以创建新的数据建模项目"}
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4 py-2">

            {/* ── 项目信息 ── */}
            <FormField
              control={form.control}
              name="title"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>项目名称 *</FormLabel>
                  <FormControl>
                    <Input
                      placeholder="我的项目"
                      {...field}
                      onChange={(e) => {
                        const newTitle = e.target.value
                        field.onChange(newTitle)
                        // 自动生成项目标识（仅在创建模式下）
                        if (!isEditing) {
                          const generatedSlug = generateProjectSlug(newTitle)
                          form.setValue("slug", generatedSlug)
                        }
                      }}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="slug"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>项目标识 *</FormLabel>
                  <FormControl>
                    <Input
                      placeholder="myproject"
                      disabled={isEditing}
                      className="font-mono"
                      {...field}
                    />
                  </FormControl>
                  {!isEditing && (
                    <FormDescription>
                      用于 URL 和 API 调用，创建后不可修改
                    </FormDescription>
                  )}
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="description"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>项目描述</FormLabel>
                  <FormControl>
                    <Textarea
                      placeholder="描述此项目的用途（可选）"
                      rows={2}
                      className="resize-none"
                      {...field}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="loginUrl"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>登录页面 URL</FormLabel>
                  <FormControl>
                    <Input placeholder="https://example.com/login" {...field} />
                  </FormControl>
                  <FormDescription>可选，用于项目的自定义登录跳转</FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            {/* ── 集群配置（仅创建时） ── */}
            {!isEditing && (
              <>
                <Separator />
                <p className="pt-1 text-sm font-semibold text-foreground">数据库集群配置</p>

                <div className="grid grid-cols-2 gap-4">
                  <FormField
                    control={form.control}
                    name="clusterInput.title"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>集群名称 *</FormLabel>
                        <FormControl>
                          <Input placeholder="主数据库集群" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>

                <FormField
                  control={form.control}
                  name="clusterInput.description"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>集群描述</FormLabel>
                      <FormControl>
                        <Textarea
                          placeholder="描述此集群的用途（可选）"
                          rows={2}
                          className="resize-none"
                          {...field}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <DatabaseConfigFields
                  form={form}
                  fieldPrefix="clusterInput.connectionInfo"
                />

                <FormField
                  control={form.control}
                  name="skipConnectionTest"
                  render={({ field }) => (
                    <FormItem className="flex items-center gap-2 space-y-0">
                      <FormControl>
                        <Checkbox
                          checked={field.value}
                          onCheckedChange={field.onChange}
                        />
                      </FormControl>
                      <div>
                        <FormLabel className="cursor-pointer font-normal">
                          跳过连接测试
                        </FormLabel>
                        <FormDescription>
                          不验证数据库连接即可创建项目
                        </FormDescription>
                      </div>
                    </FormItem>
                  )}
                />
              </>
            )}

            <DialogFooter className="pt-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => onOpenChange(false)}
                disabled={loading}
              >
                取消
              </Button>
              <Button type="submit" disabled={loading}>
                {loading ? "保存中..." : isEditing ? "保存修改" : "创建项目"}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
