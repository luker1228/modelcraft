-- Add batch_id and database_id to model_sync_job

ALTER TABLE `model_sync_job`
  ADD COLUMN `batch_id`    VARCHAR(36) NOT NULL DEFAULT '' COMMENT '批次 ID，同批次多条 job 共享' AFTER `id`,
  ADD COLUMN `database_id` VARCHAR(36) NOT NULL DEFAULT '' COMMENT '关联 model_database.id' AFTER `batch_id`,
  ADD INDEX `idx_model_sync_job_batch` (`batch_id`),
  ADD INDEX `idx_model_sync_job_database_id` (`database_id`);
