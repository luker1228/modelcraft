#!/bin/bash
set -e

# 为现有测试文件添加缺失的测试
# 用法: ./scripts/add-missing-tests.sh internal/domain/membership coverage.out

PACKAGE_PATH="$1"
COVERAGE_FILE="${2:-coverage.out}"

if [ -z "$PACKAGE_PATH" ]; then
    echo "❌ 错误: 请提供包路径"
    exit 1
fi

if [ ! -f "$COVERAGE_FILE" ]; then
    echo "❌ 错误: 找不到覆盖率文件 $COVERAGE_FILE"
    exit 1
fi

PACKAGE_NAME=$(basename "$PACKAGE_PATH")
TEST_FILE="${PACKAGE_PATH}/${PACKAGE_NAME}_test.go"

if [ ! -f "$TEST_FILE" ]; then
    echo "⚠️  测试文件不存在: $TEST_FILE"
    exit 0
fi

echo "🔍 分析未覆盖的函数..."

# 查找覆盖率为0%的函数
UNCOVERED_FUNCS=$(go tool cover -func="$COVERAGE_FILE" 2>/dev/null | \
    grep "^modelcraft/${PACKAGE_PATH}/" | \
    grep "0.0%$" | \
    awk '{print $2}' | \
    grep -v "^init$" || true)

if [ -z "$UNCOVERED_FUNCS" ]; then
    echo "✅ 该包所有函数都已有覆盖"
    exit 0
fi

echo "📝 找到未覆盖的函数:"
echo "$UNCOVERED_FUNCS" | while read -r func; do
    echo "   - $func"
done

# 检查测试文件中是否已有对应的测试
echo ""
echo "🔧 检查测试文件中缺失的测试..."

ADDED_COUNT=0

echo "$UNCOVERED_FUNCS" | while read -r func; do
    # 去掉包名前缀，提取函数名
    func_name=$(echo "$func" | sed 's/.*\.\([^.]*\)$/\1/')
    
    # 检查测试是否已存在
    if grep -q "func Test.*${func_name}" "$TEST_FILE" 2>/dev/null; then
        echo "   ⏭️  测试已存在: Test${func_name}"
        continue
    fi
    
    echo "   ➕ 添加测试模板: Test${func_name}"
    
    # 在文件末尾添加测试模板
    cat >> "$TEST_FILE" << EOF

// TODO: 自动生成的测试模板，需要实现具体逻辑
func Test${func_name}(t *testing.T) {
	t.Skip("TODO: 实现 ${func_name} 的测试")
	// TODO: 添加测试用例
	// 1. 准备测试数据
	// 2. 调用被测试函数
	// 3. 验证结果
}
EOF
    
    ADDED_COUNT=$((ADDED_COUNT + 1))
done

if [ $ADDED_COUNT -gt 0 ]; then
    echo ""
    echo "✅ 添加了 $ADDED_COUNT 个测试模板"
    echo "💡 请编辑 $TEST_FILE 实现具体的测试逻辑"
else
    echo ""
    echo "ℹ️  所有函数都已有对应的测试（可能需要实现具体逻辑）"
fi
