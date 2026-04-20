'use client'

import { useState } from 'react'
import { useForm, type FieldValues } from 'react-hook-form'
import { IdentityFormSection } from '@web/components/ui/identity-form-section'
import { Badge } from '@web/components/ui/badge'
import { Separator } from '@web/components/ui/separator'

// ── Sections ───────────────────────────────────────────────────────────────────

const SECTIONS = [
  { id: 'cluster',   label: '数据库集群' },
  { id: 'model',     label: '数据模型'   },
  { id: 'enum',      label: '枚举类型'   },
  { id: 'view',      label: '查看模式'   },
  { id: 'readonly',  label: '附加字段'   },
  { id: 'prefix',    label: 'URL 前缀'   },
  { id: 'status',    label: '状态变体'   },
]

// ── Section Variants ───────────────────────────────────────────────────────────

function ClusterSection() {
  const { control, formState } = useForm<FieldValues>({ defaultValues: { title: '主数据库集群' } })
  const dirty = !!formState.dirtyFields.title
  const [saving, setSaving] = useState(false)

  const handleSave = () => {
    setSaving(true)
    setTimeout(() => setSaving(false), 1500)
  }

  return (
    <IdentityFormSection
      title="基本信息"
      displayNameField={{
        name: 'title',
        label: '显示名称',
        placeholder: '主数据库集群',
        description: '用于在界面上展示的可读名称',
        control,
      }}
      identifierField={{
        name: 'name',
        label: '集群标识符',
        value: 'prod-mysql-cluster',
        copyable: true,
        description: '创建后不可修改',
      }}
      saveDisabled={!dirty}
      saveLoading={saving}
      onSave={handleSave}
      onCancel={() => {}}
    />
  )
}

function ModelSection() {
  const { control, formState } = useForm<FieldValues>({ defaultValues: { title: '用户' } })
  const dirty = !!formState.dirtyFields.title
  const [saving, setSaving] = useState(false)

  return (
    <IdentityFormSection
      title="模型信息"
      displayNameField={{
        name: 'title',
        label: '模型展示名称',
        description: '用于在数据模型列表中展示',
        control,
      }}
      identifierField={{
        name: 'name',
        label: 'API 标识符',
        value: 'User',
        copyable: true,
        regeneratable: true,
        onRegenerate: () => {},
        description: '用于 GraphQL Schema 和 REST API',
      }}
      identifierStatus={{ type: 'auto' }}
      saveDisabled={!dirty}
      saveLoading={saving}
      onSave={() => { setSaving(true); setTimeout(() => setSaving(false), 1200) }}
      onCancel={() => {}}
    />
  )
}

function EnumSection() {
  const { control, formState } = useForm<FieldValues>({ defaultValues: { title: '用户角色' } })
  const dirty = !!formState.dirtyFields.title

  return (
    <IdentityFormSection
      title="枚举信息"
      displayNameField={{
        name: 'title',
        label: '枚举名称',
        placeholder: '如：用户角色、订单状态',
        control,
      }}
      identifierField={{
        name: 'name',
        label: '枚举标识符',
        value: 'UserRole',
        copyable: true,
      }}
      identifierStatus={{ type: 'manual' }}
      saveDisabled={!dirty}
      onSave={() => {}}
      onCancel={() => {}}
    />
  )
}

function ViewSection() {
  return (
    <IdentityFormSection
      mode="view"
      title="集群信息（查看模式）"
      displayNameField={{
        name: 'title',
        label: '显示名称',
        value: '生产环境 MySQL 集群',
      }}
      identifierField={{
        name: 'name',
        label: '集群标识符',
        value: 'prod-mysql-cluster',
        copyable: true,
      }}
      showActions={false}
    />
  )
}

function ReadOnlySection() {
  const { control, formState } = useForm<FieldValues>({ defaultValues: { title: '用户表' } })
  const dirty = !!formState.dirtyFields.title

  return (
    <IdentityFormSection
      title="模型信息（附加只读字段）"
      displayNameField={{
        name: 'title',
        label: '模型展示名称',
      }}
      readOnlyFields={[
        {
          label: 'API 标识符',
          value: 'user_table',
          copyable: true,
        },
        {
          label: '数据库表名',
          value: 'mc_user_table',
          description: '由系统自动映射',
          copyable: true,
        },
        {
          label: '字段数量',
          value: (
            <span className="font-mono text-sm font-medium text-foreground">12 个字段</span>
          ),
        },
      ]}
      saveDisabled={!dirty}
      onSave={() => {}}
      onCancel={() => {}}
    />
  )
}

function PrefixSection() {
  const { control, formState } = useForm<FieldValues>({ defaultValues: { title: '官方文档站' } })
  const dirty = !!formState.dirtyFields.title

  return (
    <IdentityFormSection
      title="站点信息（URL 前缀）"
      displayNameField={{
        name: 'title',
        label: '站点名称',
        placeholder: '如：官方文档站、API 参考',
        control,
      }}
      identifierField={{
        name: 'slug',
        label: '站点路径',
        value: 'docs',
        prefix: 'modelcraft.io/',
        copyable: true,
        description: '将出现在公开访问 URL 中',
      }}
      saveDisabled={!dirty}
      onSave={() => {}}
      onCancel={() => {}}
    />
  )
}

function StatusSection() {
  const statuses = [
    { type: 'auto'    as const, label: '自动生成' },
    { type: 'manual'  as const, label: '手动设置' },
    { type: 'loading' as const, label: '生成中'   },
    { type: 'error'   as const, label: '生成失败', message: '标识符已被占用' },
  ]
  const { control } = useForm<FieldValues>({ defaultValues: { title: '资源名称' } })

  return (
    <div className="space-y-4">
      {statuses.map((s) => (
        <IdentityFormSection
          key={s.type}
          displayNameField={{ name: 'title', label: '显示名称', control }}
          identifierField={{
            name: 'name',
            label: '标识符',
            value: 'example-resource',
            copyable: true,
          }}
          identifierStatus={{ type: s.type, message: s.message }}
          showActions={false}
        />
      ))}
    </div>
  )
}

// ── Page ───────────────────────────────────────────────────────────────────────

const SECTION_MAP: Record<string, { title: string; desc: string; component: React.FC }> = {
  cluster:  { title: '数据库集群',           desc: '标题变更后保存按钮才启用',         component: ClusterSection  },
  model:    { title: '数据模型（可重新生成）', desc: '标识符支持重新生成 + 复制',         component: ModelSection    },
  enum:     { title: '枚举类型',             desc: '标识符状态为手动设置',             component: EnumSection     },
  view:     { title: '查看模式',             desc: '只读展示，不渲染操作栏',           component: ViewSection     },
  readonly: { title: '附加只读字段',          desc: 'readOnlyFields 额外展示元数据',   component: ReadOnlySection },
  prefix:   { title: 'URL 前缀',            desc: 'identifierField.prefix 前缀显示', component: PrefixSection   },
  status:   { title: '标识符状态变体',        desc: 'auto / manual / loading / error', component: StatusSection   },
}

export default function IdentitySettingsPage() {
  const [activeSection, setActiveSection] = useState('cluster')
  const section = SECTION_MAP[activeSection]
  const SectionComponent = section.component

  return (
    <div className="h-full overflow-auto bg-white">
      <div className="mx-auto max-w-7xl p-6">

        {/* Page header */}
        <div className="mb-8">
          <h1 className="font-heading text-xl font-semibold tracking-tight text-foreground">
            身份配置
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">
            统一管理集群、模型、枚举等资源的显示名称与标识符配置。
          </p>
        </div>

        <div className="flex items-start gap-6">
          {/* Sidebar navigation */}
          <nav className="w-44 shrink-0 overflow-hidden rounded-lg border border-border bg-white">
            <div className="border-b border-border px-3 py-2">
              <p className="font-mono text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
                配置分组
              </p>
            </div>
            <ul className="py-1">
              {SECTIONS.map((item) => (
                <li key={item.id}>
                  <button
                    onClick={() => setActiveSection(item.id)}
                    className={`w-full rounded-none px-3 py-2 text-left text-sm transition-colors ${
                      activeSection === item.id
                        ? 'bg-muted font-semibold text-foreground'
                        : 'text-muted-foreground hover:bg-muted/50 hover:text-foreground'
                    }`}
                  >
                    {item.label}
                  </button>
                </li>
              ))}
            </ul>
          </nav>

          {/* Content panel */}
          <div className="min-w-0 flex-1 space-y-4">
            <div className="flex items-center gap-3">
              <h2 className="font-heading text-base font-semibold text-foreground">
                {section.title}
              </h2>
              <Badge variant="secondary" className="font-mono text-xs font-normal">
                {activeSection}
              </Badge>
            </div>
            <p className="text-sm text-muted-foreground">{section.desc}</p>
            <Separator />
            <SectionComponent />
          </div>
        </div>

        {/* Field reference */}
        <div className="mt-12 overflow-hidden rounded-lg border border-border bg-white">
          <div className="border-b border-border bg-muted/30 px-6 py-3">
            <h3 className="font-heading text-sm font-semibold text-foreground">配置项说明</h3>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border bg-muted/20">
                  <th className="whitespace-nowrap px-6 py-3 text-left font-semibold text-foreground">Prop</th>
                  <th className="whitespace-nowrap px-4 py-3 text-left font-semibold text-foreground">类型</th>
                  <th className="px-4 py-3 text-left font-semibold text-foreground">说明</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {[
                  ['title',             'string',                          '配置区块标题，显示在卡片顶部'],
                  ['mode',              "'edit' | 'view'",                 '显示模式，默认 edit'],
                  ['displayNameField',  '{ name, label, control, value… }','显示名称字段配置'],
                  ['identifierField',   '{ name, value, copyable, … }',    '标识符字段配置（只读）'],
                  ['identifierStatus',  '{ type, message? }',              "auto / manual / loading / error"],
                  ['readOnlyFields',    'ReadOnlyField[]',                  '额外只读字段列表'],
                  ['showActions',       'boolean',                         '是否显示操作栏，默认 true'],
                  ['saveDisabled',      'boolean',                         '保存按钮禁用，默认 false'],
                  ['saveLoading',       'boolean',                         '保存按钮加载态'],
                  ['onSave',            '() => void',                      '保存回调'],
                  ['onCancel',          '() => void',                      '取消回调（传入后才显示取消按钮）'],
                  ['saveText',          'string',                          '保存按钮文字，默认「保存」'],
                  ['cancelText',        'string',                          '取消按钮文字，默认「取消」'],
                  ['actions',           'ReactNode',                       '自定义操作按钮（覆盖默认）'],
                ].map(([prop, type, desc]) => (
                  <tr key={prop} className="hover:bg-muted/20">
                    <td className="whitespace-nowrap px-6 py-3 font-mono font-medium text-foreground">{prop}</td>
                    <td className="whitespace-nowrap px-4 py-3 font-mono text-xs text-muted-foreground">{type}</td>
                    <td className="px-4 py-3 text-muted-foreground">{desc}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

      </div>
    </div>
  )
}
