#!/bin/bash

# CMS 功能验证脚本
# 用于测试 CMS 的完整数据流

set -e

echo "🧪 ModelCraft CMS 功能验证"
echo "================================"

# 配置
GRAPHQL_ENDPOINT="${GRAPHQL_ENDPOINT:-http://localhost:8080/graphql}"
PROJECT_ID="${PROJECT_ID:-default}"

echo ""
echo "📋 测试配置:"
echo "  GraphQL Endpoint: $GRAPHQL_ENDPOINT"
echo "  Project ID: $PROJECT_ID"
echo ""

# 检查 GraphQL 服务是否运行
echo "1️⃣ 检查 GraphQL 服务..."
if ! curl -s -f "$GRAPHQL_ENDPOINT" > /dev/null; then
    echo "❌ GraphQL 服务未运行 (端口 8080)"
    echo "   请先启动: cd modelcraft-go && task run"
    exit 1
fi
echo "✅ GraphQL 服务正常"

# 测试查询 Models
echo ""
echo "2️⃣ 测试查询 Models..."
MODELS_QUERY='{
  "query": "query GetModels($input: ModelQueryInput!) { models(input: $input) { edges { node { id name title description } } totalCount } }",
  "variables": {
    "input": {
      "projectId": "'$PROJECT_ID'",
      "databaseName": "",
      "limit": 10,
      "offset": 0
    }
  }
}'

MODELS_RESPONSE=$(curl -s -X POST "$GRAPHQL_ENDPOINT" \
  -H "Content-Type: application/json" \
  -d "$MODELS_QUERY")

echo "$MODELS_RESPONSE" | jq .

MODEL_COUNT=$(echo "$MODELS_RESPONSE" | jq -r '.data.models.totalCount // 0')
echo ""
echo "📊 找到 $MODEL_COUNT 个 Model"

if [ "$MODEL_COUNT" -eq 0 ]; then
    echo "⚠️  没有 Model，无法测试 CMS 内容管理"
    echo "   请先创建 Model: 访问 http://localhost:3000/app/models"
    exit 0
fi

# 获取第一个 Model
FIRST_MODEL_ID=$(echo "$MODELS_RESPONSE" | jq -r '.data.models.edges[0].node.id')
FIRST_MODEL_NAME=$(echo "$MODELS_RESPONSE" | jq -r '.data.models.edges[0].node.name')

echo ""
echo "🎯 选择测试 Model:"
echo "  ID: $FIRST_MODEL_ID"
echo "  Name: $FIRST_MODEL_NAME"

# 测试获取 Model Schema
echo ""
echo "3️⃣ 测试获取 Model Schema..."
SCHEMA_QUERY='{
  "query": "query ModelJsonSchema($projectId: ID!, $id: ID!) { modelJsonSchema(projectId: $projectId, id: $id) { modelId modelName schema } }",
  "variables": {
    "projectId": "'$PROJECT_ID'",
    "id": "'$FIRST_MODEL_ID'"
  }
}'

SCHEMA_RESPONSE=$(curl -s -X POST "$GRAPHQL_ENDPOINT" \
  -H "Content-Type: application/json" \
  -d "$SCHEMA_QUERY")

echo "$SCHEMA_RESPONSE" | jq .

MODEL_NAME=$(echo "$SCHEMA_RESPONSE" | jq -r '.data.modelJsonSchema.modelName')
echo ""
echo "📝 Model Schema 名称: $MODEL_NAME"

if [ "$MODEL_NAME" == "null" ] || [ -z "$MODEL_NAME" ]; then
    echo "❌ 无法获取 Model Schema"
    exit 1
fi

echo "✅ Model Schema 获取成功"

# 测试运行时 API (需要 modelcraft-agent 运行)
echo ""
echo "4️⃣ 测试运行时 API (Runtime GraphQL)..."
RUNTIME_ENDPOINT="${RUNTIME_ENDPOINT:-http://localhost:8000/graphql}"

if ! curl -s -f "$RUNTIME_ENDPOINT" > /dev/null 2>&1; then
    echo "⚠️  Runtime API 未运行 (端口 8000)"
    echo "   CMS 内容读写需要 modelcraft-agent 服务"
    echo "   启动命令: cd modelcraft-agent && make dev"
else
    echo "✅ Runtime API 正常"

    # 尝试查询内容
    echo ""
    echo "5️⃣ 测试查询内容 (findMany${MODEL_NAME})..."

    CONTENT_QUERY='{
      "query": "query { findMany'$MODEL_NAME'(take: 10, skip: 0) { id } }"
    }'

    CONTENT_RESPONSE=$(curl -s -X POST "$RUNTIME_ENDPOINT" \
      -H "Content-Type: application/json" \
      -d "$CONTENT_QUERY" 2>/dev/null || echo '{"errors":[{"message":"Query failed"}]}')

    echo "$CONTENT_RESPONSE" | jq .

    if echo "$CONTENT_RESPONSE" | jq -e '.data' > /dev/null 2>&1; then
        CONTENT_COUNT=$(echo "$CONTENT_RESPONSE" | jq -r ".data.findMany${MODEL_NAME} | length")
        echo ""
        echo "📦 找到 $CONTENT_COUNT 条内容"
        echo "✅ CMS 内容查询成功"
    else
        echo ""
        echo "⚠️  内容查询失败 (可能 Model 还没有生成表)"
    fi
fi

# 总结
echo ""
echo "================================"
echo "✅ CMS 功能验证完成"
echo ""
echo "📌 访问 CMS 页面:"
echo "  - 列表页: http://localhost:3000/app/cms"
echo "  - 内容管理: http://localhost:3000/app/cms/$FIRST_MODEL_ID"
echo ""
echo "📚 验证步骤:"
echo "  1. 访问 /app/cms 应该能看到 $MODEL_COUNT 个 Model"
echo "  2. 点击 Model 卡片进入内容管理页面"
echo "  3. 点击 'Create New' 可以创建内容"
echo "  4. 内容列表应该显示所有已创建的内容"
echo ""
