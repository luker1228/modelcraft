-- Create "models" table
CREATE TABLE `models` (
  `id` varchar(36) NOT NULL COMMENT "模型唯一标识符",
  `org_name` varchar(36) NOT NULL COMMENT "所属组织名称（来自projects表复合主键）",
  `project_slug` varchar(64) NOT NULL COMMENT "所属项目标识符（来自projects表复合主键）",
  `name` varchar(64) NOT NULL COMMENT "模型名称（唯一标识）",
  `title` varchar(255) NOT NULL COMMENT "模型显示标题",
  `description` text NULL COMMENT "模型描述信息",
  `storage_type` varchar(100) NOT NULL COMMENT "存储类型",
  `database_name` varchar(64) NOT NULL COMMENT "数据库名称",
  `display_field` varchar(64) NULL COMMENT "用于 runtime _displayName 解析的字段名（必须是模型中存在且可字符串化的字段）",
  `insertion_order_field` varchar(64) NULL COMMENT "用于 listPage cursor 分页的插入序字段名（单调递增字段，如 created_at）",
  `version` bigint NULL DEFAULT 1 COMMENT "数据版本号",
  `status` varchar(50) NULL DEFAULT "draft" COMMENT "模型状态：draft/published/archived",
  `group_id` varchar(36) NULL COMMENT "所属分组ID，NULL表示未分组（ungrouped）",
  `deployment_status` varchar(50) NULL DEFAULT "pending" COMMENT "部署状态：pending/success/failed",
  `last_sync_at` datetime(3) NULL COMMENT "最后同步时间",
  `sync_error` text NULL COMMENT "同步错误信息",
  `created_via` enum('NEW','IMPORTED') NOT NULL DEFAULT "NEW" COMMENT "模型创建来源：NEW=新建，IMPORTED=导入",
  `is_read_only` bool NOT NULL DEFAULT 0 COMMENT "是否只读：1=只读（禁止结构修改），0=可编辑",
  `created_at` datetime(3) NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT "创建时间",
  `updated_at` datetime(3) NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT "更新时间",
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT "软删除时间戳，0 表示活跃",
  `delete_token` bigint unsigned NOT NULL DEFAULT 0 COMMENT "唯一键避让位，0 表示活跃",
  PRIMARY KEY (`id`),
  INDEX `idx_models_created_via` (`created_via`) COMMENT "模型创建来源查询索引",
  INDEX `idx_models_is_read_only` (`is_read_only`) COMMENT "只读模型查询索引",
  INDEX `idx_models_live_project` (`org_name`, `project_slug`, `deleted_at`) COMMENT "项目活跃模型查询索引",
  UNIQUE INDEX `idx_models_name` (`org_name`, `project_slug`, `database_name`, `name`, `delete_token`) COMMENT "组织+项目内模型名称唯一索引",
  INDEX `idx_models_project` (`org_name`, `project_slug`) COMMENT "项目查询索引"
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT "模型定义主表";
-- Create "user_api_tokens" table
CREATE TABLE `user_api_tokens` (
  `id` varchar(36) NOT NULL COMMENT "唯一标识符 (UUID v7)",
  `org_name` varchar(255) NOT NULL COMMENT "所属组织",
  `end_user_id` varchar(36) NOT NULL COMMENT "创建者 EndUser ID",
  `name` varchar(255) NOT NULL COMMENT "用户自定义名称",
  `token_hash` varchar(64) NOT NULL COMMENT "SHA-256(plaintext) hex，用于验证",
  `expires_at` datetime NULL COMMENT "NULL 表示永不过期",
  `last_used_at` datetime NULL COMMENT "最近使用时间，异步更新",
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT "创建时间",
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT "软删除时间戳，0 表示活跃",
  `delete_token` bigint unsigned NOT NULL DEFAULT 0 COMMENT "唯一键避让位，0 表示活跃",
  PRIMARY KEY (`id`),
  INDEX `idx_user_tokens` (`org_name`, `end_user_id`, `deleted_at`),
  UNIQUE INDEX `uq_token_hash` (`token_hash`),
  UNIQUE INDEX `uq_user_token_name` (`org_name`, `end_user_id`, `name`, `delete_token`)
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT "EndUser Personal Access Token 注册表（PAT）";
-- Create "project_auth_schemas" table
CREATE TABLE `project_auth_schemas` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `org_name` varchar(36) NOT NULL COMMENT "所属组织名称",
  `project_slug` varchar(64) NOT NULL COMMENT "所属项目标识符",
  `variables` json NOT NULL COMMENT "认证变量列表 [{name, source, type}]",
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT "创建时间",
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT "更新时间",
  PRIMARY KEY (`id`),
  INDEX `idx_auth_schemas_project` (`org_name`, `project_slug`) COMMENT "项目查询索引",
  UNIQUE INDEX `uk_auth_schemas_project` (`org_name`, `project_slug`)
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT "Project 认证变量配置";
-- Create "model_database" table
CREATE TABLE `model_database` (
  `id` varchar(36) NOT NULL COMMENT "数据库唯一标识符 (UUID)",
  `org_name` varchar(64) NOT NULL COMMENT "所属组织名称",
  `project_slug` varchar(64) NOT NULL COMMENT "所属项目标识符",
  `cluster_id` varchar(36) NOT NULL COMMENT "所属数据库集群 ID",
  `name` varchar(64) NOT NULL COMMENT "MySQL database 原始名，来自 SHOW DATABASES，注册后不可修改",
  `title` varchar(128) NOT NULL COMMENT "用户设置的友好名称，默认等于 name",
  `description` text NULL COMMENT "可选描述",
  `mode` enum('self_hosted','managed') NOT NULL COMMENT "self_hosted=可读写; managed=只读",
  `latest_sync_job_id` varchar(36) NULL COMMENT "最近一次 ModelSyncJob ID",
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT "软删除时间戳，0 表示活跃",
  `delete_token` bigint unsigned NOT NULL DEFAULT 0 COMMENT "唯一键避让位，0 表示活跃",
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  INDEX `idx_model_database_cluster` (`cluster_id`),
  INDEX `idx_model_database_live_project` (`org_name`, `project_slug`, `deleted_at`),
  INDEX `idx_model_database_project` (`org_name`, `project_slug`),
  UNIQUE INDEX `uq_project_database` (`project_slug`, `name`, `delete_token`)
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT "Project 已接管的 MySQL database 注册表";
-- Create "model_database_sync_job" table
CREATE TABLE `model_database_sync_job` (
  `id` varchar(36) NOT NULL COMMENT "同步任务唯一标识符 (UUID)",
  `org_name` varchar(64) NOT NULL COMMENT "所属组织名称",
  `project_slug` varchar(64) NOT NULL COMMENT "所属项目标识符",
  `database_id` varchar(36) NOT NULL COMMENT "关联的已接管数据库 ID",
  `status` enum('pending','running','succeeded','partial_success','failed') NOT NULL COMMENT "任务状态",
  `total_tables` int NOT NULL DEFAULT 0 COMMENT "扫描到的表总数",
  `processed_tables` int NOT NULL DEFAULT 0 COMMENT "已处理表数",
  `created_models` int NOT NULL DEFAULT 0 COMMENT "新导入模型数",
  `synced_models` int NOT NULL DEFAULT 0 COMMENT "已同步 schema 的模型数",
  `failed_count` int NOT NULL DEFAULT 0 COMMENT "失败表数",
  `failed_tables` json NOT NULL COMMENT "失败表详情 [{tableName,message}]",
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT "软删除时间戳，0 表示活跃",
  `delete_token` bigint unsigned NOT NULL DEFAULT 0 COMMENT "唯一键避让位，0 表示活跃",
  `started_at` datetime(3) NULL COMMENT "任务开始时间",
  `finished_at` datetime(3) NULL COMMENT "任务结束时间",
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  INDEX `idx_model_database_sync_job_live_project` (`org_name`, `project_slug`, `deleted_at`),
  INDEX `idx_model_database_sync_job_project_db_created` (`org_name`, `project_slug`, `database_id`, `created_at`),
  INDEX `idx_model_database_sync_job_project_status` (`org_name`, `project_slug`, `status`)
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT "数据库同步为模型的异步任务表";
-- Create "model_enums" table
CREATE TABLE `model_enums` (
  `id` varchar(36) NOT NULL COMMENT "枚举唯一标识符",
  `org_name` varchar(36) NOT NULL COMMENT "所属组织名称（来自projects表复合主键）",
  `project_slug` varchar(64) NOT NULL COMMENT "所属项目标识符（来自projects表复合主键）",
  `name` varchar(64) NOT NULL COMMENT "枚举名称（唯一标识）",
  `display_name` varchar(255) NOT NULL COMMENT "枚举显示名称",
  `description` text NULL COMMENT "枚举描述信息",
  `options` json NOT NULL COMMENT "枚举选项配置（JSON数组格式）",
  `is_multi_select` bool NULL DEFAULT 0 COMMENT "是否支持多选",
  `created_at` datetime(3) NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT "创建时间",
  `updated_at` datetime(3) NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT "更新时间",
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT "软删除时间戳，0 表示活跃",
  `delete_token` bigint unsigned NOT NULL DEFAULT 0 COMMENT "唯一键避让位，0 表示活跃",
  PRIMARY KEY (`id`),
  INDEX `idx_model_enums_live_project` (`org_name`, `project_slug`, `deleted_at`) COMMENT "项目活跃枚举查询索引",
  UNIQUE INDEX `idx_model_enums_name` (`org_name`, `project_slug`, `name`, `delete_token`) COMMENT "组织+项目内枚举名称唯一索引",
  INDEX `idx_model_enums_project` (`org_name`, `project_slug`) COMMENT "项目查询索引"
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT "模型枚举定义表";
-- Create "database_clusters" table
CREATE TABLE `database_clusters` (
  `id` varchar(36) NOT NULL COMMENT "集群唯一标识符",
  `org_name` varchar(36) NOT NULL DEFAULT "" COMMENT "所属组织名称（来自projects表复合主键）",
  `project_slug` varchar(64) NOT NULL DEFAULT "" COMMENT "所属项目标识符（来自projects表复合主键）",
  `title` varchar(255) NOT NULL COMMENT "集群显示标题",
  `description` text NULL COMMENT "集群描述信息",
  `host` varchar(255) NOT NULL COMMENT "数据库主机地址",
  `port` bigint NOT NULL DEFAULT 3306 COMMENT "数据库端口号",
  `username` varchar(255) NOT NULL COMMENT "数据库用户名",
  `password` text NOT NULL COMMENT "数据库密码",
  `connection_timeout` int NOT NULL DEFAULT 5 COMMENT "连接超时(秒,5-15)",
  `charset` varchar(50) NULL DEFAULT "utf8mb4" COMMENT "数据库字符集",
  `max_open_conns` bigint NULL DEFAULT 100 COMMENT "最大打开连接数",
  `max_idle_conns` bigint NULL DEFAULT 10 COMMENT "最大空闲连接数",
  `conn_max_lifetime` bigint NULL DEFAULT 3600 COMMENT "连接最大生命周期（秒）",
  `status` varchar(20) NULL DEFAULT "active" COMMENT "集群状态:active/inactive",
  `version` bigint NULL DEFAULT 1 COMMENT "数据版本号",
  `created_at` datetime(3) NULL COMMENT "创建时间",
  `updated_at` datetime(3) NULL COMMENT "更新时间",
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT "软删除时间戳，0 表示活跃",
  `delete_token` bigint unsigned NOT NULL DEFAULT 0 COMMENT "唯一键避让位，0 表示活跃",
  PRIMARY KEY (`id`),
  INDEX `idx_cluster_live_project` (`org_name`, `project_slug`, `deleted_at`) COMMENT "项目活跃集群查询索引",
  INDEX `idx_cluster_project` (`org_name`, `project_slug`) COMMENT "项目查询索引",
  UNIQUE INDEX `idx_cluster_project_unique` (`org_name`, `project_slug`, `delete_token`) COMMENT "一对一约束：一个项目只能有一个活跃集群",
  INDEX `idx_status` (`status`) COMMENT "状态查询索引"
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT "数据库集群配置信息表";
-- Create "projects" table
CREATE TABLE `projects` (
  `org_name` varchar(36) NOT NULL COMMENT "组织名称（复合主键之一）",
  `slug` varchar(64) NOT NULL COMMENT "项目标识符（组织内唯一，复合主键之二）",
  `title` varchar(255) NOT NULL COMMENT "项目显示标题",
  `description` text NULL COMMENT "项目描述信息",
  `status` varchar(20) NOT NULL DEFAULT "active" COMMENT "项目状态：active/archived（永不物理删除）",
  `cluster_id` varchar(36) NULL COMMENT "关联的集群ID（一对一关系）",
  `created_at` datetime(3) NULL COMMENT "创建时间",
  `updated_at` datetime(3) NULL COMMENT "更新时间",
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT "软删除时间戳，0 表示活跃",
  `delete_token` bigint unsigned NOT NULL DEFAULT 0 COMMENT "唯一键避让位，0 表示活跃",
  PRIMARY KEY (`org_name`, `slug`, `delete_token`),
  INDEX `idx_org_name` (`org_name`) COMMENT "组织查询索引",
  INDEX `idx_project_cluster` (`cluster_id`) COMMENT "集群查询索引",
  INDEX `idx_project_live_org` (`org_name`, `deleted_at`) COMMENT "组织活跃项目查询索引",
  INDEX `idx_project_slug` (`slug`) COMMENT "项目标识符查询索引",
  INDEX `idx_project_status` (`status`) COMMENT "状态查询索引",
  INDEX `idx_projects_org_slug` (`org_name`, `slug`) COMMENT "活跃项目查询索引"
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT "项目信息表（多租户，复合主键，只归档不删除）";
-- Create "security_audit_logs" table
CREATE TABLE `security_audit_logs` (
  `id` varchar(36) NOT NULL,
  `user_id` varchar(36) NOT NULL,
  `event` varchar(50) NOT NULL COMMENT "如 REUSE_DETECTED",
  `detail` json NULL COMMENT "token_id, ip 等上下文",
  `created_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_user_id_created` (`user_id`, `created_at`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "model_sync_job" table
CREATE TABLE `model_sync_job` (
  `id` varchar(36) NOT NULL COMMENT "任务唯一标识符",
  `batch_id` varchar(36) NOT NULL DEFAULT "" COMMENT "批次 ID，同批次多条 job 共享",
  `database_id` varchar(36) NOT NULL DEFAULT "" COMMENT "关联 model_database.id",
  `org_name` varchar(64) NOT NULL COMMENT "所属组织名称",
  `project_slug` varchar(64) NOT NULL COMMENT "所属项目标识符",
  `database_name` varchar(128) NOT NULL COMMENT "目标数据库名称",
  `table_names` json NOT NULL COMMENT "指定同步的表名列表，空数组表示全量",
  `status` enum('pending','running','succeeded','partial_success','failed') NOT NULL COMMENT "任务状态",
  `total_tables` int NOT NULL DEFAULT 0 COMMENT "扫描到的总表数",
  `processed_tables` int NOT NULL DEFAULT 0 COMMENT "已处理表数",
  `created_models` int NOT NULL DEFAULT 0 COMMENT "新建模型数",
  `synced_models` int NOT NULL DEFAULT 0 COMMENT "已同步模型数",
  `failed_count` int NOT NULL DEFAULT 0 COMMENT "失败表数",
  `failed_tables` json NOT NULL COMMENT "失败明细，格式：[{tableName, message}]",
  `started_at` datetime(3) NULL COMMENT "worker 开始时间",
  `finished_at` datetime(3) NULL COMMENT "任务结束时间",
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT "创建时间",
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT "更新时间",
  PRIMARY KEY (`id`),
  INDEX `idx_model_sync_job_batch` (`org_name`, `project_slug`, `batch_id`),
  INDEX `idx_model_sync_job_database_id` (`org_name`, `project_slug`, `database_id`),
  INDEX `idx_model_sync_job_project_db` (`org_name`, `project_slug`, `database_name`, `created_at`),
  INDEX `idx_model_sync_job_status` (`org_name`, `project_slug`, `status`)
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT "syncModelsFromDB 异步任务表";
-- Create "project_auth_configs" table
CREATE TABLE `project_auth_configs` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT "Primary key",
  `org_name` varchar(36) NOT NULL COMMENT "所属组织名称（来自projects表复合主键）",
  `project_slug` varchar(64) NOT NULL COMMENT "所属项目标识符（来自projects表复合主键）",
  `provider` varchar(32) NOT NULL COMMENT "Authentication provider type: auth_provider, keycloak, oidc",
  `enabled` bool NOT NULL DEFAULT 1 COMMENT "Whether authentication is enabled for this project",
  `config` json NOT NULL COMMENT "Provider-specific configuration (endpoint, client_id, certificate, etc.)",
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT "Record creation timestamp",
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT "Record last update timestamp",
  PRIMARY KEY (`id`),
  INDEX `idx_project_auth_configs_project` (`org_name`, `project_slug`),
  INDEX `idx_project_auth_configs_provider` (`provider`),
  UNIQUE INDEX `idx_project_auth_unique` (`org_name`, `project_slug`),
  CONSTRAINT `chk_provider_type` CHECK (`provider` in (_utf8mb4'auth_provider',_utf8mb4'keycloak',_utf8mb4'oidc'))
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT "Project-level authentication configuration";
-- Create "refresh_tokens" table
CREATE TABLE `refresh_tokens` (
  `id` varchar(36) NOT NULL,
  `user_id` varchar(36) NOT NULL,
  `token_hash` varchar(64) NOT NULL COMMENT "SHA256 hash，不存明文",
  `expires_at` datetime NOT NULL,
  `created_at` datetime NOT NULL,
  `revoked_at` datetime NULL COMMENT "NULL = 有效",
  PRIMARY KEY (`id`),
  INDEX `idx_token_hash` (`token_hash`),
  INDEX `idx_user_id` (`user_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "model_groups" table
CREATE TABLE `model_groups` (
  `id` varchar(36) NOT NULL COMMENT "分组唯一标识符",
  `org_name` varchar(36) NOT NULL COMMENT "所属组织名称（来自projects表复合主键）",
  `project_slug` varchar(64) NOT NULL COMMENT "所属项目标识符（来自projects表复合主键）",
  `name` varchar(64) NOT NULL COMMENT "分组名称，仅小写字母、数字和下划线，字母开头",
  `display_order` varchar(255) NOT NULL COMMENT "字典序排序键，用于拖拽排序（lexicographic fractional index）",
  `created_at` datetime(3) NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT "创建时间",
  `updated_at` datetime(3) NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT "更新时间",
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT "软删除时间戳，0 表示活跃",
  `delete_token` bigint unsigned NOT NULL DEFAULT 0 COMMENT "唯一键避让位，0 表示活跃",
  PRIMARY KEY (`id`),
  INDEX `idx_model_groups_live_project` (`org_name`, `project_slug`, `deleted_at`) COMMENT "项目活跃分组查询索引",
  UNIQUE INDEX `idx_model_groups_name` (`org_name`, `project_slug`, `name`, `delete_token`) COMMENT "组织+项目内分组名称唯一索引",
  INDEX `idx_model_groups_project_order` (`org_name`, `project_slug`, `display_order`) COMMENT "项目分组排序查询索引"
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT "模型分组定义表";
-- Create "field_definitions" table
CREATE TABLE `field_definitions` (
  `model_id` varchar(36) NOT NULL COMMENT "所属模型ID",
  `name` varchar(64) NOT NULL COMMENT "字段名称",
  `org_name` varchar(36) NOT NULL COMMENT "所属组织名称（多租户隔离）",
  `project_slug` varchar(64) NOT NULL COMMENT "项目标识符（多租户隔离）",
  `model_name` varchar(64) NOT NULL COMMENT "模型名称",
  `database_name` varchar(64) NOT NULL COMMENT "数据库名称",
  `enum_name` varchar(64) NULL COMMENT "关联的枚举名称（format=ENUM）",
  `belongs_to_fk_id` varchar(36) NULL COMMENT "所属逻辑外键ID（FK列字段使用）",
  `relate_fk_id` varchar(36) NULL COMMENT "关联的逻辑外键ID（RELATION格式字段使用）",
  `title` varchar(255) NOT NULL COMMENT "字段显示标题",
  `description` text NULL COMMENT "字段描述信息",
  `format` varchar(50) NOT NULL COMMENT "字段格式类型",
  `non_null` bool NULL DEFAULT 0 COMMENT "是否可为null",
  `required` bool NULL DEFAULT 0 COMMENT "是否创建时必填",
  `is_unique` bool NULL DEFAULT 0 COMMENT "是否唯一",
  `is_primary` bool NULL DEFAULT 0 COMMENT "是否主键",
  `is_deprecated` bool NULL DEFAULT 0 COMMENT "是否已废弃",
  `storage_hint` varchar(128) NULL COMMENT "存储优化提示，通常为 DB 列名；非空表示该字段映射到实际 DB 列，参与 syncModelsFromDB 的 full sync",
  `status` varchar(20) NOT NULL DEFAULT "init" COMMENT "字段状态：init/active/inactive",
  `validation` json NULL COMMENT "字段验证规则配置",
  `display_order` varchar(32) NOT NULL DEFAULT "" COMMENT "字典序排序键，用于拖拽排序（lexicographic fractional index）",
  `metadata` json NULL COMMENT "字段元数据配置",
  `created_at` datetime(3) NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT "创建时间",
  `updated_at` datetime(3) NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT "更新时间",
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT "软删除时间戳，0 表示活跃",
  `delete_token` bigint unsigned NOT NULL DEFAULT 0 COMMENT "唯一键避让位，0 表示活跃",
  PRIMARY KEY (`model_id`, `name`, `delete_token`),
  INDEX `idx_field_definitions_live_project` (`org_name`, `project_slug`, `deleted_at`) COMMENT "项目活跃字段查询索引",
  INDEX `idx_field_definitions_model` (`org_name`, `project_slug`, `model_name`) COMMENT "模型查询索引",
  INDEX `idx_field_definitions_project` (`org_name`, `project_slug`) COMMENT "项目查询索引",
  CONSTRAINT `fk_field_definitions_model` FOREIGN KEY (`model_id`) REFERENCES `models` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT "模型字段定义表";
-- Create "logical_foreign_keys" table
CREATE TABLE `logical_foreign_keys` (
  `id` varchar(36) NOT NULL COMMENT "逻辑外键唯一标识符",
  `pair_id` varchar(36) NOT NULL COMMENT "外键对ID，normal 和 reverse 行共享",
  `org_name` varchar(36) NOT NULL COMMENT "所属组织名称",
  `direction` enum('normal','reverse') NOT NULL COMMENT "外键方向：normal（正向）/ reverse（反向）",
  `model_id` varchar(36) NOT NULL COMMENT "所属模型ID",
  `model_name` varchar(255) NOT NULL COMMENT "所属模型名称（冗余存储，model_name 不变）",
  `ref_model_id` varchar(36) NULL COMMENT "引用模型ID（外部表场景可为空）",
  `ref_model_name` varchar(255) NOT NULL COMMENT "引用模型名称（冗余存储，model_name 不变）",
  `ref_database_name` varchar(64) NULL COMMENT "引用数据库名（外部表场景）",
  `ref_table_name` varchar(64) NULL COMMENT "引用表名（外部表场景）",
  `source_fields` json NOT NULL COMMENT "源字段列表（JSON数组）",
  `target_fields` json NOT NULL COMMENT "目标字段列表（JSON数组）",
  `is_deletable` bool NOT NULL DEFAULT 1 COMMENT "是否允许删除（系统FK为0）",
  `created_at` datetime(3) NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT "创建时间",
  `updated_at` datetime(3) NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT "更新时间",
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT "软删除时间戳，0 表示活跃",
  `delete_token` bigint unsigned NOT NULL DEFAULT 0 COMMENT "唯一键避让位，0 表示活跃",
  PRIMARY KEY (`id`),
  INDEX `idx_logical_fk_model_live` (`model_id`, `org_name`, `deleted_at`) COMMENT "模型活跃逻辑外键查询索引",
  INDEX `idx_logical_fk_pair_live` (`pair_id`, `org_name`, `deleted_at`) COMMENT "外键对活跃记录查询索引",
  INDEX `idx_logical_fk_ref_model_live` (`ref_model_id`, `org_name`, `deleted_at`) COMMENT "引用模型活跃逻辑外键查询索引",
  CONSTRAINT `fk_logical_fk_model` FOREIGN KEY (`model_id`) REFERENCES `models` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `fk_logical_fk_ref_model` FOREIGN KEY (`ref_model_id`) REFERENCES `models` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT "逻辑外键定义表";
-- Create "model_field_enum_associations" table
CREATE TABLE `model_field_enum_associations` (
  `model_id` varchar(36) NOT NULL COMMENT "模型ID",
  `field_name` varchar(64) NOT NULL COMMENT "字段名称",
  `org_name` varchar(36) NOT NULL COMMENT "所属组织名称",
  `project_slug` varchar(64) NOT NULL COMMENT "所属项目标识符",
  `enum_name` varchar(64) NOT NULL COMMENT "关联的枚举名称",
  `database_name` varchar(64) NOT NULL COMMENT "数据库名称",
  `created_at` datetime(3) NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT "创建时间",
  `updated_at` datetime(3) NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT "更新时间",
  PRIMARY KEY (`model_id`, `field_name`),
  INDEX `idx_field_enum_assoc_cluster_db` (`database_name`) COMMENT "数据库查询索引",
  INDEX `idx_field_enum_assoc_enum_name` (`org_name`, `project_slug`, `enum_name`) COMMENT "组织+项目内枚举名称查询索引",
  INDEX `idx_field_enum_assoc_project` (`org_name`, `project_slug`) COMMENT "项目查询索引",
  CONSTRAINT `fk_field_enum_assoc_model` FOREIGN KEY (`model_id`) REFERENCES `models` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT "模型字段与枚举关联表";
-- Create "model_rls_policies" table
CREATE TABLE `model_rls_policies` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `org_name` varchar(36) NOT NULL COMMENT "组织名",
  `project_slug` varchar(64) NOT NULL COMMENT "项目标识",
  `model_id` varchar(36) NOT NULL COMMENT "模型 ID",
  `policy_name` varchar(255) NOT NULL COMMENT "策略名称（model 内唯一）",
  `action` enum('read','create','update','delete') NOT NULL COMMENT "操作类型",
  `role` varchar(255) NOT NULL DEFAULT "*" COMMENT "匹配角色（*=通配，匹配所有 EndUser）",
  `using_expr` text NULL COMMENT "USING 表达式（read/update/delete）",
  `with_check_expr` text NULL COMMENT "WITH CHECK 表达式（create/update）",
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  INDEX `fk_rls_policies_model` (`model_id`),
  INDEX `idx_policy_name` (`org_name`, `project_slug`, `model_id`, `policy_name`),
  UNIQUE INDEX `uk_policy_role_action` (`org_name`, `project_slug`, `model_id`, `role`, `action`),
  CONSTRAINT `fk_rls_policies_model` FOREIGN KEY (`model_id`) REFERENCES `models` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT "RLS 策略表（多策略存储）";
-- Create "organizations" table
CREATE TABLE `organizations` (
  `name` varchar(36) NOT NULL COMMENT "唯一标识符，随机slug",
  `display_name` varchar(255) NULL COMMENT "用于 UI 显示的名称",
  `owner_id` varchar(36) NULL COMMENT "组织创建者/所有者（引用 users.id）",
  `phone` varchar(32) NOT NULL DEFAULT "" COMMENT "Org 注册手机号，全局唯一",
  `status` varchar(20) NOT NULL DEFAULT "active" COMMENT "状态：active、suspended、deleted",
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT "创建时间",
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT "更新时间",
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT "软删除时间戳，0 表示活跃",
  `delete_token` bigint unsigned NOT NULL DEFAULT 0 COMMENT "唯一键避让位，0 表示活跃",
  PRIMARY KEY (`name`),
  INDEX `idx_org_live_status` (`deleted_at`, `status`) COMMENT "按活跃状态筛选组织",
  INDEX `idx_org_owner` (`owner_id`) COMMENT "按所有者查找组织",
  INDEX `idx_org_status` (`status`) COMMENT "按状态筛选",
  UNIQUE INDEX `uk_org_phone` (`phone`, `delete_token`) COMMENT "Org 手机号全局唯一"
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT "组织表（多租户容器）";
-- Create "users" table
CREATE TABLE `users` (
  `id` varchar(36) NOT NULL COMMENT "内部 UUID",
  `external_id` varchar(255) NULL COMMENT "外部认证提供者用户 ID（来自 JWT.sub，AuthProvider 用户有值，本地注册用户为 NULL）",
  `name` varchar(255) NOT NULL COMMENT "用户名（userName）",
  `phone` varchar(32) NOT NULL DEFAULT "" COMMENT "用户手机号",
  `password_hash` varchar(255) NOT NULL DEFAULT "" COMMENT "bcrypt 密码哈希（本地注册用户有值，AuthProvider 用户为空）",
  `display_name` varchar(255) NULL COMMENT "用于 UI 显示的名称",
  `org_name` varchar(36) NOT NULL DEFAULT "" COMMENT "所属 Org，创建时绑定（引用 organizations.name）",
  `is_admin` bool NOT NULL DEFAULT 0 COMMENT "是否为管理员",
  `status` varchar(20) NOT NULL DEFAULT "active" COMMENT "状态：active | suspended",
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT "创建时间",
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT "更新时间",
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT "软删除时间戳，0 表示活跃",
  `delete_token` bigint unsigned NOT NULL DEFAULT 0 COMMENT "唯一键避让位，0 表示活跃",
  PRIMARY KEY (`id`),
  INDEX `idx_external_id` (`external_id`) COMMENT "按外部 ID 快速查找",
  INDEX `idx_users_live_name` (`deleted_at`, `org_name`, `name`) COMMENT "活跃用户查询索引",
  UNIQUE INDEX `uk_org_user_name` (`org_name`, `name`, `delete_token`) COMMENT "Org 内用户名唯一",
  UNIQUE INDEX `uk_org_user_phone` (`org_name`, `phone`, `delete_token`) COMMENT "Org 内手机号唯一",
  CONSTRAINT `fk_users_org` FOREIGN KEY (`org_name`) REFERENCES `organizations` (`name`) ON UPDATE CASCADE ON DELETE NO ACTION
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT "用户表";
-- Create "profile" table
CREATE TABLE `profile` (
  `id` varchar(36) NOT NULL COMMENT "profile UUID",
  `user_id` varchar(36) NOT NULL COMMENT "用户 ID（引用 users.id）",
  `nickname` varchar(32) NOT NULL COMMENT "昵称",
  `avatar_url` varchar(512) NULL COMMENT "头像 URL（当前可为空，上传能力后续实现）",
  `bio` varchar(256) NULL COMMENT "个人简介",
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT "创建时间",
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT "更新时间",
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT "软删除时间戳，0 表示活跃",
  `delete_token` bigint unsigned NOT NULL DEFAULT 0 COMMENT "唯一键避让位，0 表示活跃",
  PRIMARY KEY (`id`),
  INDEX `idx_profile_live_user` (`deleted_at`, `user_id`) COMMENT "活跃 profile 查询索引",
  UNIQUE INDEX `uk_profile_user_id` (`user_id`, `delete_token`) COMMENT "保证一个用户仅有一个活跃 profile",
  CONSTRAINT `fk_profile_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT "用户资料表";
-- Create "roles" table
CREATE TABLE `roles` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT "角色 ID（主键）",
  `name` varchar(64) NOT NULL COMMENT "角色名称（如 owner, admin, editor, viewer, 或自定义）",
  `description` text NULL COMMENT "角色描述",
  `is_system` bool NOT NULL DEFAULT 0 COMMENT "系统角色标识（true = 不可修改删除）",
  `org_name` varchar(36) NOT NULL DEFAULT "__SYSTEM__" COMMENT "所属组织名称（__SYSTEM__ = 系统角色，其他 = 租户自定义角色）",
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT "创建时间",
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT "更新时间",
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT "软删除时间戳，0 表示活跃",
  `delete_token` bigint unsigned NOT NULL DEFAULT 0 COMMENT "唯一键避让位，0 表示活跃",
  PRIMARY KEY (`id`),
  INDEX `idx_is_system` (`is_system`) COMMENT "筛选系统角色",
  INDEX `idx_org_name` (`org_name`) COMMENT "按组织查询角色",
  INDEX `idx_roles_live_org` (`org_name`, `deleted_at`) COMMENT "组织活跃角色查询索引",
  UNIQUE INDEX `uk_role_name_org` (`name`, `org_name`, `delete_token`) COMMENT "租户内角色名称唯一"
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT "角色表（支持系统角色和租户自定义角色）";
-- Create "role_permissions" table
CREATE TABLE `role_permissions` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT "主键 ID",
  `role_id` bigint NOT NULL COMMENT "角色 ID（引用 roles.id）",
  `org_name` varchar(36) NOT NULL COMMENT "组织名称（冗余字段，优化查询）",
  `obj` varchar(64) NOT NULL COMMENT "资源对象（如 project, model, cluster, enum）",
  `act` varchar(64) NOT NULL COMMENT "操作动作（如 create, read, update, delete, *）",
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT "创建时间",
  PRIMARY KEY (`id`),
  INDEX `idx_org_name` (`org_name`) COMMENT "按组织查询权限",
  INDEX `idx_role_org` (`role_id`, `org_name`) COMMENT "查询某角色在某组织的权限（优化主查询路径）",
  UNIQUE INDEX `uk_role_obj_act` (`role_id`, `obj`, `act`) COMMENT "防止重复权限",
  CONSTRAINT `fk_role_perms_role` FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT "角色权限表（仅存储自定义角色权限）";
-- Create "user_roles" table
CREATE TABLE `user_roles` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT "主键 ID",
  `user_id` varchar(36) NOT NULL COMMENT "用户 ID（引用 users.id）",
  `role_id` bigint NOT NULL COMMENT "角色 ID（引用 roles.id）",
  `org_name` varchar(36) NOT NULL COMMENT "组织名称（多租户隔离）",
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT "创建时间",
  PRIMARY KEY (`id`),
  INDEX `idx_org_name` (`org_name`) COMMENT "按组织查询所有用户角色",
  INDEX `idx_role_org` (`role_id`, `org_name`) COMMENT "查询某角色在某组织的所有用户",
  INDEX `idx_user_org` (`user_id`, `org_name`) COMMENT "查询用户在某组织的所有角色",
  UNIQUE INDEX `uk_user_role_org` (`user_id`, `role_id`, `org_name`) COMMENT "防止重复角色绑定",
  CONSTRAINT `fk_user_roles_role` FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `fk_user_roles_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT "用户角色绑定表（支持多租户）";
