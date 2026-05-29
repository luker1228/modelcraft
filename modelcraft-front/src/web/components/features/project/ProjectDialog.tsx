"use client"

import { useEffect, useState } from "react"
import { useForm, type FieldPath } from "react-hook-form"
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
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@web/components/ui/form"
import { Button } from "@web/components/ui/button"
import { Input } from "@web/components/ui/input"
import { Textarea } from "@web/components/ui/textarea"
import { DatabaseConfigFields } from "@web/components/features/database/DatabaseConfigFields"
import { CheckCircle, Loader2, XCircle } from "lucide-react"
import type { Project } from "@/types"
import type { DatabaseConnectionInfo } from "@/types/cluster"

const PROJECT_SLUG_MIN_LEN = 3
const PROJECT_SLUG_MAX_LEN = 53

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

export const projectFormSchema = z.object({
  slug: z
    .string()
    .min(PROJECT_SLUG_MIN_LEN, `项目标识至少需要${PROJECT_SLUG_MIN_LEN}个字符`)
    .max(
      PROJECT_SLUG_MAX_LEN,
      `项目标识最多${PROJECT_SLUG_MAX_LEN}个字符（会用于创建 mc_private_{slug} 数据库）`,
    )
    .regex(/^[a-z][a-z0-9_]*$/, "必须以字母开头，只能包含小写字母、数字和下划线"),
  title: z.string().min(1, "项目名称不能为空").max(100),
  description: z.string().max(500).optional(),
  clusterInput: clusterInputSchema.optional(),
})

export type ProjectFormValues = z.infer<typeof projectFormSchema>

const createSchema = projectFormSchema.required({ clusterInput: true }).extend({
  clusterInput: clusterInputSchema,
})

const editSchema = projectFormSchema.omit({ clusterInput: true })

export interface ProjectConnectionTestResult {
  success: boolean
  message: string
}

interface ProjectDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  project?: Project | null
  onSubmit: (data: ProjectFormValues) => void
  onTestConnection?: (
    connectionInfo: DatabaseConnectionInfo,
  ) => Promise<ProjectConnectionTestResult>
  loading?: boolean
}

const CREATE_DEFAULTS: ProjectFormValues = {
  slug: "",
  title: "",
  description: "",
  clusterInput: {
    title: "主数据库集群",
    description: "",
    connectionInfo: { host: "", port: 3306, username: "", password: "" },
  },
}

const CREATE_STEP_1_FIELDS: FieldPath<ProjectFormValues>[] = [
  "title",
  "slug",
  "description",
]

const CREATE_STEP_2_FIELDS: FieldPath<ProjectFormValues>[] = [
  "clusterInput.title",
  "clusterInput.description",
  "clusterInput.connectionInfo.host",
  "clusterInput.connectionInfo.port",
  "clusterInput.connectionInfo.username",
  "clusterInput.connectionInfo.password",
]

const CONNECTION_TEST_FIELDS: FieldPath<ProjectFormValues>[] = [
  "clusterInput.connectionInfo.host",
  "clusterInput.connectionInfo.port",
  "clusterInput.connectionInfo.username",
  "clusterInput.connectionInfo.password",
]

const CREATE_STEPS = [
  { index: 1 as const, title: "项目信息", description: "填写项目的基础信息" },
  { index: 2 as const, title: "数据库集群", description: "配置项目默认数据库连接" },
]

function generateProjectSlug(title: string): string {
  if (!title) return ""

  const pinyinStr = pinyin(title, {
    toneType: "none",
    type: "array",
  }).join("")

  let result = pinyinStr
    .toLowerCase()
    .replace(/[\s\-]+/g, "_")
    .replace(/[^a-z0-9_]/g, "")

  if (result && !/^[a-z]/.test(result)) {
    result = "p" + result
  }

  return result.slice(0, PROJECT_SLUG_MAX_LEN)
}

export function ProjectDialog({
  open,
  onOpenChange,
  project,
  onSubmit,
  onTestConnection,
  loading = false,
}: ProjectDialogProps) {
  const isEditing = !!project
  const [currentStep, setCurrentStep] = useState<1 | 2>(1)
  const [testingConnection, setTestingConnection] = useState(false)
  const [testResult, setTestResult] = useState<ProjectConnectionTestResult | null>(null)

  const form = useForm<ProjectFormValues>({
    resolver: zodResolver(isEditing ? editSchema : createSchema),
    defaultValues: CREATE_DEFAULTS,
  })

  const isCreateMode = !isEditing
  const currentStepConfig = CREATE_STEPS[currentStep - 1]
  const createValues = form.watch()
  const canProceedStep2 = testResult?.success === true

  const handleNextStep = async () => {
    if (!isCreateMode || currentStep >= 2) return

    const fields = currentStep === 1 ? CREATE_STEP_1_FIELDS : CREATE_STEP_2_FIELDS
    const passed = await form.trigger(fields, { shouldFocus: true })
    if (!passed) return

    if (currentStep === 2 && !canProceedStep2) {
      if (!testResult) {
        setTestResult({ success: false, message: "请先测试连接并确保连接成功" })
      }
      return
    }

    setCurrentStep((currentStep + 1) as 1 | 2)
  }

  const handlePrevStep = () => {
    if (!isCreateMode || currentStep <= 1) return
    setCurrentStep((currentStep - 1) as 1 | 2)
  }

  const handleTestConnection = async () => {
    if (!isCreateMode || !onTestConnection || loading) return

    const passed = await form.trigger(CONNECTION_TEST_FIELDS, { shouldFocus: true })
    if (!passed) return

    const connectionInfo = form.getValues("clusterInput.connectionInfo")
    if (!connectionInfo) return

    setTestingConnection(true)
    setTestResult(null)

    try {
      const result = await onTestConnection(connectionInfo)
      setTestResult(result)
    } catch (error) {
      setTestResult({
        success: false,
        message: error instanceof Error ? error.message : "连接测试失败",
      })
    } finally {
      setTestingConnection(false)
    }
  }

  useEffect(() => {
    if (!open) return

    setCurrentStep(1)
    setTestingConnection(false)
    setTestResult(null)

    if (project) {
      form.reset({
        slug: project.slug,
        title: project.title,
        description: project.description ?? "",
      })
    } else {
      form.reset(CREATE_DEFAULTS)
    }
  }, [open, project, form])

  useEffect(() => {
    if (!isCreateMode || currentStep !== 2) return
    setTestResult(null)
  }, [
    isCreateMode,
    currentStep,
    createValues.clusterInput?.connectionInfo.host,
    createValues.clusterInput?.connectionInfo.port,
    createValues.clusterInput?.connectionInfo.username,
    createValues.clusterInput?.connectionInfo.password,
  ])

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
            {isCreateMode && (
              <div className="space-y-3 rounded-md border border-border bg-muted/30 p-3">
                <div className="flex items-center justify-between">
                  <p className="text-xs text-muted-foreground">步骤 {currentStep}/2</p>
                  <p className="text-xs text-muted-foreground">{currentStepConfig.title}</p>
                </div>
                <p className="text-sm font-medium text-foreground">{currentStepConfig.description}</p>
                <div className="grid grid-cols-2 gap-2">
                  {CREATE_STEPS.map((step) => {
                    const isActive = currentStep === step.index
                    const isDone = currentStep > step.index
                    return (
                      <div
                        key={step.index}
                        className={`h-1.5 rounded-full ${isActive || isDone ? "bg-primary" : "bg-border"}`}
                      />
                    )
                  })}
                </div>
              </div>
            )}

            {(isEditing || currentStep === 1) && (
              <>
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
              </>
            )}

            {!isEditing && currentStep === 2 && (
              <>
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

                <div className="flex items-center justify-between gap-3 rounded-md border border-border bg-muted/20 px-3 py-2">
                  <div className="min-h-[20px] text-sm">
                    {testResult && (
                      <span
                        className={`flex items-center gap-1.5 ${testResult.success ? "text-[#059669]" : "text-[#ef4444]"}`}
                      >
                        {testResult.success ? (
                          <CheckCircle className="size-4 shrink-0" strokeWidth={1.5} />
                        ) : (
                          <XCircle className="size-4 shrink-0" strokeWidth={1.5} />
                        )}
                        {testResult.message}
                      </span>
                    )}
                  </div>

                  <Button
                    type="button"
                    variant="outline"
                    onClick={handleTestConnection}
                    disabled={loading || testingConnection || !onTestConnection}
                    className="h-9 px-4 text-sm font-medium"
                  >
                    {testingConnection && (
                      <Loader2 className="mr-1.5 size-4 animate-spin" strokeWidth={1.5} />
                    )}
                    测试连接
                  </Button>
                </div>
              </>
            )}

            {isEditing ? (
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
                  {loading ? "保存中..." : "保存修改"}
                </Button>
              </DialogFooter>
            ) : (
              <DialogFooter className="pt-2">
                {currentStep === 1 ? (
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => onOpenChange(false)}
                    disabled={loading}
                  >
                    取消
                  </Button>
                ) : (
                  <Button
                    type="button"
                    variant="outline"
                    onClick={handlePrevStep}
                    disabled={loading}
                  >
                    上一步
                  </Button>
                )}

                {currentStep < 2 ? (
                  <Button
                    type="button"
                    onClick={handleNextStep}
                    disabled={loading || (currentStep === 2 && !canProceedStep2)}
                  >
                    下一步
                  </Button>
                ) : (
                  <Button type="submit" disabled={loading}>
                    {loading ? "创建中..." : "创建项目"}
                  </Button>
                )}
              </DialogFooter>
            )}
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
