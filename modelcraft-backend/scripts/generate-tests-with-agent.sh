#!/bin/bash
set -e

# 使用 Agent 智能生成单元测试
# 用法: ./scripts/generate-tests-with-agent.sh internal/domain/membership coverage.out

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

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🤖 AI Agent 测试生成器 - $PACKAGE_NAME"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  规范: Given-When-Then"
echo "  要求: 有意义的测试用例，不生成空测试"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 1. 分析包的当前覆盖率
echo "📊 分析包覆盖率..."
CURRENT_COVERAGE=$(go tool cover -func="$COVERAGE_FILE" 2>/dev/null | \
    grep "^modelcraft/${PACKAGE_PATH}/" | \
    awk '{
        split($1, parts, "/");
        pkg = parts[1]"/"parts[2]"/"parts[3]"/"parts[4];
        split($NF, cov, "%");
        count[pkg]++;
        sum[pkg] += cov[1];
    }
    END {
        for (pkg in count) {
            avg = sum[pkg] / count[pkg];
            printf "%.1f", avg;
        }
    }')

if [ -z "$CURRENT_COVERAGE" ]; then
    CURRENT_COVERAGE="0.0"
fi

echo "  当前覆盖率: ${CURRENT_COVERAGE}%"
echo "  目标覆盖率: 95%"
echo ""

# 2. 收集包的所有源文件
echo "📂 收集源文件..."
SOURCE_FILES=$(find "$PACKAGE_PATH" -name "*.go" -not -name "*_test.go" -type f)
if [ -z "$SOURCE_FILES" ]; then
    echo "❌ 找不到源文件"
    exit 1
fi

echo "$SOURCE_FILES" | while read -r file; do
    echo "  - $(basename $file)"
done
echo ""

# 3. 查找未覆盖或覆盖不足的函数
echo "🔍 分析未充分覆盖的函数..."
UNCOVERED_FUNCS=$(go tool cover -func="$COVERAGE_FILE" 2>/dev/null | \
    grep "^modelcraft/${PACKAGE_PATH}/" | \
    awk '{
        cov = $NF;
        gsub(/%/, "", cov);
        if (cov < 100 && $2 !~ /^init$/) {
            print $1 ":" $2 "|" cov;
        }
    }' || true)

if [ -z "$UNCOVERED_FUNCS" ]; then
    echo "✅ 所有函数都已有完整覆盖"
    exit 0
fi

FUNC_COUNT=$(echo "$UNCOVERED_FUNCS" | wc -l)
echo "  找到 $FUNC_COUNT 个需要补充测试的函数"
echo ""

# 4. 创建临时目录存储上下文信息
CONTEXT_DIR=$(mktemp -d)
trap "rm -rf $CONTEXT_DIR" EXIT

# 收集包的完整上下文
cat > "$CONTEXT_DIR/package_context.txt" << EOF
Package: $PACKAGE_PATH
Current Coverage: ${CURRENT_COVERAGE}%
Target Coverage: 95%

Source Files:
$SOURCE_FILES

Uncovered Functions:
$UNCOVERED_FUNCS

EOF

# 收集所有源代码
echo "📚 收集代码上下文..."
echo "$SOURCE_FILES" | while read -r file; do
    echo "=== $(basename $file) ===" >> "$CONTEXT_DIR/source_code.txt"
    cat "$file" >> "$CONTEXT_DIR/source_code.txt"
    echo "" >> "$CONTEXT_DIR/source_code.txt"
done

# 收集现有测试（如果有）
TEST_FILE="${PACKAGE_PATH}/${PACKAGE_NAME}_test.go"
if [ -f "$TEST_FILE" ]; then
    echo "  现有测试文件: $TEST_FILE"
    cat "$TEST_FILE" > "$CONTEXT_DIR/existing_tests.txt"
else
    echo "  没有现有测试文件"
    echo "" > "$CONTEXT_DIR/existing_tests.txt"
fi
echo ""

# 5. 调用 CodeBuddy Agent 生成测试
echo "🤖 启动 CodeBuddy Agent 生成测试..."
echo ""

# 创建 Agent 提示词
cat > "$CONTEXT_DIR/agent_prompt.md" << 'EOF'
# 任务：为 Go 包生成高质量单元测试

## 要求

1. **遵循 Given-When-Then 模式**
   - Given: 准备测试数据和前置条件
   - When: 执行被测试的函数
   - Then: 断言结果和验证行为

2. **生成有意义的测试用例**
   - 不生成空测试或 t.Skip()
   - 每个测试用例必须有明确的测试目的
   - 覆盖正常场景、边界条件、错误场景

3. **测试用例命名规范**
   - 使用表格驱动测试（table-driven tests）
   - 测试用例名称描述性强：如 "valid input", "empty string returns error", "nil pointer returns error"

4. **代码质量**
   - 使用真实的测试数据，不使用 nil 或空值作为占位符
   - 完整的断言，验证所有返回值
   - 清晰的错误消息

## 示例：好的测试

```go
func TestCreateMembership(t *testing.T) {
    tests := []struct {
        name    string
        orgID   string
        userID  string
        role    string
        wantErr bool
        errMsg  string
    }{
        {
            name:    "valid membership creation",
            orgID:   "org-123",
            userID:  "user-456",
            role:    "admin",
            wantErr: false,
        },
        {
            name:    "empty organization ID returns error",
            orgID:   "",
            userID:  "user-456",
            role:    "admin",
            wantErr: true,
            errMsg:  "organization ID cannot be empty",
        },
        {
            name:    "invalid role returns error",
            orgID:   "org-123",
            userID:  "user-456",
            role:    "invalid",
            wantErr: true,
            errMsg:  "invalid role",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Given: 准备测试数据
            membership := &Membership{
                OrgID:  tt.orgID,
                UserID: tt.userID,
                Role:   tt.role,
            }

            // When: 执行验证
            err := membership.Validate()

            // Then: 断言结果
            if tt.wantErr {
                if err == nil {
                    t.Errorf("expected error but got none")
                }
                if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
                    t.Errorf("error message = %v, want to contain %v", err.Error(), tt.errMsg)
                }
            } else {
                if err != nil {
                    t.Errorf("unexpected error: %v", err)
                }
            }
        })
    }
}
```

## 任务

请分析提供的代码，为所有覆盖率不足的函数生成高质量的单元测试。

EOF

echo "  提示词已准备: $CONTEXT_DIR/agent_prompt.md"
echo ""

# 显示将要处理的函数列表
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📋 待生成测试的函数列表:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "$UNCOVERED_FUNCS" | while IFS='|' read -r func_info coverage; do
    file_func=$(echo "$func_info" | sed 's/modelcraft\///')
    func_name=$(echo "$func_info" | sed 's/.*://')
    echo "  • $func_name (当前覆盖率: ${coverage}%)"
done
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 保存函数列表供 Agent 使用
echo "$UNCOVERED_FUNCS" > "$CONTEXT_DIR/functions_to_test.txt"

echo "💡 现在需要调用 CodeBuddy Code Agent 来生成测试..."
echo ""
echo "请使用以下命令手动启动 Agent:"
echo ""
echo "  cd /root/modelcraft_project/modelcraft/modelcraft-go"
echo "  codebuddy task run --agent test-generator \\"
echo "    --context '$CONTEXT_DIR' \\"
echo "    --package '$PACKAGE_PATH'"
echo ""
echo "或者继续使用自动化流程..."
echo ""

# 输出上下文文件位置供参考
echo "📁 上下文文件位置:"
echo "  - 包信息: $CONTEXT_DIR/package_context.txt"
echo "  - 源代码: $CONTEXT_DIR/source_code.txt"
echo "  - 现有测试: $CONTEXT_DIR/existing_tests.txt"
echo "  - 待测试函数: $CONTEXT_DIR/functions_to_test.txt"
echo "  - Agent 提示词: $CONTEXT_DIR/agent_prompt.md"
echo ""

exit 0
