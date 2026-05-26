export type RouteCatalogEntry = {
  /** 路由模板，使用 :param 占位符，例如 /org/:orgName/project/:projectSlug/models */
  routeTemplate: string
  /** 页面标题（中文，AI 用来匹配意图） */
  title: string
  /** 页面功能描述（AI 判断跳转依据） */
  description: string
  /** 关键词列表（触发跳转的语义词） */
  keywords: string[]
}

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
  },
  {
    routeTemplate: '/org/:orgName/project/:projectSlug/model-editor',
    title: '数据模型编辑器',
    description: '创建和管理数据模型、字段结构，查看模型数据记录',
    keywords: ['模型', '字段', '数据模型', '模型编辑器', '新建模型', '字段管理'],
  },
  {
    routeTemplate: '/org/:orgName/project/:projectSlug/enums',
    title: '枚举管理',
    description: '管理项目中的枚举类型，限制字段的可选值范围',
    keywords: ['枚举', 'enum', '枚举值', '枚举类型'],
  },
  {
    routeTemplate: '/org/:orgName/project/:projectSlug/rbac/roles',
    title: 'RBAC 角色管理',
    description: '管理项目内的角色与权限包，控制用户对数据的增删改查权限',
    keywords: ['权限', 'RBAC', '角色', '权限管理', '角色配置'],
  },
  {
    routeTemplate: '/org/:orgName/project/:projectSlug/rbac/bundles',
    title: '权限包管理',
    description: '管理权限包版本，配置细粒度操作权限',
    keywords: ['权限包', 'bundle', '权限版本', '权限快照'],
  },
  {
    routeTemplate: '/org/:orgName/project/:projectSlug/end-users',
    title: '终端用户管理',
    description: '管理访问本项目的终端用户账号',
    keywords: ['终端用户', 'end user', '用户管理', '外部用户'],
  },
  {
    routeTemplate: '/org/:orgName/project/:projectSlug/settings',
    title: '项目设置',
    description: '修改项目基本信息、归档或删除项目、管理数据库集群',
    keywords: ['项目设置', '集群配置', '数据库连接', '项目信息', '设置'],
  },
  {
    routeTemplate: '/org/:orgName/developers',
    title: '成员管理',
    description: '管理组织内的开发者成员和角色（Owner / Admin / Member）',
    keywords: ['成员', '开发者', '邀请成员', '成员管理'],
  },
  {
    routeTemplate: '/org/:orgName/end-users',
    title: '终端用户（Org 级）',
    description: '管理组织下所有终端用户账号',
    keywords: ['终端用户', 'org 级用户', '用户账号'],
  },
  {
    routeTemplate: '/org/:orgName/settings',
    title: '组织设置',
    description: '配置组织基础信息和安全设置',
    keywords: ['组织设置', 'org 设置', '组织信息'],
  },
]
