-- model_sync_job 表 SQL 定义
-- syncModelsFromDB 异步任务表（以 databaseName 为维度，独立于 model_database_sync_job）

CREATE TABLE IF NOT EXISTS `model_sync_job` (
  `id`               VARCHAR(36)  NOT NULL COMMENT '任务唯一标识符',
  `org_name`         VARCHAR(64)  NOT NULL COMMENT '所属组织名称',
  `project_slug`     VARCHAR(64)  NOT NULL COMMENT '所属项目标识符',
  `database_name`    VARCHAR(128) NOT NULL COMMENT '目标数据库名称',
  `table_names`      JSON         NOT NULL COMMENT '指定同步的表名列表，空数组表示全量',
  `status`           ENUM('pending','running','succeeded','partial_success','failed') NOT NULL COMMENT '任务状态',
  `total_tables`     INT          NOT NULL DEFAULT 0 COMMENT '扫描到的总表数',
  `processed_tables` INT          NOT NULL DEFAULT 0 COMMENT '已处理表数',
  `created_models`   INT          NOT NULL DEFAULT 0 COMMENT '新建模型数',
  `synced_models`    INT          NOT NULL DEFAULT 0 COMMENT '已同步模型数',
  `failed_count`     INT          NOT NULL DEFAULT 0 COMMENT '失败表数',
  `failed_tables`    JSON         NOT NULL COMMENT '失败明细，格式：[{tableName, message}]',
  `started_at`       DATETIME(3)  NULL     COMMENT 'worker 开始时间',
  `finished_at`      DATETIME(3)  NULL     COMMENT '任务结束时间',
  `created_at`       DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at`       DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',

  PRIMARY KEY (`id`),
  INDEX `idx_model_sync_job_project_db` (`org_name`, `project_slug`, `database_name`, `created_at`),
  INDEX `idx_model_sync_job_status`     (`org_name`, `project_slug`, `status`)

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='syncModelsFromDB 异步任务表';
