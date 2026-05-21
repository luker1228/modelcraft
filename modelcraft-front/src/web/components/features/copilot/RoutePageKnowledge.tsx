'use client'

import { memo, useMemo } from 'react'
import { usePathname } from 'next/navigation'
import { useCopilotReadable } from '@copilotkit/react-core'

/**
 * Page-level knowledge per route segment.
 * Keep descriptions concise — they're injected into every AI context window.
 */
const PAGE_KNOWLEDGE: Record<string, { name: string; description: string; workflow: string }> = {
  'model-editor': {
    name: '模型编辑器',
    description: '管理项目的数据模型结构。左侧栏：数据库选择器 + 模型列表；主区域：选中模型后展示数据记录；右侧抽屉：字段详情与编辑。',
    workflow: `
1. 用数据库选择器选择目标数据库
2. 在左侧模型列表中点击模型，右侧主区域显示该模型的数据记录
3. 点击模型行末的设置图标（→）打开右侧字段管理抽屉
4. 在抽屉中：新建字段（+ 按钮）、编辑字段、废弃字段、删除已废弃字段
5. 支持在主区域直接新建 / 编辑数据记录`.trim(),
  },
  'enums': {
    name: '枚举管理',
    description: '管理项目中的枚举类型（Enum）。枚举可绑定到字段，限制字段的可选值范围。',
    workflow: `
1. 查看现有枚举列表
2. 新建枚举：填写枚举名称，添加枚举值
3. 编辑枚举：修改枚举值列表
4. 删除枚举前需确保无字段依赖它`.trim(),
  },
  'roles': {
    name: '角色权限管理 (RBAC)',
    description: '管理项目内的角色与权限包。角色可分配给终端用户，控制其对数据的增删改查权限。',
    workflow: `
1. 查看角色列表，点击角色进入详情
2. 在角色详情中配置权限包（Permission Bundle）
3. 创建新角色，绑定权限包
4. 权限包版本化：快照只读，最多 5 个版本，支持回滚`.trim(),
  },
  'end-user-access': {
    name: '用户授权',
    description: '为终端用户（EndUser）分配项目访问权限和角色。终端用户在 Org 级别创建，在项目级别授权。',
    workflow: `
1. 查看已被授权访问本项目的终端用户列表
2. 添加授权：选择已有终端用户，分配角色
3. 修改角色：更改用户在本项目中的权限角色
4. 撤销授权：移除用户对本项目的访问`.trim(),
  },
  'identity-settings': {
    name: '身份认证设置',
    description: '配置终端用户的登录认证方式，例如邮箱密码、第三方 OAuth 等。',
    workflow: `
1. 查看当前启用的认证方式
2. 开启 / 关闭认证方式
3. 配置 OAuth 应用参数（Client ID / Secret）`.trim(),
  },
  'settings': {
    name: '项目设置',
    description: '修改项目基本信息，包括名称、描述、状态（活跃 / 归档）。',
    workflow: `
1. 修改项目名称或描述，点击保存
2. 归档项目：项目将不再出现在活跃列表中
3. 删除项目：不可恢复，需二次确认`.trim(),
  },
  'workspace': {
    name: '所有项目',
    description: '组织下的项目列表。支持网格 / 列表视图切换，可搜索、新建、编辑、归档项目。',
    workflow: `
1. 点击项目卡片进入该项目（跳转到 model-editor）
2. 新建项目：填写名称、slug、数据库连接信息
3. 项目卡片右上角菜单：编辑 / 归档 / 删除`.trim(),
  },
  'end-users': {
    name: '终端用户管理',
    description: '管理组织下的所有终端用户账号。终端用户是通过 API 访问数据的外部用户（非管理员）。',
    workflow: `
1. 查看组织下所有终端用户
2. 新建终端用户：设置用户名、邮箱、密码
3. 查看用户详情：已授权的项目及角色
4. 重置密码、禁用账号`.trim(),
  },
  'developers': {
    name: '成员管理',
    description: '管理组织内的开发者成员账号与角色（Owner / Admin / Member）。',
    workflow: `
1. 查看成员列表
2. 邀请新成员：输入邮箱发送邀请
3. 修改成员角色
4. 移除成员`.trim(),
  },
  'cluster': {
    name: '数据库集群',
    description: '管理项目连接的数据库集群。每个集群对应一个 MySQL 数据库连接，模型在集群下的数据库中创建。',
    workflow: `
1. 查看已配置的数据库集群列表
2. 新建集群：填写 host / port / user / password / database
3. 测试连接：验证连接信息是否正确
4. 删除集群前需确保无模型依赖`.trim(),
  },
}

/**
 * Detects the current route and injects page-level knowledge into CopilotKit context.
 * Must be mounted inside a <CopilotKit> provider tree.
 * Updates automatically when the pathname changes (Next.js navigation).
 */
export const RoutePageKnowledge = memo(function RoutePageKnowledge() {
  const pathname = usePathname()

  const pageInfo = useMemo(() => {
    // Match the last meaningful segment of the path
    const segments = pathname.split('/').filter(Boolean)
    // Walk from the end; skip dynamic segments that look like IDs or slugs if needed
    for (let i = segments.length - 1; i >= 0; i--) {
      const seg = segments[i]
      if (PAGE_KNOWLEDGE[seg]) return { seg, ...PAGE_KNOWLEDGE[seg] }
    }
    return null
  }, [pathname])

  useCopilotReadable({
    description: '当前页面信息',
    value: pageInfo
      ? {
          route: pathname,
          pageName: pageInfo.name,
          pageDescription: pageInfo.description,
          pageWorkflow: pageInfo.workflow,
        }
      : { route: pathname, pageName: '未知页面', pageDescription: '' },
  })

  return null
})
