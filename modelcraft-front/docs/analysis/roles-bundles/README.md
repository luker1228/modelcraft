# Roles/Bundles 页面分析文档

> URL: `/org/[orgName]/project/[projectSlug]/roles/bundles/[bundleId]`

## 📚 文档导航

### 🚀 快速开始
**首先阅读**: [QUICK_REFERENCE.md](./QUICK_REFERENCE.md)
- 4 个问题的直接答案
- 核心文件地图
- FAQ 常见问题

### 📖 深入学习
1. **[SUMMARY.md](./SUMMARY.md)** - 深度总结
   - 详细的答案解析
   - 组件架构说明
   - 关键实现要点

2. **[architecture-diagram.md](./architecture-diagram.md)** - 架构与数据流
   - 完整的系统架构图
   - 数据流向说明
   - 性能考虑

3. **[code-snippets-reference.md](./code-snippets-reference.md)** - 代码速查
   - 完整的代码片段
   - GraphQL 查询定义
   - 常见用例示例

## 🎯 核心文件三角

```
页面文件 (展示层)
└─ src/app/org/[orgName]/project/[projectSlug]/
   roles/bundles/[bundleId]/page.tsx

Hook 文件 (数据层)
└─ src/app/.../rbac/bundles/_hooks/useBundleManage.ts

GraphQL 文件 (API 层)
└─ src/api-client/rbac/graphql-docs.ts
```

## 🔍 4 个核心问题

### 1. 页面文件路径？
`modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/roles/bundles/[bundleId]/page.tsx`

### 2. 权限点显示组件？
第 99-146 行的 Permissions Container div
- 权限点列表容器
- 权限点项目卡片
- 空状态/加载状态/列表状态

### 3. 空状态处理？
第 109-115 行，显示盾牌图标 + "暂无关联权限点" 文字

### 4. GraphQL 查询？
`src/api-client/rbac/graphql-docs.ts`
- GET_END_USER_BUNDLE
- GET_END_USER_PERMISSIONS
- REMOVE_PERMISSION_FROM_BUNDLE

## 💡 快速提示

### 权限点名称显示优先级
```
displayName || modelDisplayName || modelId
```

### 删除权限点流程
1. 点击删除按钮
2. 显示确认对话框
3. 调用 removePermission mutation
4. 自动 refetch 相关查询
5. 页面自动更新

### 标签映射
- action: SELECT→查询, INSERT→新增, UPDATE→更新, DELETE→删除, EXPORT→导出
- rowScope: ALL→全部, SELF→本人, DEPT→本部门, DEPT_AND_CHILDREN→部门及子部门

## 📊 项目结构

```
modelcraft-front/
├── src/
│   ├── app/org/[orgName]/project/[projectSlug]/
│   │   ├── roles/bundles/[bundleId]/
│   │   │   └── page.tsx ⭐ 主页面
│   │   └── rbac/bundles/_hooks/
│   │       └── useBundleManage.ts ⭐ 数据 Hook
│   ├── api-client/rbac/
│   │   └── graphql-docs.ts ⭐ GraphQL 定义
│   └── types/
│       └── rbac.ts (类型定义)
└── docs/analysis/roles-bundles/ (你在这里)
```

## 🚀 后续可扩展的功能

- [ ] 添加权限点
- [ ] 权限点排序
- [ ] 批量删除
- [ ] 权限点搜索
- [ ] 权限点详情面板

