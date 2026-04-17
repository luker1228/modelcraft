# End User Auth 数据库 Schema 终态

## 1. 目标与范围
本文件定义 end-user-auth 的数据库终态，包含：
- 平台库 `mc_meta`（组织/用户/角色/API Key）
- 业务私有库 `private_{projectSlug}`（终端用户认证）

## 2. 新增项总览（相对现网历史版本）
- 新增表：`private_{projectSlug}.users`
- 新增表：`private_{projectSlug}.accounts`
- 新增列：`mc_meta.api_keys.role_ids`（JSON，绑定角色 ID 列表）

## 3. 终态 SQL（全量）

### 3.1 平台库 `mc_meta`

```sql
-- 推荐：先 USE mc_meta;

CREATE TABLE IF NOT EXISTS organizations (
  name VARCHAR(36) NOT NULL PRIMARY KEY,
  display_name VARCHAR(255),
  owner_id VARCHAR(36),
  status VARCHAR(20) NOT NULL DEFAULT 'active',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  INDEX idx_org_owner (owner_id),
  INDEX idx_org_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS users (
  id VARCHAR(36) NOT NULL PRIMARY KEY,
  external_id VARCHAR(255) NULL,
  name VARCHAR(255) NOT NULL,
  phone VARCHAR(32) NOT NULL DEFAULT '',
  password_hash VARCHAR(255) NOT NULL DEFAULT '',
  display_name VARCHAR(255),
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE INDEX uk_phone (phone),
  UNIQUE INDEX uk_user_name (name),
  INDEX idx_external_id (external_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS user_organizations (
  id VARCHAR(36) NOT NULL PRIMARY KEY,
  user_id VARCHAR(36) NOT NULL,
  org_name VARCHAR(36) NOT NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'active',
  invited_by VARCHAR(36),
  invited_at DATETIME(3),
  joined_at DATETIME(3),
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  CONSTRAINT fk_user_org_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_user_org_org FOREIGN KEY (org_name) REFERENCES organizations(name) ON DELETE CASCADE,
  CONSTRAINT fk_user_org_invited_by FOREIGN KEY (invited_by) REFERENCES users(id) ON DELETE SET NULL,
  UNIQUE KEY uk_user_org (user_id, org_name),
  INDEX idx_uo_user_id (user_id),
  INDEX idx_uo_status (status),
  INDEX idx_uo_org_name (org_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS profile (
  id VARCHAR(36) NOT NULL PRIMARY KEY,
  user_id VARCHAR(36) NOT NULL,
  nickname VARCHAR(32) NOT NULL,
  avatar_url VARCHAR(512) NULL,
  bio VARCHAR(256) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  CONSTRAINT fk_profile_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  UNIQUE KEY uk_profile_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

ALTER TABLE organizations
  ADD CONSTRAINT fk_org_owner
  FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE RESTRICT;

CREATE TABLE IF NOT EXISTS roles (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(64) NOT NULL,
  description TEXT,
  is_system BOOLEAN NOT NULL DEFAULT FALSE,
  org_name VARCHAR(36) NOT NULL DEFAULT '__SYSTEM__',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_role_name_org (name, org_name),
  INDEX idx_org_name (org_name),
  INDEX idx_is_system (is_system)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS user_roles (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  user_id VARCHAR(36) NOT NULL,
  role_id BIGINT NOT NULL,
  org_name VARCHAR(36) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  CONSTRAINT fk_user_roles_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_user_roles_role FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
  UNIQUE KEY uk_user_role_org (user_id, role_id, org_name),
  INDEX idx_user_org (user_id, org_name),
  INDEX idx_role_org (role_id, org_name),
  INDEX idx_org_name (org_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS role_permissions (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  role_id BIGINT NOT NULL,
  org_name VARCHAR(36) NOT NULL,
  obj VARCHAR(64) NOT NULL,
  act VARCHAR(64) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  CONSTRAINT fk_role_perms_role FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
  UNIQUE KEY uk_role_obj_act (role_id, obj, act),
  INDEX idx_role_org (role_id, org_name),
  INDEX idx_org_name (org_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS refresh_tokens (
  id VARCHAR(36) NOT NULL,
  user_id VARCHAR(36) NOT NULL,
  token_hash VARCHAR(64) NOT NULL,
  expires_at DATETIME NOT NULL,
  created_at DATETIME NOT NULL,
  revoked_at DATETIME NULL,
  PRIMARY KEY (id),
  INDEX idx_token_hash (token_hash),
  INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS api_keys (
  id VARCHAR(36) NOT NULL,
  user_id VARCHAR(36) NOT NULL,
  name VARCHAR(100) NOT NULL,
  key_hash VARCHAR(64) NOT NULL,
  key_prefix VARCHAR(10) NOT NULL,
  role_ids JSON NULL COMMENT '新增列：绑定角色 ID 列表（roles.id）',
  last_used_at DATETIME NULL,
  expires_at DATETIME NULL,
  created_at DATETIME NOT NULL,
  revoked_at DATETIME NULL,
  PRIMARY KEY (id),
  INDEX idx_key_hash (key_hash),
  INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 3.2 业务私有库 `private_{projectSlug}`

```sql
-- 示例：CREATE DATABASE IF NOT EXISTS private_crm CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
-- 推荐：先 USE private_{projectSlug};

CREATE TABLE IF NOT EXISTS users (
  id VARCHAR(36) NOT NULL PRIMARY KEY COMMENT '新增表',
  username VARCHAR(64) NOT NULL,
  password VARCHAR(255) NOT NULL COMMENT 'bcrypt hash',
  is_forbidden TINYINT(1) NOT NULL DEFAULT 0,
  created_by VARCHAR(36) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uq_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS accounts (
  id VARCHAR(36) NOT NULL PRIMARY KEY COMMENT '新增表',
  user_id VARCHAR(36) NOT NULL,
  refresh_token_hash VARCHAR(255) NOT NULL,
  expires_at DATETIME NOT NULL,
  revoked TINYINT(1) NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_user_id (user_id),
  UNIQUE KEY uq_token_hash (refresh_token_hash),
  CONSTRAINT fk_accounts_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

## 4. 存量库增量迁移 SQL

```sql
-- 场景：历史环境已存在 mc_meta.api_keys，但缺少 role_ids 列
ALTER TABLE api_keys
  ADD COLUMN IF NOT EXISTS role_ids JSON NULL COMMENT '新增列：绑定角色 ID 列表（roles.id）' AFTER key_prefix;

-- 可选回填：将 NULL 统一为 []（便于应用层按数组处理）
UPDATE api_keys
SET role_ids = JSON_ARRAY()
WHERE role_ids IS NULL;
```

## 5. 变更标识规范
- 文档中以“新增表”“新增列”字样显式标识本次新增对象。
- 以本文件 SQL 为 end-user-auth 模块数据库终态基线。
