# ModelCraft 前端项目文档索引

## 📚 核心文档

### 工作流程
- [HTML-First Prototype 工作流](.codebuddy/skills/html-first-prototype/SKILL.md) - UI 开发核心流程
- [Tailwind 配置同步指南](TAILWIND_SYNC_GUIDE.md) - 保持原型和项目配置同步
- [快速同步参考](QUICK_SYNC_REFERENCE.md) - 一行命令快速参考

### 设计系统
- [完整设计规范](../ai-metadata/style/STYLE.md) - ModelCraft 设计系统
- [组件快速参考](../ai-metadata/style/quick-start.md) - 常用组件代码片段

### 原型开发
- [Prototypes README](../prototypes/README.md) - 原型目录说明
- [Layout 原型](../prototypes/layout/) - 布局原型示例
- [Workspace 原型](../prototypes/workspace/) - 工作区原型示例

### 自动化脚本
- [Scripts README](../scripts/README.md) - 自动化脚本说明
- [sync-tailwind.js](../scripts/sync-tailwind.js) - Tailwind 配置同步脚本

## 🎯 快速开始

### 开发新 UI 组件

1. **创建原型**
   ```bash
   # 在 prototypes/ 目录下创建新的 HTML 文件
   code prototypes/my-component/index.html
   ```

2. **同步配置**（如果修改了设计系统）
   ```bash
   npm run sync-tailwind
   ```

3. **预览原型**
   ```bash
   open prototypes/my-component/index.html
   ```

4. **实现 React 组件**
   ```bash
   # 复制原型中的 Tailwind 类名到 React 组件
   code src/components/MyComponent.tsx
   ```

### 修改设计系统

1. **编辑配置**
   ```bash
   # 修改颜色、字体、间距等
   code tailwind.config.ts
   code src/app/globals.css
   ```

2. **同步到原型**
   ```bash
   npm run sync-tailwind
   ```

3. **验证效果**
   ```bash
   # 在浏览器中打开原型验证
   open prototypes/workspace/index.html
   ```

## 📖 详细文档

### Tailwind 配置同步系统

- **概览**: [TAILWIND_SYNC_SUMMARY.md](../TAILWIND_SYNC_SUMMARY.md)
- **完整指南**: [TAILWIND_SYNC_GUIDE.md](TAILWIND_SYNC_GUIDE.md)
- **快速参考**: [QUICK_SYNC_REFERENCE.md](QUICK_SYNC_REFERENCE.md)
- **脚本文档**: [scripts/README.md](../scripts/README.md)

### HTML-First Prototype 工作流

- **Skill 文档**: [.codebuddy/skills/html-first-prototype/SKILL.md](../.codebuddy/skills/html-first-prototype/SKILL.md)
- **设计规范**: [ai-metadata/style/STYLE.md](../ai-metadata/style/STYLE.md)
- **原型指南**: [prototypes/README.md](../prototypes/README.md)

## 🛠️ 常用命令

```bash
# 开发服务器
npm run dev

# 同步 Tailwind 配置
npm run sync-tailwind

# 构建生产版本
npm run build

# 代码检查
npm run lint
```

## 🔗 外部资源

- [Tailwind CSS 文档](https://tailwindcss.com/docs)
- [Next.js 文档](https://nextjs.org/docs)
- [Shadcn/ui 文档](https://ui.shadcn.com/)
- [Lucide Icons](https://lucide.dev/)

## 📋 检查清单

### 开发新页面

- [ ] 创建原型 HTML 文件
- [ ] 在浏览器中预览效果
- [ ] 确认设计符合 ModelCraft 规范
- [ ] 创建 React 页面组件
- [ ] 复制原型中的 Tailwind 类名
- [ ] 测试响应式布局
- [ ] 测试暗色模式（如适用）

### 修改设计系统

- [ ] 编辑 `tailwind.config.ts` 或 `globals.css`
- [ ] 运行 `npm run sync-tailwind`
- [ ] 验证原型中的效果
- [ ] 更新相关 React 组件
- [ ] 更新设计文档
- [ ] 提交包含 `prototypes/shared/*` 的变更

## 💡 提示

- **原型是设计的"唯一真相源"** - 先在原型中确认，再实现 React
- **Tailwind 类名 100% 兼容** - 原型 → React 直接复制
- **自动化减少错误** - 使用 `npm run sync-tailwind` 同步配置
- **不要手动修改生成的文件** - `prototypes/shared/*` 由脚本管理

## 🐛 常见问题

### Q: 原型和 React 显示不一致？

**A:** 运行 `npm run sync-tailwind` 并刷新浏览器缓存。

### Q: 如何添加新颜色？

**A:** 
1. 编辑 `tailwind.config.ts`
2. 运行 `npm run sync-tailwind`
3. 在原型和 React 中使用

### Q: 可以直接修改 `prototypes/shared/` 吗？

**A:** 不要！这些文件由脚本自动生成，修改会被覆盖。

## 📞 获取帮助

遇到问题？查看：

1. [Tailwind 同步指南](TAILWIND_SYNC_GUIDE.md) - 故障排除章节
2. [脚本文档](../scripts/README.md) - 脚本使用说明
3. [HTML-First Prototype Skill](../.codebuddy/skills/html-first-prototype/SKILL.md) - 常见问题

---

**最后更新：** 2026-03-15
