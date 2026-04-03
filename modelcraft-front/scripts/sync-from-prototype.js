#!/usr/bin/env node

/**
 * Sync From Prototype Script (Prototype → React)
 * 
 * 从原型文件同步设计系统到 React 项目
 * 
 * 工作流程：
 * 1. 在原型中设计和调整样式
 * 2. 运行此脚本同步到 React 项目
 * 3. React 组件使用同步后的配置
 * 
 * 运行：npm run sync-from-prototype
 */

const fs = require('fs')
const path = require('path')

// 颜色输出
const colors = {
  reset: '\x1b[0m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  red: '\x1b[31m',
  cyan: '\x1b[36m',
}

function log(message, color = 'reset') {
  console.log(`${colors[color]}${message}${colors.reset}`)
}

// 路径定义
const paths = {
  // 源文件（原型）
  prototypesTailwindConfig: path.join(__dirname, '../prototypes/shared/tailwind.config.js'),
  prototypesTailwindBaseCSS: path.join(__dirname, '../prototypes/shared/tailwind-base.css'),
  
  // 目标文件（React 项目）
  reactTailwindConfig: path.join(__dirname, '../tailwind.config.ts'),
  reactGlobalsCSS: path.join(__dirname, '../src/app/globals.css'),
  
  // 备份目录
  backupDir: path.join(__dirname, '../.tailwind-backups'),
}

/**
 * 创建备份
 */
function createBackup(filePath, backupDir) {
  if (!fs.existsSync(backupDir)) {
    fs.mkdirSync(backupDir, { recursive: true })
  }
  
  const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, -5)
  const fileName = path.basename(filePath)
  const backupPath = path.join(backupDir, `${fileName}.${timestamp}.backup`)
  
  fs.copyFileSync(filePath, backupPath)
  return backupPath
}

/**
 * 将 JavaScript tailwind.config.js 转换为 TypeScript tailwind.config.ts
 */
function syncTailwindConfig() {
  log('\n📝 同步 Tailwind 配置（原型 → React）...', 'blue')

  try {
    // 备份现有文件
    const backupPath = createBackup(paths.reactTailwindConfig, paths.backupDir)
    log(`   备份已创建: ${path.basename(backupPath)}`, 'yellow')

    // 读取原型的 tailwind.config.js
    let jsConfig = fs.readFileSync(paths.prototypesTailwindConfig, 'utf-8')

    // 提取 tailwind.config = {...} 中的配置对象
    const configMatch = jsConfig.match(/tailwind\.config\s*=\s*(\{[\s\S]*\})\s*$/m)
    if (!configMatch) {
      throw new Error('无法从原型配置中提取配置对象')
    }

    let configObject = configMatch[1]

    // 转换为 TypeScript 格式
    const tsConfig = `import type { Config } from 'tailwindcss'

const config: Config = ${configObject}

export default config
`

    // 写入 React 项目配置
    fs.writeFileSync(paths.reactTailwindConfig, tsConfig, 'utf-8')

    log('✅ Tailwind 配置同步成功', 'green')
    log(`   ${paths.prototypesTailwindConfig}`, 'reset')
    log(`   → ${paths.reactTailwindConfig}`, 'reset')
  } catch (error) {
    log(`❌ Tailwind 配置同步失败: ${error.message}`, 'red')
    throw error
  }
}

/**
 * 同步 CSS 变量和自定义类
 */
function syncTailwindBaseCSS() {
  log('\n📝 同步 CSS 变量（原型 → React）...', 'blue')

  try {
    // 备份现有文件
    const backupPath = createBackup(paths.reactGlobalsCSS, paths.backupDir)
    log(`   备份已创建: ${path.basename(backupPath)}`, 'yellow')

    // 读取原型的 tailwind-base.css
    let baseCSS = fs.readFileSync(paths.prototypesTailwindBaseCSS, 'utf-8')

    // 生成 globals.css
    const globalsCSS = `@tailwind base;
@tailwind components;
@tailwind utilities;

${baseCSS}
`

    // 写入 React 项目 globals.css
    fs.writeFileSync(paths.reactGlobalsCSS, globalsCSS, 'utf-8')

    log('✅ CSS 变量同步成功', 'green')
    log(`   ${paths.prototypesTailwindBaseCSS}`, 'reset')
    log(`   → ${paths.reactGlobalsCSS}`, 'reset')
  } catch (error) {
    log(`❌ CSS 变量同步失败: ${error.message}`, 'red')
    throw error
  }
}

/**
 * 验证文件是否存在
 */
function validateFiles() {
  log('\n🔍 验证文件...', 'blue')

  const missingFiles = []

  // 检查源文件（原型）
  if (!fs.existsSync(paths.prototypesTailwindConfig)) {
    missingFiles.push(paths.prototypesTailwindConfig)
  }
  if (!fs.existsSync(paths.prototypesTailwindBaseCSS)) {
    missingFiles.push(paths.prototypesTailwindBaseCSS)
  }

  if (missingFiles.length > 0) {
    log('❌ 缺少原型文件:', 'red')
    missingFiles.forEach(file => log(`   - ${file}`, 'red'))
    throw new Error('请先创建原型文件')
  }

  log('✅ 所有文件存在', 'green')
}

/**
 * 显示变更摘要
 */
function showSummary() {
  log('\n📊 变更摘要:', 'cyan')
  log('   ┌─────────────────────────────────────────────────┐', 'cyan')
  log('   │  原型设计 → React 项目                          │', 'cyan')
  log('   ├─────────────────────────────────────────────────┤', 'cyan')
  log('   │  ✓ Tailwind 配置已更新                          │', 'cyan')
  log('   │  ✓ CSS 变量已更新                               │', 'cyan')
  log('   │  ✓ 旧文件已备份到 .tailwind-backups/           │', 'cyan')
  log('   └─────────────────────────────────────────────────┘', 'cyan')
}

/**
 * 主函数
 */
async function main() {
  try {
    log('╔═══════════════════════════════════════════════════╗', 'blue')
    log('║   Sync From Prototype (Prototype → React)       ║', 'blue')
    log('╚═══════════════════════════════════════════════════╝', 'blue')

    // 验证文件
    validateFiles()

    // 同步配置
    syncTailwindConfig()

    // 同步 CSS
    syncTailwindBaseCSS()

    // 显示摘要
    showSummary()

    log('\n✨ 同步完成！', 'green')
    log('\n💡 下一步：', 'yellow')
    log('   1. 重启开发服务器以应用新配置', 'yellow')
    log('      npm run dev', 'yellow')
    log('   2. 在 React 组件中使用原型的类名', 'yellow')
    log('   3. 测试样式是否符合预期\n', 'yellow')

    log('⚠️  注意：旧配置已备份到 .tailwind-backups/ 目录', 'yellow')
    log('   如需回滚，请手动恢复备份文件\n', 'yellow')

    process.exit(0)
  } catch (error) {
    log(`\n❌ 同步失败: ${error.message}`, 'red')
    log('\n💡 故障排除：', 'yellow')
    log('   - 确保原型文件存在且格式正确', 'yellow')
    log('   - 检查文件权限', 'yellow')
    log('   - 查看备份文件以恢复', 'yellow')
    process.exit(1)
  }
}

// 运行
main()
