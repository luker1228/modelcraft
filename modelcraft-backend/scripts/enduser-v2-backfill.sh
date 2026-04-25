#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="${1:-.env}"
RUN_ID="${2:-enduser_v2_$(date -u +%Y%m%dT%H%M%SZ)}"

# shellcheck source=./db-env.sh
source "${ROOT_DIR}/scripts/db-env.sh" "${ENV_FILE}"

MYSQL_CMD=(mysql -h "${DB_HOST}" -P "${DB_PORT}" -u "${DB_USER}" "-p${DB_PASSWORD}")

has_access_table="$(${MYSQL_CMD[@]} -N -e "
SELECT COUNT(*)
FROM information_schema.tables
WHERE table_schema = '${DB_NAME}'
  AND table_name = 'end_user_project_access';
")"

if [[ "${has_access_table}" -eq 0 ]]; then
  echo "❌ missing table end_user_project_access in ${DB_NAME}."
  echo "   请先完成 tasks 1.1（创建 end_user_project_access）后再执行回填。"
  exit 1
fi

echo "🔄 开始 enduser-v2 数据回填"
echo "   run_id=${RUN_ID}"

echo "🧮 执行合并与回填 SQL..."
"${MYSQL_CMD[@]}" "${DB_NAME}" <<SQL
SET @run_id := '${RUN_ID}';
SET @has_project_slug := (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'end_user_users'
    AND column_name = 'project_slug'
);
SET @has_user_bundle_table := (
  SELECT COUNT(*)
  FROM information_schema.tables
  WHERE table_schema = DATABASE()
    AND table_name = 'end_user_user_bundles'
);
SET @has_role_user_table := (
  SELECT COUNT(*)
  FROM information_schema.tables
  WHERE table_schema = DATABASE()
    AND table_name = 'end_user_role_users'
);
SET @has_role_user_project_slug := (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'end_user_role_users'
    AND column_name = 'project_slug'
);
SET @has_accounts_table := (
  SELECT COUNT(*)
  FROM information_schema.tables
  WHERE table_schema = DATABASE()
    AND table_name = 'end_user_accounts'
);

CREATE TABLE IF NOT EXISTS migration_audit_enduser_v2_counts (
  id BIGINT NOT NULL AUTO_INCREMENT,
  run_id VARCHAR(64) NOT NULL,
  snapshot VARCHAR(32) NOT NULL,
  table_name VARCHAR(64) NOT NULL,
  row_count BIGINT NOT NULL,
  captured_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_euv2_counts_run_snapshot (run_id, snapshot)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS migration_audit_enduser_v2_user_map (
  run_id VARCHAR(64) NOT NULL,
  old_user_id VARCHAR(36) NOT NULL,
  keep_user_id VARCHAR(36) NOT NULL,
  org_name VARCHAR(64) NOT NULL,
  username VARCHAR(64) NOT NULL,
  source_project_slug VARCHAR(64) NULL,
  source_created_at DATETIME NULL,
  source_updated_at DATETIME NULL,
  mapped_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (run_id, old_user_id),
  KEY idx_euv2_map_keep (run_id, keep_user_id),
  KEY idx_euv2_map_org_user (run_id, org_name, username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS migration_audit_enduser_v2_merge_log (
  run_id VARCHAR(64) NOT NULL,
  old_user_id VARCHAR(36) NOT NULL,
  keep_user_id VARCHAR(36) NOT NULL,
  org_name VARCHAR(64) NOT NULL,
  username VARCHAR(64) NOT NULL,
  source_project_slug VARCHAR(64) NULL,
  merged_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (run_id, old_user_id),
  KEY idx_euv2_merge_keep (run_id, keep_user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS migration_audit_enduser_v2_bundle_conflicts (
  id BIGINT NOT NULL AUTO_INCREMENT,
  run_id VARCHAR(64) NOT NULL,
  org_name VARCHAR(64) NOT NULL,
  canonical_user_id VARCHAR(36) NOT NULL,
  project_slug VARCHAR(64) NOT NULL,
  bundle_count INT NOT NULL,
  bundle_ids TEXT NOT NULL,
  detected_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_euv2_bundle_conflicts_run (run_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

DELETE FROM migration_audit_enduser_v2_counts WHERE run_id = @run_id;
DELETE FROM migration_audit_enduser_v2_user_map WHERE run_id = @run_id;
DELETE FROM migration_audit_enduser_v2_merge_log WHERE run_id = @run_id;
DELETE FROM migration_audit_enduser_v2_bundle_conflicts WHERE run_id = @run_id;

INSERT INTO migration_audit_enduser_v2_counts(run_id, snapshot, table_name, row_count)
SELECT @run_id, 'before_backfill', 'end_user_users', COUNT(*) FROM end_user_users;

INSERT INTO migration_audit_enduser_v2_counts(run_id, snapshot, table_name, row_count)
SELECT @run_id, 'before_backfill', 'end_user_project_access', COUNT(*) FROM end_user_project_access;

SET @sql_insert_accounts_before := IF(
  @has_accounts_table > 0,
  'INSERT INTO migration_audit_enduser_v2_counts(run_id, snapshot, table_name, row_count)
   SELECT @run_id, ''before_backfill'', ''end_user_accounts'', COUNT(*) FROM end_user_accounts',
  'SELECT 1'
);
PREPARE stmt_accounts_before FROM @sql_insert_accounts_before;
EXECUTE stmt_accounts_before;
DEALLOCATE PREPARE stmt_accounts_before;

SET @sql_insert_role_users_before := IF(
  @has_role_user_table > 0,
  'INSERT INTO migration_audit_enduser_v2_counts(run_id, snapshot, table_name, row_count)
   SELECT @run_id, ''before_backfill'', ''end_user_role_users'', COUNT(*) FROM end_user_role_users',
  'SELECT 1'
);
PREPARE stmt_role_users_before FROM @sql_insert_role_users_before;
EXECUTE stmt_role_users_before;
DEALLOCATE PREPARE stmt_role_users_before;

SET @sql_insert_user_bundles_before := IF(
  @has_user_bundle_table > 0,
  'INSERT INTO migration_audit_enduser_v2_counts(run_id, snapshot, table_name, row_count)
   SELECT @run_id, ''before_backfill'', ''end_user_user_bundles'', COUNT(*) FROM end_user_user_bundles',
  'SELECT 1'
);
PREPARE stmt_user_bundles_before FROM @sql_insert_user_bundles_before;
EXECUTE stmt_user_bundles_before;
DEALLOCATE PREPARE stmt_user_bundles_before;

SET @project_expr := IF(@has_project_slug > 0, 'u.project_slug', 'NULL');
SET @sql_build_user_map := CONCAT(
  'INSERT INTO migration_audit_enduser_v2_user_map (',
  ' run_id, old_user_id, keep_user_id, org_name, username, source_project_slug, source_created_at, source_updated_at',
  ') ',
  'SELECT @run_id, ranked.old_user_id, ranked.keep_user_id, ranked.org_name, ranked.username, ranked.source_project_slug, ranked.source_created_at, ranked.source_updated_at ',
  'FROM (',
  '  SELECT ',
  '    u.id AS old_user_id,',
  '    FIRST_VALUE(u.id) OVER (PARTITION BY u.org_name, u.username ORDER BY u.created_at ASC, u.id ASC) AS keep_user_id,',
  '    u.org_name,',
  '    u.username,',
  '    ', @project_expr, ' AS source_project_slug,',
  '    u.created_at AS source_created_at,',
  '    u.updated_at AS source_updated_at ',
  '  FROM end_user_users u',
  ') ranked'
);
PREPARE stmt_build_user_map FROM @sql_build_user_map;
EXECUTE stmt_build_user_map;
DEALLOCATE PREPARE stmt_build_user_map;

INSERT INTO migration_audit_enduser_v2_merge_log (
  run_id,
  old_user_id,
  keep_user_id,
  org_name,
  username,
  source_project_slug
)
SELECT
  run_id,
  old_user_id,
  keep_user_id,
  org_name,
  username,
  source_project_slug
FROM migration_audit_enduser_v2_user_map
WHERE run_id = @run_id
  AND old_user_id <> keep_user_id;

START TRANSACTION;

INSERT INTO end_user_project_access (
  id,
  end_user_id,
  org_name,
  project_slug,
  permission_bundle_id,
  granted_by,
  granted_at
)
SELECT
  UUID(),
  m.keep_user_id,
  m.org_name,
  m.source_project_slug,
  NULL,
  NULL,
  NOW(3)
FROM migration_audit_enduser_v2_user_map m
WHERE m.run_id = @run_id
  AND m.source_project_slug IS NOT NULL
  AND m.source_project_slug <> ''
GROUP BY m.keep_user_id, m.org_name, m.source_project_slug
ON DUPLICATE KEY UPDATE
  end_user_id = end_user_id;

SET @sql_bundle_conflict := IF(
  @has_user_bundle_table > 0,
  'INSERT INTO migration_audit_enduser_v2_bundle_conflicts (
      run_id,
      org_name,
      canonical_user_id,
      project_slug,
      bundle_count,
      bundle_ids
    )
    SELECT
      @run_id,
      ub.org_name,
      m.keep_user_id,
      ub.project_slug,
      COUNT(DISTINCT ub.bundle_id),
      GROUP_CONCAT(DISTINCT ub.bundle_id ORDER BY ub.bundle_id SEPARATOR '','')
    FROM end_user_user_bundles ub
    JOIN migration_audit_enduser_v2_user_map m
      ON m.run_id = @run_id
     AND m.org_name = ub.org_name
     AND m.old_user_id = ub.user_id
    GROUP BY ub.org_name, m.keep_user_id, ub.project_slug
    HAVING COUNT(DISTINCT ub.bundle_id) > 1',
  'SELECT 1'
);
PREPARE stmt_bundle_conflict FROM @sql_bundle_conflict;
EXECUTE stmt_bundle_conflict;
DEALLOCATE PREPARE stmt_bundle_conflict;

SET @sql_bundle_to_access := IF(
  @has_user_bundle_table > 0,
  'INSERT INTO end_user_project_access (
      id,
      end_user_id,
      org_name,
      project_slug,
      permission_bundle_id,
      granted_by,
      granted_at
    )
    SELECT
      UUID(),
      m.keep_user_id,
      ub.org_name,
      ub.project_slug,
      MIN(ub.bundle_id) AS permission_bundle_id,
      NULL,
      NOW(3)
    FROM end_user_user_bundles ub
    JOIN migration_audit_enduser_v2_user_map m
      ON m.run_id = @run_id
     AND m.org_name = ub.org_name
     AND m.old_user_id = ub.user_id
    GROUP BY m.keep_user_id, ub.org_name, ub.project_slug
    ON DUPLICATE KEY UPDATE
      permission_bundle_id = COALESCE(end_user_project_access.permission_bundle_id, VALUES(permission_bundle_id))',
  'SELECT 1'
);
PREPARE stmt_bundle_to_access FROM @sql_bundle_to_access;
EXECUTE stmt_bundle_to_access;
DEALLOCATE PREPARE stmt_bundle_to_access;

SET @sql_merge_accounts := IF(
  @has_accounts_table > 0,
  'UPDATE end_user_accounts a
   JOIN migration_audit_enduser_v2_user_map m
     ON m.run_id = @run_id
    AND m.org_name = a.org_name
    AND m.old_user_id = a.user_id
   SET a.user_id = m.keep_user_id
   WHERE m.old_user_id <> m.keep_user_id',
  'SELECT 1'
);
PREPARE stmt_merge_accounts FROM @sql_merge_accounts;
EXECUTE stmt_merge_accounts;
DEALLOCATE PREPARE stmt_merge_accounts;

SET @role_user_insert_sql := IF(
  @has_role_user_table = 0,
  'SELECT 1',
  IF(
    @has_role_user_project_slug > 0,
    'INSERT IGNORE INTO end_user_role_users (id, org_name, project_slug, role_id, user_id, created_at)
     SELECT UUID(), ru.org_name, ru.project_slug, ru.role_id, m.keep_user_id, ru.created_at
     FROM end_user_role_users ru
     JOIN migration_audit_enduser_v2_user_map m
       ON m.run_id = @run_id
      AND m.org_name = ru.org_name
      AND m.old_user_id = ru.user_id
     WHERE m.old_user_id <> m.keep_user_id',
    'INSERT IGNORE INTO end_user_role_users (id, org_name, role_id, user_id, created_at)
     SELECT UUID(), ru.org_name, ru.role_id, m.keep_user_id, ru.created_at
     FROM end_user_role_users ru
     JOIN migration_audit_enduser_v2_user_map m
       ON m.run_id = @run_id
      AND m.org_name = ru.org_name
      AND m.old_user_id = ru.user_id
     WHERE m.old_user_id <> m.keep_user_id'
  )
);
PREPARE stmt_role_user_insert FROM @role_user_insert_sql;
EXECUTE stmt_role_user_insert;
DEALLOCATE PREPARE stmt_role_user_insert;

SET @role_user_delete_sql := IF(
  @has_role_user_table = 0,
  'SELECT 1',
  'DELETE ru FROM end_user_role_users ru
   JOIN migration_audit_enduser_v2_user_map m
     ON m.run_id = @run_id
    AND m.org_name = ru.org_name
    AND m.old_user_id = ru.user_id
   WHERE m.old_user_id <> m.keep_user_id'
);
PREPARE stmt_role_user_delete FROM @role_user_delete_sql;
EXECUTE stmt_role_user_delete;
DEALLOCATE PREPARE stmt_role_user_delete;

SET @sql_user_bundle_insert := IF(
  @has_user_bundle_table > 0,
  'INSERT IGNORE INTO end_user_user_bundles (id, org_name, project_slug, user_id, bundle_id, granted_at)
   SELECT UUID(), ub.org_name, ub.project_slug, m.keep_user_id, ub.bundle_id, ub.granted_at
   FROM end_user_user_bundles ub
   JOIN migration_audit_enduser_v2_user_map m
     ON m.run_id = @run_id
    AND m.org_name = ub.org_name
    AND m.old_user_id = ub.user_id
   WHERE m.old_user_id <> m.keep_user_id',
  'SELECT 1'
);
PREPARE stmt_user_bundle_insert FROM @sql_user_bundle_insert;
EXECUTE stmt_user_bundle_insert;
DEALLOCATE PREPARE stmt_user_bundle_insert;

SET @sql_user_bundle_delete := IF(
  @has_user_bundle_table > 0,
  'DELETE ub FROM end_user_user_bundles ub
   JOIN migration_audit_enduser_v2_user_map m
     ON m.run_id = @run_id
    AND m.org_name = ub.org_name
    AND m.old_user_id = ub.user_id
   WHERE m.old_user_id <> m.keep_user_id',
  'SELECT 1'
);
PREPARE stmt_user_bundle_delete FROM @sql_user_bundle_delete;
EXECUTE stmt_user_bundle_delete;
DEALLOCATE PREPARE stmt_user_bundle_delete;

INSERT INTO end_user_project_access (
  id,
  end_user_id,
  org_name,
  project_slug,
  permission_bundle_id,
  granted_by,
  granted_at
)
SELECT
  UUID(),
  m.keep_user_id,
  a.org_name,
  a.project_slug,
  a.permission_bundle_id,
  a.granted_by,
  a.granted_at
FROM end_user_project_access a
JOIN migration_audit_enduser_v2_user_map m
  ON m.run_id = @run_id
 AND m.org_name = a.org_name
 AND m.old_user_id = a.end_user_id
WHERE m.old_user_id <> m.keep_user_id
ON DUPLICATE KEY UPDATE
  permission_bundle_id = COALESCE(end_user_project_access.permission_bundle_id, VALUES(permission_bundle_id));

DELETE a FROM end_user_project_access a
JOIN migration_audit_enduser_v2_user_map m
  ON m.run_id = @run_id
 AND m.org_name = a.org_name
 AND m.old_user_id = a.end_user_id
WHERE m.old_user_id <> m.keep_user_id;

DELETE u
FROM end_user_users u
JOIN migration_audit_enduser_v2_merge_log l
  ON l.run_id = @run_id
 AND l.old_user_id = u.id;

COMMIT;

INSERT INTO migration_audit_enduser_v2_counts(run_id, snapshot, table_name, row_count)
SELECT @run_id, 'after_backfill', 'end_user_users', COUNT(*) FROM end_user_users;

INSERT INTO migration_audit_enduser_v2_counts(run_id, snapshot, table_name, row_count)
SELECT @run_id, 'after_backfill', 'end_user_project_access', COUNT(*) FROM end_user_project_access;

SET @sql_insert_accounts_after := IF(
  @has_accounts_table > 0,
  'INSERT INTO migration_audit_enduser_v2_counts(run_id, snapshot, table_name, row_count)
   SELECT @run_id, ''after_backfill'', ''end_user_accounts'', COUNT(*) FROM end_user_accounts',
  'SELECT 1'
);
PREPARE stmt_accounts_after FROM @sql_insert_accounts_after;
EXECUTE stmt_accounts_after;
DEALLOCATE PREPARE stmt_accounts_after;

SET @sql_insert_role_users_after := IF(
  @has_role_user_table > 0,
  'INSERT INTO migration_audit_enduser_v2_counts(run_id, snapshot, table_name, row_count)
   SELECT @run_id, ''after_backfill'', ''end_user_role_users'', COUNT(*) FROM end_user_role_users',
  'SELECT 1'
);
PREPARE stmt_role_users_after FROM @sql_insert_role_users_after;
EXECUTE stmt_role_users_after;
DEALLOCATE PREPARE stmt_role_users_after;

SET @sql_insert_user_bundles_after := IF(
  @has_user_bundle_table > 0,
  'INSERT INTO migration_audit_enduser_v2_counts(run_id, snapshot, table_name, row_count)
   SELECT @run_id, ''after_backfill'', ''end_user_user_bundles'', COUNT(*) FROM end_user_user_bundles',
  'SELECT 1'
);
PREPARE stmt_user_bundles_after FROM @sql_insert_user_bundles_after;
EXECUTE stmt_user_bundles_after;
DEALLOCATE PREPARE stmt_user_bundles_after;
SQL

echo "✅ 回填完成，run_id=${RUN_ID}"

echo "📋 合并摘要"
"${MYSQL_CMD[@]}" "${DB_NAME}" -e "
SELECT
  '${RUN_ID}' AS run_id,
  (SELECT COUNT(*) FROM migration_audit_enduser_v2_merge_log WHERE run_id='${RUN_ID}') AS merged_user_rows,
  (SELECT COUNT(*) FROM migration_audit_enduser_v2_bundle_conflicts WHERE run_id='${RUN_ID}') AS bundle_conflicts,
  (SELECT row_count FROM migration_audit_enduser_v2_counts WHERE run_id='${RUN_ID}' AND snapshot='before_backfill' AND table_name='end_user_users' ORDER BY id DESC LIMIT 1) AS users_before,
  (SELECT row_count FROM migration_audit_enduser_v2_counts WHERE run_id='${RUN_ID}' AND snapshot='after_backfill'  AND table_name='end_user_users' ORDER BY id DESC LIMIT 1) AS users_after,
  (SELECT row_count FROM migration_audit_enduser_v2_counts WHERE run_id='${RUN_ID}' AND snapshot='after_backfill'  AND table_name='end_user_project_access' ORDER BY id DESC LIMIT 1) AS project_access_after;
"

echo "ℹ️ 如需进一步一致性核对，请执行: ./scripts/enduser-v2-validate.sh ${ENV_FILE} ${RUN_ID}"