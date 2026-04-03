#!/bin/bash
set -e

# 生成有意义的测试脚本
# 用法: ./scripts/generate-meaningful-tests.sh internal/domain/membership coverage.out

PACKAGE_PATH="$1"
COVERAGE_FILE="${2:-coverage.out}"

if [ -z "$PACKAGE_PATH" ]; then
    echo "❌ 错误: 请提供包路径"
    echo "用法: $0 internal/domain/membership [coverage.out]"
    exit 1
fi

if [ ! -f "$COVERAGE_FILE" ]; then
    echo "❌ 错误: 找不到覆盖率文件 $COVERAGE_FILE"
    exit 1
fi

PACKAGE_NAME=$(basename "$PACKAGE_PATH")
TEST_FILE="${PACKAGE_PATH}/${PACKAGE_NAME}_test.go"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🧠 智能测试生成器 - $PACKAGE_NAME"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 分析未覆盖的函数
echo "🔍 分析未覆盖的代码..."
UNCOVERED_FUNCS=$(go tool cover -func="$COVERAGE_FILE" 2>/dev/null | \
    grep "^modelcraft/${PACKAGE_PATH}/" | \
    awk '{
        cov = $NF;
        gsub(/%/, "", cov);
        if (cov < 100) {
            print $2 "|" cov;
        }
    }' || true)

if [ -z "$UNCOVERED_FUNCS" ]; then
    echo "✅ 该包所有函数都已有 100% 覆盖"
    exit 0
fi

echo "📋 需要补充测试的函数:"
echo "$UNCOVERED_FUNCS" | while IFS='|' read -r func coverage; do
    echo "   • $func (${coverage}%)"
done
echo ""

# 收集源文件信息用于分析
SOURCE_FILES=$(find "$PACKAGE_PATH" -name "*.go" -not -name "*_test.go")

echo "🧪 生成有意义的测试用例..."
echo ""

# 为每个未完全覆盖的函数生成测试
echo "$UNCOVERED_FUNCS" | while IFS='|' read -r func coverage; do
    # 提取函数名（去掉包名前缀）
    func_name=$(echo "$func" | sed 's/.*\.\([^.]*\)$/\1/')
    
    # 检查是否已有测试
    if [ -f "$TEST_FILE" ] && grep -q "func Test.*${func_name}" "$TEST_FILE" 2>/dev/null; then
        echo "   ⏭️  ${func_name}: 测试已存在，需要完善"
        continue
    fi
    
    echo "   ✨ ${func_name}: 正在生成测试..."
    
    # 调用 AI 辅助脚本生成有意义的测试
    ./scripts/ai-generate-test.sh "$PACKAGE_PATH" "$func_name" "$TEST_FILE"
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ 测试生成完成"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "💡 下一步:"
echo "   1. 运行 go test ./${PACKAGE_PATH} -v 验证测试"
echo "   2. 检查生成的测试逻辑是否符合预期"
echo "   3. 根据需要调整测试用例"
echo ""
