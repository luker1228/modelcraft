# Scripts 目录

此目录包含项目的自动化脚本。

## sync-from-prototype.js

**Tailwind 配置同步脚本** - 将原型的设计系统配置同步到 React 项目。

### 设计理念

**原型优先（Prototype-First）**

- 原型是设计系统的唯一真相源
- 设计流向实现：Prototype → React
- 在原型中设计和验证 UI，然后同步到 React

### 用途

自动同步以下文件：

| 设计源（Prototype） | 实现目标（React） |
|-----------------|-------------------|
| `prototypes/shared/tailwind.config.js` | `tailwind.config.ts` |
| `prototypes/shared/tailwind-base.css` | `src/app/globals.css` |

### 使用方式

```bash
# 运行同步脚本
npm run sync-tailwind
```

### 何时运行

当您在原型中修改以下内容时需要运行此脚本：

1. **修改原型 Tailwind 配置** (`prototypes/shared/tailwind.config.js`)
   - 添加/修改颜色
   - 添加/修改字体
   - 添加/修改间距
   - 添加/修改圆角
   - 添加/修改动画

2. **修改原型 CSS** (`prototypes/shared/tailwind-base.css`)
   - 修改设计系统变量（`--primary`, `--background` 等）
   - 添加/修改自定义类（`.tech-bg`, `.card-interactive` 等）
   - 修改主题颜色

### 工作流程

```
1. 在原型中设计 UI
   ↓
   编辑 prototypes/shared/tailwind.config.js 或 tailwind-base.css
   ↓
2. 在浏览器中预览原型
   ↓
   在 prototypes/ 目录下的 HTML 文件中验证样式
   ↓
3. 确认设计后运行同步
   ↓
   npm run sync-tailwind
   ↓
4. 实现 React 组件
   ↓
   将原型中的 Tailwind 类复制到 React 组件
```

### 脚本输出

成功运行时会看到：

```
╔═══════════════════════════════════════════════════╗
║   Sync From Prototype (Prototype → React)       ║
╚═══════════════════════════════════════════════════╝

🔍 验证文件...
✅ 所有文件存在

📝 同步 Tailwind 配置（原型 → React）...
   备份已创建: tailwind.config.ts.2026-03-15T02-54-33.backup
✅ Tailwind 配置同步成功
   /path/to/prototypes/shared/tailwind.config.js
   → /path/to/tailwind.config.ts

📝 同步 CSS 变量（原型 → React）...
   备份已创建: globals.css.2026-03-15T02-54-33.backup
✅ CSS 变量同步成功
   /path/to/prototypes/shared/tailwind-base.css
   → /path/to/src/app/globals.css

✨ 同步完成！

💡 下一步：
   1. 重启开发服务器以应用新配置
      npm run dev
   2. 在 React 组件中使用原型的类名
   3. 测试样式是否符合预期

⚠️  注意：旧配置已备份到 .tailwind-backups/ 目录
   如需回滚，请手动恢复备份文件
```

### 备份系统

脚本会自动创建备份文件到 `.tailwind-backups/` 目录：

```bash
# 查看备份文件
ls -la .tailwind-backups/

# 恢复备份（如需要）
cp .tailwind-backups/tailwind.config.ts.2026-03-15T02-54-33.backup tailwind.config.ts
cp .tailwind-backups/globals.css.2026-03-15T02-54-33.backup src/app/globals.css
```

### 注意事项

1. **在原型中设计和验证**
   - `prototypes/shared/tailwind.config.js` - 可以编辑
   - `prototypes/shared/tailwind-base.css` - 可以编辑
   
   这些是设计系统的源文件，在这里进行设计工作。

2. **不要直接修改 React 配置**
   - `tailwind.config.ts` - 由脚本生成
   - `src/app/globals.css` - 由脚本生成
   
   除非是 React 特定配置（如插件），否则应该从原型同步。

3. **备份保护**
   - 每次同步前自动创建备份
   - 出错时可以快速恢复
   - 定期清理旧备份文件

### 故障排除

**问题：文件不存在错误**

```
❌ 文件不存在: /path/to/file
```

**解决方案：** 确保原型文件存在：
```bash
ls prototypes/shared/tailwind.config.js
ls prototypes/shared/tailwind-base.css
```

**问题：权限错误**

```
❌ 同步失败: EACCES: permission denied
```

**解决方案：** 
```bash
chmod +x scripts/sync-from-prototype.js
```

**问题：同步后 React 样式没更新**

**解决方案：**
```bash
# 清理 Next.js 缓存
rm -rf .next

# 重启开发服务器
npm run dev
```

### 相关文档

- [Tailwind 配置同步完整指南](../docs/TAILWIND_SYNC_GUIDE.md)
- [快速参考](../docs/QUICK_SYNC_REFERENCE.md)
- [HTML-First Prototype 工作流](../.codebuddy/skills/html-first-prototype/SKILL.md)
- [Prototypes README](../prototypes/README.md)
- [设计系统规范](../ai-metadata/style/STYLE.md)
