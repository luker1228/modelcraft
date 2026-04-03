#!/bin/bash
set -e

# 生成测试模板脚本
# 用法: ./scripts/generate-test-template.sh internal/domain/auth

PACKAGE_PATH="$1"

if [ -z "$PACKAGE_PATH" ]; then
    echo "❌ 错误: 请提供包路径"
    echo "用法: $0 internal/domain/package"
    exit 1
fi

# 提取包名
PACKAGE_NAME=$(basename "$PACKAGE_PATH")
TEST_FILE="${PACKAGE_PATH}/${PACKAGE_NAME}_test.go"

if [ -f "$TEST_FILE" ]; then
    echo "⚠️  测试文件已存在: $TEST_FILE"
    exit 0
fi

echo "✨ 创建测试文件: $TEST_FILE"

# 获取包声明
PACKAGE_DECL=$(grep -h "^package " "${PACKAGE_PATH}"/*.go 2>/dev/null | head -1 || echo "package ${PACKAGE_NAME}")

# 查找所有公开函数和方法
echo "🔍 分析源文件，查找需要测试的函数..."

# 创建测试文件
cat > "$TEST_FILE" << EOF
$PACKAGE_DECL

import (
	"testing"
)

// TODO: 这是自动生成的测试模板，请根据实际情况补充测试用例

EOF

# 查找所有公开的函数（大写字母开头）
find "$PACKAGE_PATH" -name "*.go" -not -name "*_test.go" -exec grep -h "^func [A-Z]" {} \; 2>/dev/null | while read -r func_line; do
    # 提取函数名
    func_name=$(echo "$func_line" | sed 's/func \([A-Z][a-zA-Z0-9]*\).*/\1/')
    
    cat >> "$TEST_FILE" << EOF
func Test${func_name}(t *testing.T) {
	t.Skip("TODO: 实现 ${func_name} 的测试")
	// TODO: 添加测试用例
}

EOF
done

# 查找所有公开的方法（接收器方法）
find "$PACKAGE_PATH" -name "*.go" -not -name "*_test.go" -exec grep -h "^func ([^)]*) [A-Z]" {} \; 2>/dev/null | while read -r method_line; do
    # 提取方法名
    method_name=$(echo "$method_line" | sed 's/func ([^)]*) \([A-Z][a-zA-Z0-9]*\).*/\1/')
    receiver=$(echo "$method_line" | sed 's/func (\([^)]*\)).*/\1/')
    
    cat >> "$TEST_FILE" << EOF
func Test${method_name}(t *testing.T) {
	t.Skip("TODO: 实现 ${method_name} 的测试 (receiver: ${receiver})")
	// TODO: 添加测试用例
}

EOF
done

echo "✅ 测试文件创建完成: $TEST_FILE"
echo "💡 请编辑该文件并实现具体的测试用例"
