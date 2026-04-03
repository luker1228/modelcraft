# Tailwind 同步快速参考

## 🚀 一行命令

```bash
npm run sync-tailwind
```

## 📋 何时运行

| 修改内容 | 是否需要同步 |
|---------|------------|
| ✏️ 修改原型 `tailwind.config.js` | ✅ **需要** |
| ✏️ 修改原型 `tailwind-base.css` | ✅ **需要** |
| ✏️ 在原型中调整颜色/字体/间距 | ✅ **需要** |
| ✏️ 在原型中修改 CSS 变量 | ✅ **需要** |
| ✏️ 在原型中添加自定义类 | ✅ **需要** |
| 📝 修改 React 组件 | ❌ 不需要 |
| 📝 修改原型 HTML 结构 | ❌ 不需要 |

## ⚡ 快速工作流（原型优先）

```bash
# 1. 在原型中设计 UI
vim prototypes/shared/tailwind.config.js
vim prototypes/shared/tailwind-base.css

# 2. 预览原型效果
open prototypes/workspace/index.html

# 3. 确认设计后同步到 React
npm run sync-tailwind

# 4. 重启开发服务器
npm run dev

# 5. 实现 React 组件（复制原型类名）
# 从原型复制 → 粘贴到 JSX
```

## 🎯 文件映射（设计流向实现）

```
设计源 (原型)                     实现目标 (React)
├── prototypes/shared/            →    tailwind.config.ts
│   └── tailwind.config.js
└── prototypes/shared/            →    src/app/globals.css
    └── tailwind-base.css
```

## ⚠️ 核心原则

- ✅ **DO**: 在原型中设计 → 同步到 React
- ✅ **DO**: 使用 `npm run sync-tailwind` 作为标准工作流
- ❌ **DON'T**: 直接修改 React 的 `tailwind.config.ts` 或 `globals.css`

## 💾 备份系统

同步脚本自动创建备份：

```bash
# 查看备份
ls -la .tailwind-backups/

# 恢复备份（如需要）
cp .tailwind-backups/tailwind.config.ts.TIMESTAMP.backup tailwind.config.ts
```

## 🔗 完整文档

详细信息请查看：[Tailwind 配置同步指南](./TAILWIND_SYNC_GUIDE.md)
