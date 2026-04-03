#!/bin/bash
set -e

# 分析未覆盖代码并生成测试脚本
# 用法: ./scripts/analyze-uncovered.sh coverage.out

COVERAGE_FILE="${1:-coverage.out}"

if [ ! -f "$COVERAGE_FILE" ]; then
    echo "❌ 错误: 找不到覆盖率文件 $COVERAGE_FILE"
    exit 1
fi

echo "🔍 分析未覆盖的代码..."
echo ""

# 创建临时文件
TEMP_UNCOVERED=$(mktemp)
trap "rm -f $TEMP_UNCOVERED" EXIT

# 找出覆盖率低于95%的domain包
go tool cover -func="$COVERAGE_FILE" 2>/dev/null | grep "^modelcraft/internal/domain" | awk '{
  split($1, parts, "/");
  pkg = parts[1]"/"parts[2]"/"parts[3]"/"parts[4];
  split($NF, cov, "%");
  count[pkg]++;
  sum[pkg] += cov[1];
}
END {
  for (pkg in count) {
    avg = sum[pkg] / count[pkg];
    if (avg < 95) {
      printf "%s|%.1f\n", pkg, avg;
    }
  }
}' | sort -t'|' -k2 -rn > "$TEMP_UNCOVERED"

if [ ! -s "$TEMP_UNCOVERED" ]; then
    echo "✅ 所有包都已达标"
    exit 0
fi

echo "📋 需要补充测试的包:"
echo ""

# 按优先级处理（从覆盖率高到低）
while IFS='|' read -r pkg coverage; do
    pkg_path="${pkg#modelcraft/}"  # 移除 modelcraft/ 前缀
    gap=$(awk "BEGIN {printf \"%.1f\", 95 - $coverage}")
    
    echo "  📦 $pkg (当前: ${coverage}%, 差距: ${gap}%)"
    
    # 检查是否有测试文件
    test_file="${pkg_path}_test.go"
    if [ ! -f "$test_file" ]; then
        echo "     ⚠️  测试文件不存在: $test_file"
        echo "     ✨ 正在创建基础测试文件..."
        ./scripts/generate-test-template.sh "$pkg_path"
    else
        echo "     ✅ 测试文件已存在: $test_file"
        echo "     🔧 分析未覆盖函数..."
        ./scripts/add-missing-tests.sh "$pkg_path" "$COVERAGE_FILE"
    fi
    
    echo ""
done < "$TEMP_UNCOVERED"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ 分析完成"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
