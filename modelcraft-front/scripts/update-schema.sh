#!/bin/bash
# 从远程 Git 仓库更新 GraphQL schema 文件
# 用法：./scripts/update-schema.sh [分支名]
# 默认分支：master

set -e

REPO_URL="https://git.woa.com/lukemxjia/modelcraft-go.git"
BRANCH="${1:-master}"
SCHEMA_DIR="src/graphql/schema"
FILES=(
    "api/graph/schema/base.graphql"
    "api/graph/schema/cluster.graphql"
    "api/graph/schema/enum.graphql"
    "api/graph/schema/field.graphql"
    "api/graph/schema/model.graphql"
    "api/graph/schema/permission.graphql"
    "api/graph/schema/project.graphql"
    "api/graph/schema/schema.graphql"
    "api/graph/schema/user_management.graphql"
)

echo "正在从 $REPO_URL ($BRANCH) 更新 GraphQL schema 文件..."
echo "目标目录: $SCHEMA_DIR"

# 确保目标目录存在
mkdir -p "$SCHEMA_DIR"

# 下载每个文件
for file in "${FILES[@]}"; do
    filename=$(basename "$file")
    echo "下载 $filename..."
    
    # 使用 git archive 下载单个文件
    if git archive --remote="$REPO_URL" "$BRANCH" "$file" | tar -xO > "$SCHEMA_DIR/$filename"; then
        echo "  ✓ $filename 下载成功"
    else
        echo "  ✗ $filename 下载失败"
        exit 1
    fi
done

echo ""
echo "✅ 所有 GraphQL schema 文件已更新到 $SCHEMA_DIR/"
echo ""
echo "文件列表:"
ls -la "$SCHEMA_DIR/"