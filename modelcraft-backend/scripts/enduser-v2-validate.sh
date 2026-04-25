#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="${1:-.env}"
SOURCE_RUN_ID="${2:-}"
RUN_ID="${3:-enduser_v2_validate_$(date -u +%Y%m%dT%H%M%SZ)}"

# shellcheck source=./db-env.sh
source "${ROOT_DIR}/scripts/db-env.sh" "${ENV_FILE}"

MYSQL_CMD=(mysql -h "${DB_HOST}" -P "${DB_PORT}" -u "${DB_USER}" "-p${DB_PASSWORD}")

if [[ -z "${SOURCE_RUN_ID}" ]]; then
  SOURCE_RUN_ID="$(${MYSQL_CMD[@]} -N -e "
SELECT run_id
FROM ${DB_NAME}.migration_audit_enduser_v2_counts
WHERE snapshot = 'after_backfill'
ORDER BY captured_at DESC
LIMIT 1;
" 2>/dev/null || true)"
fi

if [[ -z "${SOURCE_RUN_ID}" ]]; then
  echo "❌ 未找到可用的 backfill run_id。"
  echo "   请先执行 ./scripts/enduser-v2-backfill.sh，再执行校验。"
  exit 1
fi

echo "🔍 开始 enduser-v2 迁移校验"
echo "   validate_run_id=${RUN_ID}"
echo "   source_run_id=${SOURCE_RUN_ID}"

"${MYSQL_CMD[@]}" "${DB_NAME}" <<SQL
SET @validate_run_id := '${RUN_ID}';
SET @source_run_id := '${SOURCE_RUN_ID}';
SET @has_user_bundle_table := (
  SELECT COUNT(*)
  FROM information_schema.tables
  WHERE table_schema = DATABASE()
    AND table_name = 'end_user_user_bundles'
);
SET @has_accounts_table := (
  SELECT COUNT(*)
  FROM information_schema.tables
  WHERE table_schema = DATABASE()
    AND table_name = 'end_user_accounts'
);

CREATE TABLE IF NOT EXISTS migration_audit_enduser_v2_validation_checks (
  id BIGINT NOT NULL AUTO_INCREMENT,
  validate_run_id VARCHAR(64) NOT NULL,
  source_run_id VARCHAR(64) NOT NULL,
  check_name VARCHAR(128) NOT NULL,
  pass_flag TINYINT(1) NOT NULL,
  actual_value BIGINT NULL,
  expected_value BIGINT NULL,
  details VARCHAR(512) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_euv2_validate_checks_run (validate_run_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS migration_audit_enduser_v2_validation_conflicts (
  id BIGINT NOT NULL AUTO_INCREMENT,
  validate_run_id VARCHAR(64) NOT NULL,
  source_run_id VARCHAR(64) NOT NULL,
  conflict_type VARCHAR(64) NOT NULL,
  org_name VARCHAR(64) NULL,
  username VARCHAR(64) NULL,
  user_id VARCHAR(36) NULL,
  project_slug VARCHAR(64) NULL,
  detail TEXT NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_euv2_validate_conflicts_run (validate_run_id),
  KEY idx_euv2_validate_conflicts_type (validate_run_id, conflict_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

DELETE FROM migration_audit_enduser_v2_validation_checks WHERE validate_run_id = @validate_run_id;
DELETE FROM migration_audit_enduser_v2_validation_conflicts WHERE validate_run_id = @validate_run_id;

SET @before_users := IFNULL((
  SELECT row_count
  FROM migration_audit_enduser_v2_counts
  WHERE run_id = @source_run_id
    AND snapshot = 'before_backfill'
    AND table_name = 'end_user_users'
  ORDER BY id DESC
  LIMIT 1
), -1);

SET @after_users_snapshot := IFNULL((
  SELECT row_count
  FROM migration_audit_enduser_v2_counts
  WHERE run_id = @source_run_id
    AND snapshot = 'after_backfill'
    AND table_name = 'end_user_users'
  ORDER BY id DESC
  LIMIT 1
), -1);

SET @merged_rows := (
  SELECT COUNT(*)
  FROM migration_audit_enduser_v2_merge_log
  WHERE run_id = @source_run_id
);

SET @current_users := (SELECT COUNT(*) FROM end_user_users);
SET @expected_users := IF(@before_users >= 0, @before_users - @merged_rows, NULL);

SET @duplicate_org_username := (
  SELECT COUNT(*)
  FROM (
    SELECT org_name, username
    FROM end_user_users
    GROUP BY org_name, username
    HAVING COUNT(*) > 1
  ) d
);

SET @orphan_project_access := (
  SELECT COUNT(*)
  FROM end_user_project_access a
  LEFT JOIN end_user_users u
    ON u.id = a.end_user_id
   AND u.org_name = a.org_name
  WHERE u.id IS NULL
);

SET @users_without_access := (
  SELECT COUNT(*)
  FROM end_user_users u
  LEFT JOIN end_user_project_access a
    ON a.end_user_id = u.id
   AND a.org_name = u.org_name
  WHERE a.id IS NULL
);

SET @legacy_rows_remaining := (
  SELECT COUNT(*)
  FROM end_user_users u
  JOIN migration_audit_enduser_v2_merge_log m
    ON m.run_id = @source_run_id
   AND m.old_user_id = u.id
);

SET @bundle_missing_access := 0;
SET @sql_bundle_missing_access := IF(
  @has_user_bundle_table > 0,
  'SELECT COUNT(*)
   FROM end_user_user_bundles ub
   LEFT JOIN end_user_project_access a
     ON a.org_name = ub.org_name
    AND a.end_user_id = ub.user_id
    AND a.project_slug = ub.project_slug
   WHERE a.id IS NULL',
  'SELECT 0'
);
PREPARE stmt_bundle_missing_access FROM @sql_bundle_missing_access;
EXECUTE stmt_bundle_missing_access INTO @bundle_missing_access;
DEALLOCATE PREPARE stmt_bundle_missing_access;

SET @orphan_accounts := 0;
SET @sql_orphan_accounts := IF(
  @has_accounts_table > 0,
  'SELECT COUNT(*)
   FROM end_user_accounts acc
   LEFT JOIN end_user_users u
     ON u.org_name = acc.org_name
    AND u.id = acc.user_id
   WHERE u.id IS NULL',
  'SELECT 0'
);
PREPARE stmt_orphan_accounts FROM @sql_orphan_accounts;
EXECUTE stmt_orphan_accounts INTO @orphan_accounts;
DEALLOCATE PREPARE stmt_orphan_accounts;

INSERT INTO migration_audit_enduser_v2_validation_checks (
  validate_run_id,
  source_run_id,
  check_name,
  pass_flag,
  actual_value,
  expected_value,
  details
)
VALUES
  (@validate_run_id, @source_run_id, 'user_count_matches_before_minus_merged', IF(@expected_users IS NULL, 0, @current_users = @expected_users), @current_users, @expected_users, 'current users should equal before_backfill - merged rows'),
  (@validate_run_id, @source_run_id, 'user_count_matches_after_snapshot', IF(@after_users_snapshot < 0, 0, @current_users = @after_users_snapshot), @current_users, @after_users_snapshot, 'current users should match after_backfill snapshot'),
  (@validate_run_id, @source_run_id, 'duplicate_org_username', IF(@duplicate_org_username = 0, 1, 0), @duplicate_org_username, 0, 'org_name + username must remain unique'),
  (@validate_run_id, @source_run_id, 'orphan_project_access', IF(@orphan_project_access = 0, 1, 0), @orphan_project_access, 0, 'every access row must reference an existing end_user_users row'),
  (@validate_run_id, @source_run_id, 'users_without_project_access', IF(@users_without_access = 0, 1, 0), @users_without_access, 0, 'every end user should have at least one project access relation'),
  (@validate_run_id, @source_run_id, 'legacy_merged_user_remaining', IF(@legacy_rows_remaining = 0, 1, 0), @legacy_rows_remaining, 0, 'merged old_user_id rows should be removed from end_user_users'),
  (@validate_run_id, @source_run_id, 'bundle_rows_missing_access', IF(@bundle_missing_access = 0, 1, 0), @bundle_missing_access, 0, 'every end_user_user_bundles row should map to an access relation'),
  (@validate_run_id, @source_run_id, 'orphan_accounts', IF(@orphan_accounts = 0, 1, 0), @orphan_accounts, 0, 'every end_user_accounts row should reference an existing user');

INSERT INTO migration_audit_enduser_v2_validation_conflicts (
  validate_run_id,
  source_run_id,
  conflict_type,
  org_name,
  username,
  detail
)
SELECT
  @validate_run_id,
  @source_run_id,
  'duplicate_org_username',
  t.org_name,
  t.username,
  CONCAT('duplicate_rows=', t.duplicate_count)
FROM (
  SELECT org_name, username, COUNT(*) AS duplicate_count
  FROM end_user_users
  GROUP BY org_name, username
  HAVING COUNT(*) > 1
) t;

INSERT INTO migration_audit_enduser_v2_validation_conflicts (
  validate_run_id,
  source_run_id,
  conflict_type,
  org_name,
  username,
  user_id,
  detail
)
SELECT
  @validate_run_id,
  @source_run_id,
  'user_without_access',
  u.org_name,
  u.username,
  u.id,
  'no row found in end_user_project_access'
FROM end_user_users u
LEFT JOIN end_user_project_access a
  ON a.org_name = u.org_name
 AND a.end_user_id = u.id
WHERE a.id IS NULL;

INSERT INTO migration_audit_enduser_v2_validation_conflicts (
  validate_run_id,
  source_run_id,
  conflict_type,
  org_name,
  user_id,
  project_slug,
  detail
)
SELECT
  @validate_run_id,
  @source_run_id,
  'bundle_conflict',
  c.org_name,
  c.canonical_user_id,
  c.project_slug,
  CONCAT('bundle_count=', c.bundle_count, ', bundle_ids=', c.bundle_ids)
FROM migration_audit_enduser_v2_bundle_conflicts c
WHERE c.run_id = @source_run_id;

SET @sql_insert_bundle_missing_conflicts := IF(
  @has_user_bundle_table > 0,
  'INSERT INTO migration_audit_enduser_v2_validation_conflicts (
      validate_run_id,
      source_run_id,
      conflict_type,
      org_name,
      user_id,
      project_slug,
      detail
    )
    SELECT
      @validate_run_id,
      @source_run_id,
      ''bundle_without_access'',
      ub.org_name,
      ub.user_id,
      ub.project_slug,
      CONCAT(''bundle_id='', ub.bundle_id)
    FROM end_user_user_bundles ub
    LEFT JOIN end_user_project_access a
      ON a.org_name = ub.org_name
     AND a.end_user_id = ub.user_id
     AND a.project_slug = ub.project_slug
    WHERE a.id IS NULL',
  'SELECT 1'
);
PREPARE stmt_insert_bundle_missing_conflicts FROM @sql_insert_bundle_missing_conflicts;
EXECUTE stmt_insert_bundle_missing_conflicts;
DEALLOCATE PREPARE stmt_insert_bundle_missing_conflicts;
SQL

echo "✅ 校验完成，输出检查结果"

"${MYSQL_CMD[@]}" "${DB_NAME}" -e "
SELECT
  check_name,
  pass_flag,
  actual_value,
  expected_value,
  details
FROM migration_audit_enduser_v2_validation_checks
WHERE validate_run_id='${RUN_ID}'
ORDER BY id;
"

"${MYSQL_CMD[@]}" "${DB_NAME}" -e "
SELECT
  conflict_type,
  COUNT(*) AS conflict_count
FROM migration_audit_enduser_v2_validation_conflicts
WHERE validate_run_id='${RUN_ID}'
GROUP BY conflict_type
ORDER BY conflict_count DESC;
"

echo "ℹ️ 如需查看冲突明细:"
echo "   SELECT * FROM migration_audit_enduser_v2_validation_conflicts WHERE validate_run_id='${RUN_ID}' ORDER BY id;"