export type RouteCatalogEntry = {
  /** 路由模板，使用 :param 占位符，例如 /org/:orgName/project/:projectSlug/models */
  routeTemplate: string
  /** 页面标题（中文，AI 用来匹配意图） */
  title: string
  /** 页面功能描述（AI 判断跳转依据） */
  description: string
  /** 关键词列表（触发跳转的语义词） */
  keywords: string[]
  /** 典型用户表达示例（AI 意图匹配的最强信号） */
  examples?: string[]
  /** 是否需要项目上下文（projectSlug / projectName） */
  requiresProject: boolean
}

export const PROJECT_REQUIRED_ROUTE_POLICY = {
  appliesTo: 'routeCatalog 中 requiresProject=true 的页面，以及 list_databases/list_models/get_model_fields/query_model 等项目级工具',
  beforeAction: '每次执行前先调用 list_projects，不依赖历史会话中的 projectSlug',
  whenProjectsExist: '用 ui_present_proposal 让用户选择项目，再基于所选项目生成目标 route',
  whenNoProjects: '不要调用项目级工具；推荐用户先到项目列表创建项目',
  createProjectRouteTemplate: '/org/:orgName/dashboard',
} as const

/**
 * All navigable pages in ModelCraft.
 * Agent reads this via useCopilotReadable to decide which page to navigate to.
 * Routes with :param are resolved at runtime using current org/project context.
 */
export const ROUTE_CATALOG: RouteCatalogEntry[] = [
  {
    routeTemplate: '/org/:orgName/dashboard',
    title: '项目列表',
    description: '查看、搜索、创建和管理组织下的所有项目',
    keywords: ['项目列表', '所有项目', 'workspace', '项目管理'],
    examples: [
      '看一下我有哪些项目',
      '帮我创建一个新项目',
      '我想切换到另一个项目',
      '项目在哪里',
      '返回项目列表',
    ],
    requiresProject: false,
  },
  {
    routeTemplate: '/org/:orgName/project/:projectSlug/model-editor',
    title: '模型管理',
    description: '创建和管理数据模型、字段结构、模型字段与模型配置',
    keywords: ['模型', '字段', '数据模型', '模型编辑器', '新建模型', '字段管理', '模型管理'],
    examples: [
      '帮我新建一个 user 模型',
      '我要给 order 模型加一个字段',
      '查看一下 product 模型的结构',
      '模型列表在哪',
      '我想修改模型的字段',
      '删掉这个模型',
    ],
    requiresProject: true,
  },
  {
    routeTemplate: '/org/:orgName/project/:projectSlug/model-editor?view=data&db=:databaseName',
    title: '数据管理',
    description: '查看、查询和管理模型数据记录',
    keywords: ['数据管理', '数据记录', '记录管理', '查看数据', '查询数据', '模型数据'],
    examples: [
      '看一下 user 表里有哪些数据',
      '帮我查一下订单记录',
      '这个模型现在有多少条数据',
      '我想往 product 里插入一条记录',
      '数据在哪里看',
    ],
    requiresProject: true,
  },
  {
    routeTemplate: '/org/:orgName/project/:projectSlug/databases',
    title: '数据库管理',
    description: '接管项目使用的 MySQL 数据库，设置访问模式（托管/自建），触发同步将数据库表导入为数据模型',
    keywords: ['数据库', '数据库管理', '接管数据库', '数据库接管', '接管', '数据库同步', '同步数据库', 'MySQL', '托管', '自建'],
    examples: [
      '我想接管 demo_pm',
      '帮我把 test_db 接管进来',
      '去数据库管理页',
      '同步一下数据库',
      '我要把数据库的表导入为模型',
      '接管数据库在哪里操作',
      '查看已接管的数据库',
    ],
    requiresProject: true,
  },
  {
    routeTemplate: '/org/:orgName/project/:projectSlug/enums',
    title: '枚举管理',
    description: '管理项目中的枚举类型，限制字段的可选值范围',
    keywords: ['枚举', 'enum', '枚举值', '枚举类型'],
    examples: [
      '帮我创建一个 status 枚举',
      '查看项目里有哪些枚举',
      '我要给枚举加一个选项',
      '枚举在哪里管理',
    ],
    requiresProject: true,
  },
  {
    routeTemplate: '/org/:orgName/project/:projectSlug/access-control',
    title: 'RLS 策略管理',
    description: '管理项目内各模型的 RLS 行级安全策略，按 action + role 匹配',
    keywords: ['RLS', '策略', '行级安全', '权限', 'policy', 'action', 'role'],
    examples: [
      '帮我查看 User 模型的 RLS 策略',
      '创建一个 admin 角色的 read 策略',
      '删除某条 RLS 策略',
      'RLS 策略管理在哪',
    ],
    requiresProject: true,
  },
  {
    routeTemplate: '/org/:orgName/project/:projectSlug/settings',
    title: '项目设置',
    description: '修改项目基本信息、归档或删除项目；配置数据库集群连接（MySQL 主机、端口、账号密码），是使用数据库接管功能的前提',
    keywords: ['项目设置', '集群配置', '数据库连接', '配置数据库', '数据库集群', '连接配置', 'MySQL 连接', '项目信息', '设置'],
    examples: [
      '帮我配置一下数据库连接',
      '怎么填 MySQL 的连接信息',
      '修改项目名称',
      '我要删除这个项目',
      '数据库集群在哪里配置',
      '接管数据库前需要先配置什么',
    ],
    requiresProject: true,
  },
  {
    routeTemplate: '/org/:orgName/developers',
    title: '成员管理',
    description: '管理组织内的开发者成员和角色（Owner / Admin / Member）',
    keywords: ['成员', '开发者', '邀请成员', '成员管理'],
    examples: [
      '帮我邀请一个新成员',
      '查看组织里有哪些开发者',
      '我要把某人的角色改成 Admin',
      '移除一个成员',
    ],
    requiresProject: false,
  },
  {
    routeTemplate: '/org/:orgName/settings',
    title: '组织设置',
    description: '配置组织基础信息和安全设置',
    keywords: ['组织设置', 'org 设置', '组织信息'],
    examples: [
      '修改一下组织名称',
      '去组织设置',
      '配置组织的安全策略',
    ],
    requiresProject: false,
  },
]
