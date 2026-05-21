'use client'

import { useCopilotAction } from '@copilotkit/react-core'
import { useRouter } from 'next/navigation'
import { useOnboarding } from '@shared/onboarding/OnboardingContext'
import type { DevelopRecordWorkspaceAIRef } from '@web/components/features/model-editor/model-record-form/DevelopRecordWorkspace'

interface ProjectCopilotActionsProps {
  orgName: string
  projectSlug: string
  /** 由 WorkspaceAIRefContext 提供的命令接口 */
  workspaceAiRef?: React.MutableRefObject<DevelopRecordWorkspaceAIRef | null>
}

export function ProjectCopilotActions({ orgName, projectSlug, workspaceAiRef }: ProjectCopilotActionsProps) {
  const router = useRouter()
  const { setPendingAction } = useOnboarding()

  useCopilotAction({
    name: 'navigate_to_org',
    description: '退出当前项目，返回组织 workspace 页面',
    parameters: [],
    handler: async () => {
      router.push(`/org/${orgName}/workspace`)
      return '已返回组织页面'
    },
  })

  useCopilotAction({
    name: 'navigate_to_model',
    description: '跳转到模型编辑器，查看或编辑指定模型的字段结构',
    parameters: [
      { name: 'db', type: 'string', description: '数据库名称', required: true },
      { name: 'model', type: 'string', description: '模型名称', required: true },
    ],
    handler: async ({ db, model }: { db: string; model: string }) => {
      router.push(`/org/${orgName}/project/${projectSlug}/model-editor?db=${db}&model=${model}`)
      return `已跳转到模型 ${model} 的编辑器`
    },
  })

  useCopilotAction({
    name: 'navigate_to_data',
    description: '跳转到数据视图，查看指定模型的数据记录',
    parameters: [
      { name: 'db', type: 'string', description: '数据库名称', required: true },
      { name: 'model', type: 'string', description: '模型名称', required: true },
    ],
    handler: async ({ db, model }: { db: string; model: string }) => {
      router.push(`/org/${orgName}/project/${projectSlug}/data?db=${db}&model=${model}`)
      return `已跳转到模型 ${model} 的数据视图`
    },
  })

  useCopilotAction({
    name: 'open_create_model',
    description: '打开新建模型的表单，可预填名称和描述。用户需手动点击 Create 完成创建。',
    parameters: [
      { name: 'db', type: 'string', description: '数据库名称', required: true },
      { name: 'name', type: 'string', description: '模型名称（英文小写+下划线）', required: false },
      { name: 'title', type: 'string', description: '模型显示名称', required: false },
    ],
    handler: async ({ db, name, title }: { db: string; name?: string; title?: string }) => {
      router.push(
        `/org/${orgName}/project/${projectSlug}/model-editor?db=${db}&openCreate=1&prefillName=${name ?? ''}&prefillTitle=${title ?? ''}`
      )
      return `已打开新建模型表单${name ? `，名称预填为 ${name}` : ''}，请确认后点击 Create。`
    },
  })

  useCopilotAction({
    name: 'open_create_record',
    description: '打开新建记录表单，并预填指定字段值。用户需手动点击 Save 完成创建。写操作不会自动执行。',
    parameters: [
      { name: 'model', type: 'string', description: '模型名称', required: true },
      { name: 'db', type: 'string', description: '数据库名称', required: true },
      { name: 'prefill', type: 'object', description: '要预填的字段值，例如 {"name": "张三", "age": 25}', required: false },
    ],
    handler: async (args: { model: string; db: string; prefill?: object }) => {
      const ref = workspaceAiRef?.current
      if (ref) {
        ref.openCreate((args.prefill as Record<string, unknown>) ?? {})
        return `已打开新建 ${args.model} 记录的表单${args.prefill ? '，字段已预填' : ''}，请确认后点击 Save。`
      }
      router.push(`/org/${orgName}/project/${projectSlug}/data?db=${args.db}&model=${args.model}`)
      return `已跳转到 ${args.model} 数据页，请在页面加载后重新调用此工具以预填表单。`
    },
  })

  useCopilotAction({
    name: 'open_edit_record',
    description: '打开指定记录的编辑表单，并预填要修改的字段。用户需手动点击 Save 完成保存。写操作不会自动执行。',
    parameters: [
      { name: 'model', type: 'string', description: '模型名称', required: true },
      { name: 'db', type: 'string', description: '数据库名称', required: true },
      { name: 'record_id', type: 'string', description: '要编辑的记录 ID', required: true },
      { name: 'patch', type: 'object', description: '要修改的字段值，例如 {"amount": 950}', required: true },
    ],
    handler: async (args: { model: string; db: string; record_id: string; patch: object }) => {
      const ref = workspaceAiRef?.current
      if (ref) {
        await ref.openEdit(args.record_id, args.patch as Record<string, unknown>)
        return `已打开记录 ${args.record_id} 的编辑表单，修改字段已预填，请确认后点击 Save。`
      }
      router.push(`/org/${orgName}/project/${projectSlug}/data?db=${args.db}&model=${args.model}`)
      return `已跳转到 ${args.model} 数据页，请在页面加载后重新调用此工具。`
    },
  })

  useCopilotAction({
    name: 'highlight_records',
    description: '在数据表格中高亮指定记录行，并显示说明原因（鼠标悬停可见）',
    parameters: [
      { name: 'model', type: 'string', description: '模型名称', required: true },
      { name: 'record_ids', type: 'string[]', description: '要高亮的记录 ID 列表', required: true },
      { name: 'reason', type: 'string', description: '高亮原因', required: true },
    ],
    handler: async (args: { model: string; record_ids: string[]; reason: string }) => {
      const ref = workspaceAiRef?.current
      if (ref) {
        const reasons = args.record_ids.reduce<Record<string, string>>(
          (acc, id) => { acc[id] = args.reason; return acc },
          {}
        )
        ref.setHighlight(args.record_ids, reasons)
        return `已高亮 ${args.record_ids.length} 条 ${args.model} 记录：${args.reason}`
      }
      return '当前页面没有数据表格，请先导航到数据视图。'
    },
  })

  useCopilotAction({
    name: 'navigate_to_enums',
    description: '跳转到枚举管理页',
    parameters: [
      { name: 'db', type: 'string', description: '数据库名称（可选）', required: false },
    ],
    handler: async ({ db }: { db?: string }) => {
      const query = db ? `?db=${db}` : ''
      router.push(`/org/${orgName}/project/${projectSlug}/enums${query}`)
      return '已跳转到枚举管理页'
    },
  })

  useCopilotAction({
    name: 'navigate_to_cluster',
    description: '跳转到数据库集群配置页',
    parameters: [],
    handler: async () => {
      router.push(`/org/${orgName}/project/${projectSlug}/settings`)
      return '已跳转到集群配置页'
    },
  })

  useCopilotAction({
    name: 'navigate_to_rbac',
    description: '跳转到 RBAC 权限管理页',
    parameters: [
      {
        name: 'section',
        type: 'string',
        description: 'roles | users | bundles | permissions（默认 roles）',
        required: false,
      },
    ],
    handler: async ({ section }: { section?: string }) => {
      const sub = section ?? 'roles'
      router.push(`/org/${orgName}/project/${projectSlug}/rbac/${sub}`)
      return `已跳转到 RBAC ${sub} 页`
    },
  })

  useCopilotAction({
    name: 'navigate_to_end_users',
    description: '跳转到 end-user 管理页',
    parameters: [],
    handler: async () => {
      router.push(`/org/${orgName}/project/${projectSlug}/end-users`)
      return '已跳转到 end-user 管理页'
    },
  })

  // ── 引导工具：通过高亮 UI 元素指引用户操作 ────────────────────────────────

  useCopilotAction({
    name: 'guide_select_database',
    description: '高亮左侧数据库选择器，引导用户先选择一个数据库。适用于需要用户先选 DB 才能继续的步骤（如新建模型）。',
    parameters: [
      {
        name: 'reason',
        type: 'string',
        description: '为什么需要选择数据库，会以 toast 形式提示用户',
        required: false,
      },
    ],
    handler: async ({ reason }: { reason?: string }) => {
      setPendingAction('select_database')
      return reason
        ? `已高亮数据库选择器：${reason}`
        : '已高亮数据库选择器，请在左侧点击选择一个数据库'
    },
  })

  useCopilotAction({
    name: 'guide_create_model',
    description: '高亮左侧"新建模型"按钮，引导用户点击创建。必须在用户已选好数据库之后调用。',
    parameters: [
      {
        name: 'reason',
        type: 'string',
        description: '引导说明',
        required: false,
      },
    ],
    handler: async ({ reason }: { reason?: string }) => {
      setPendingAction('nav_create_model')
      return reason
        ? `已高亮新建模型按钮：${reason}`
        : '已高亮新建模型按钮，请点击它开始创建'
    },
  })

  return null
}
