/**
 * 保留用户名黑名单
 * 与后端 internal/domain/user/username_reserved.go 保持同步
 * 所有值均为小写，校验时需先 toLowerCase()
 */
export const RESERVED_USERNAMES = new Set<string>([
  // 系统身份
  'admin',
  'administrator',
  'root',
  'system',
  'superuser',

  // 平台品牌
  'modelcraft',
  'modelcraft-admin',

  // API 路径关键字
  'api',
  'auth',
  'login',
  'logout',
  'register',
  'refresh',
  'oauth',
  'callback',

  // 资源名词（与 URL 路径冲突）
  'user',
  'users',
  'org',
  'orgs',
  'project',
  'projects',
  'model',
  'models',
  'cluster',
  'clusters',
  'field',
  'fields',
  'group',
  'groups',
  'schema',
  'schemas',
  'dashboard',
  'settings',
  'profile',
  'me',
  'self',

  // 通用保留
  'null',
  'undefined',
  'true',
  'false',
  'none',
  'anonymous',
  'guest',
  'public',
  'private',
  'static',
  'assets',
  'upload',
  'uploads',
  'test',
  'demo',
  'example',
  'sample',
  'support',
  'help',
  'info',
  'about',
  'contact',
  'home',
  'index',

  // 常见攻击向量
  'www',
  'ftp',
  'mail',
  'smtp',
  'pop3',
  'imap',
])

/**
 * 检查用户名是否为保留字（大小写不敏感）
 */
export function isReservedUserName(userName: string): boolean {
  return RESERVED_USERNAMES.has(userName.toLowerCase())
}
