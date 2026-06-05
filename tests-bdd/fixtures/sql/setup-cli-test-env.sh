#!/usr/bin/env bash
# tests-bdd/fixtures/sql/setup-cli-test-env.sh
#
# 为 CLI BDD 测试搭建 demo_ecommerce 测试数据库。
# 会在指定 MySQL 实例上创建并填充 demo_ecommerce 数据库。
#
# 使用方式:
#   ./setup-cli-test-env.sh [host] [port] [user] [password]
#
# 默认值（对应 docker-compose.local.yml）:
#   host=127.0.0.1  port=3307  user=root  password=modelcraft123
#
# 示例:
#   ./setup-cli-test-env.sh                             # 本地开发默认
#   ./setup-cli-test-env.sh 9.135.32.8 3307 root pass  # 远程集群

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SQL_FILE="$SCRIPT_DIR/demo_ecommerce_seed.sql"

DB_HOST="${1:-127.0.0.1}"
DB_PORT="${2:-3307}"
DB_USER="${3:-root}"
DB_PASS="${4:-modelcraft123}"
DB_NAME="demo_ecommerce"

echo "🔌 Connecting to MySQL at $DB_HOST:$DB_PORT as $DB_USER ..."

# 测试连接
if ! mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASS" \
    -e "SELECT 1;" > /dev/null 2>&1; then
  echo "❌ Cannot connect to MySQL. Check host/port/credentials."
  exit 1
fi

echo "📦 Creating database $DB_NAME ..."
mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASS" 2>/dev/null \
  -e "CREATE DATABASE IF NOT EXISTS \`$DB_NAME\` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

echo "🗄️  Loading schema and seed data ..."
mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASS" \
  "$DB_NAME" < "$SQL_FILE" 2>/dev/null

echo ""
echo "✅ demo_ecommerce setup complete!"
echo ""
echo "   Tables loaded:"
mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASS" "$DB_NAME" 2>/dev/null \
  -e "SELECT TABLE_NAME, TABLE_ROWS FROM information_schema.TABLES WHERE TABLE_SCHEMA='$DB_NAME' ORDER BY TABLE_NAME;" \
  | column -t
echo ""
echo "   Next steps:"
echo "   1. Create a ModelCraft project pointing to this MySQL instance"
echo "   2. Import the demo_ecommerce database as models"
echo "   3. Create a PAT token for the end-user and update .env.test CLI_PAT_TOKEN"
echo "   4. Run: cd tests-bdd && npm run test:cli"
