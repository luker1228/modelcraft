#!/bin/bash
set -e

# 自动补充单测脚本
# 用法: ./scripts/auto-fix-coverage.sh [--max-iterations N] [--package PKG]

MAX_ITERATIONS=10
TARGET_PACKAGE=""
REQUIRED_COVERAGE=95

# 解析参数
while [[ $# -gt 0 ]]; do
    case $1 in
        --max-iterations)
            MAX_ITERATIONS="$2"
            shift 2
            ;;
        --package)
            TARGET_PACKAGE="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🤖 自动补充单元测试（目标: ${REQUIRED_COVERAGE}%）"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  最大迭代次数: $MAX_ITERATIONS"
if [ -n "$TARGET_PACKAGE" ]; then
    echo "  目标包: $TARGET_PACKAGE"
else
    echo "  目标包: 所有 domain 包"
fi
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

iteration=0
last_total_coverage=0

while [ $iteration -lt $MAX_ITERATIONS ]; do
    iteration=$((iteration + 1))
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "🔄 迭代 $iteration/$MAX_ITERATIONS"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    
    # 运行测试生成覆盖率
    echo "📊 运行测试并生成覆盖率报告..."
    if [ -n "$TARGET_PACKAGE" ]; then
        go test "./$TARGET_PACKAGE" -coverprofile=coverage.out -covermode=atomic -coverpkg="./internal/domain/..." 2>&1 | grep -v "no test files" || true
    else
        go test ./internal/domain/... -coverprofile=coverage.out -covermode=atomic -coverpkg=./internal/domain/... 2>&1 | grep -v "no test files" || true
    fi
    
    echo ""
    
    # 检查覆盖率
    echo "📈 检查当前覆盖率..."
    if ./scripts/check-domain-coverage.sh coverage.out 2>&1; then
        echo ""
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "✅ 成功！所有包已达到 ${REQUIRED_COVERAGE}% 覆盖率要求"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "  总迭代次数: $iteration"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        exit 0
    fi
    
    echo ""
    echo "⚠️  仍有包未达标，开始分析并生成测试..."
    echo ""
    
    # 分析未覆盖的代码
    ./scripts/analyze-uncovered.sh coverage.out
    
    # 检查是否有进展
    current_total=$(go tool cover -func=coverage.out | grep "^total:" | awk '{print $3}' | sed 's/%//')
    current_total_int=${current_total%.*}
    last_total_int=${last_total_coverage%.*}
    
    if [ "$current_total_int" -le "$last_total_int" ]; then
        echo ""
        echo "⚠️  覆盖率没有提升（当前: ${current_total}%, 上次: ${last_total_coverage}%）"
        echo "可能需要手动检查和编写测试"
    else
        echo ""
        echo "📈 覆盖率提升: ${last_total_coverage}% → ${current_total}%"
    fi
    
    last_total_coverage=$current_total
    
    echo ""
    echo "暂停 2 秒..."
    sleep 2
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "⚠️  达到最大迭代次数 ($MAX_ITERATIONS)，仍未完全达标"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  最终覆盖率: ${last_total_coverage}%"
echo "  目标覆盖率: ${REQUIRED_COVERAGE}%"
echo ""
echo "💡 建议："
echo "  1. 查看生成的测试文件并手动完善"
echo "  2. 运行 'task test-coverage' 查看详细报告"
echo "  3. 针对高优先级的包手动补充测试"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
exit 1
