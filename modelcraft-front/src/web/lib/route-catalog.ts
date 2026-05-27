export type RouteCatalogEntry = {
  /** 路由模板，使用 :param 占位符，例如 /org/:orgName/project/:projectSlug/models */
  routeTemplate: string
  /** 页面标题（中文，AI 用来匹配意图） */
  title: string
  /** 页面功能描述（AI 判断跳转依据） */
  description: string
  /** 关键词列表（触发跳转的语义词） */
  keywords: string[]
  /** 是否需要项目上下文（projectSlug / projectName） */
  requiresProject: boolean
}

export const PROJECT_REQUIRED_ROUTE_POLICY = {
  appliesTo: 'routeCatalog 中 requiresProject=true 的页面，以及 list_databases/list_models/get_model_fields/query_model 等项目级工具',
  beforeAction: '每次执行前先调用 list_projects，不依赖历史会话中的 projectSlug',
  whenProjectsExist: '用 ui_present_proposal 让用户选择项目，再基于所选项目生成目标 route',
  whenNoProjects: '不要调用项目级工具；推荐用户先到项目列表创建项目',
  createProjectRouteTemplate: '/org/:orgName/workspace',
} as const

/**
 * All navigable pages in ModelCraft.
 * Agent reads this via useCopilotReadable to decide which page to navigate to.
 * Routes with :param are resolved at runtime using current org/project context.
 */
export const ROUTE_CATALOG: RouteCatalogEntry[] = [
  {
    routeTemplate: '/org/:orgName/workspace',
    title: '项目列表',
    description: '查看、搜索、创建和管理组织下的所有项目',
    keywords: ['项目列表', '所有项目', 'workspace', '项目管理'],
    requiresProject: false,
  },
  {
    routeTemplate: '/org/:orgName/project/:projectSlug/model-editor',
    title: '模型管理',
    description: '创建和管理数据模型、字段结构、模型字段与模型配置',
    keywords: ['模型', '字段', '数据模型', '模型编辑器', '新建模型', '字段管理', '模型管理'],
    requiresProject: true,
  },
  {
    routeTemplate: '/org/:orgName/project/:projectSlug/model-editor?view=data',
    title: '数据管理',
    description: '查看、查询和管理模型数据记录',
    keywords: ['数据管理', '数据记录', '记录管理', '查看数据', '查询数据', '模型数据'],
    requiresProject: true,
  },
  {
    routeTemplate: '/org/:orgName/project/:projectSlug/enums',
    title: '枚举管理',
    description: '管理项目中的枚举类型，限制字段的可选值范围',
    keywords: ['枚举', 'enum', '枚举值', '枚举类型'],
    requiresProject: true,
  },
  {
    routeTemplate: '/org/:orgName/project/:projectSlug/access-control?tab=roles',
    title: 'RBAC 角色管理',
    description: '管理项目内的角色与权限包，控制用户对数据的增删改查权限',
    keywords: ['权限', 'RBAC', '角色', '权限管理', '角色配置'],
    requiresProject: true,
  },
  {
    routeTemplate: '/org/:orgName/project/:projectSlug/access-control?tab=bundles',
    title: '权限包管理',
    description: '管理权限包版本，配置细粒度操作权限',
    keywords: ['权限包', 'bundle', '权限版本', '权限快照'],
    requiresProject: true,
  },
  {
    routeTemplate: '/org/:orgName/project/:projectSlug/access-control?tab=permissions',
    title: '权限点管理',
    description: '查看和管理项目内的权限点定义',
    keywords: ['权限点', '权限项', 'permission', 'permissions', '操作权限', '权限定义'],
    requiresProject: true,
  },
  {
    routeTemplate: '/org/:orgName/project/:projectSlug/end-users',
    title: '终端用户管理',
    description: '管理访问本项目的终端用户账号',
    keywords: ['终端用户', 'end user', '用户管理', '外部用户'],
    requiresProject: true,
  },
  {
    routeTemplate: '/org/:orgName/project/:projectSlug/settings',
    title: '项目设置',
    description: '修改项目基本信息、归档或删除项目、管理数据库集群',
    keywords: ['项目设置', '集群配置', '数据库连接', '项目信息', '设置'],
    requiresProject: true,
  },
  {
    routeTemplate: '/org/:orgName/developers',
    title: '成员管理',
    description: '管理组织内的开发者成员和角色（Owner / Admin / Member）',
    keywords: ['成员', '开发者', '邀请成员', '成员管理'],
    requiresProject: false,
  },
  {
    routeTemplate: '/org/:orgName/end-users',
    title: '终端用户（Org 级）',
    description: '管理组织下所有终端用户账号',
    keywords: ['终端用户', 'org 级用户', '用户账号'],
    requiresProject: false,
  },
  {
    routeTemplate: '/org/:orgName/settings',
    title: '组织设置',
    description: '配置组织基础信息和安全设置',
    keywords: ['组织设置', 'org 设置', '组织信息'],
    requiresProject: false,
  },
]
