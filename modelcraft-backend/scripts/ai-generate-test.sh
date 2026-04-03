#!/bin/bash
set -e

# AI 辅助生成有意义的测试
# 用法: ./scripts/ai-generate-test.sh internal/domain/membership CreateMembership membership_test.go

PACKAGE_PATH="$1"
FUNC_NAME="$2"
TEST_FILE="$3"

if [ -z "$PACKAGE_PATH" ] || [ -z "$FUNC_NAME" ] || [ -z "$TEST_FILE" ]; then
    echo "❌ 错误: 参数不完整"
    exit 1
fi

PACKAGE_NAME=$(basename "$PACKAGE_PATH")

# 查找函数定义
FUNC_DEF=$(find "$PACKAGE_PATH" -name "*.go" -not -name "*_test.go" -exec grep -A 10 "func.*${FUNC_NAME}" {} \; | head -20)

if [ -z "$FUNC_DEF" ]; then
    echo "❌ 找不到函数定义: $FUNC_NAME"
    exit 1
fi

# 分析函数类型和生成对应的测试
echo "📝 分析函数 $FUNC_NAME..."

# 创建测试文件（如果不存在）
if [ ! -f "$TEST_FILE" ]; then
    cat > "$TEST_FILE" << EOF
package $PACKAGE_NAME

import (
	"testing"
)

EOF
fi

# 根据函数特征生成测试
# 1. 检查是否是构造函数（New开头）
if echo "$FUNC_NAME" | grep -q "^New"; then
    cat >> "$TEST_FILE" << EOF
func Test${FUNC_NAME}(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "valid input",
			wantErr: false,
		},
		{
			name:    "invalid input",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: 准备输入参数
			
			// TODO: 调用 ${FUNC_NAME}
			
			// TODO: 验证结果
			if tt.wantErr {
				// 期望错误
			} else {
				// 验证对象创建成功
			}
		})
	}
}

EOF

# 2. 检查是否是验证函数（Validate开头）
elif echo "$FUNC_NAME" | grep -q "^Validate"; then
    cat >> "$TEST_FILE" << EOF
func Test${FUNC_NAME}(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{} // TODO: 替换为实际类型
		wantErr bool
	}{
		{
			name:    "valid data",
			input:   nil, // TODO: 提供有效数据
			wantErr: false,
		},
		{
			name:    "invalid data - empty",
			input:   nil, // TODO: 提供无效数据
			wantErr: true,
		},
		{
			name:    "invalid data - wrong format",
			input:   nil, // TODO: 提供格式错误的数据
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ${FUNC_NAME}(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("${FUNC_NAME}() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

EOF

# 3. 检查是否是查询函数（Get/Find/List开头）
elif echo "$FUNC_NAME" | grep -qE "^(Get|Find|List|Query)"; then
    cat >> "$TEST_FILE" << EOF
func Test${FUNC_NAME}(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() // 准备测试数据
		wantErr bool
		wantNil bool
	}{
		{
			name: "found",
			setup: func() {
				// TODO: 创建测试数据
			},
			wantErr: false,
			wantNil: false,
		},
		{
			name: "not found",
			setup: func() {
				// TODO: 不创建数据或创建不匹配的数据
			},
			wantErr: false,
			wantNil: true,
		},
		{
			name: "error case",
			setup: func() {
				// TODO: 准备会导致错误的场景
			},
			wantErr: true,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			
			// TODO: 调用 ${FUNC_NAME}
			result, err := ${FUNC_NAME}() // 添加实际参数
			
			if (err != nil) != tt.wantErr {
				t.Errorf("${FUNC_NAME}() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if tt.wantNil && result != nil {
				t.Errorf("${FUNC_NAME}() = %v, want nil", result)
			}
			
			if !tt.wantNil && result == nil {
				t.Error("${FUNC_NAME}() returned nil, want non-nil")
			}
		})
	}
}

EOF

# 4. 检查是否是修改函数（Update/Set/Modify开头）
elif echo "$FUNC_NAME" | grep -qE "^(Update|Set|Modify|Change)"; then
    cat >> "$TEST_FILE" << EOF
func Test${FUNC_NAME}(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() interface{} // 返回待修改对象
		input   interface{}        // 新值
		want    interface{}        // 期望结果
		wantErr bool
	}{
		{
			name: "successful update",
			setup: func() interface{} {
				// TODO: 创建初始对象
				return nil
			},
			input:   nil, // TODO: 提供有效的更新数据
			want:    nil, // TODO: 期望的结果
			wantErr: false,
		},
		{
			name: "invalid update",
			setup: func() interface{} {
				// TODO: 创建初始对象
				return nil
			},
			input:   nil, // TODO: 提供无效数据
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := tt.setup()
			
			// TODO: 调用 ${FUNC_NAME}
			err := ${FUNC_NAME}(obj, tt.input)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("${FUNC_NAME}() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				// TODO: 验证对象已正确更新
				// if !reflect.DeepEqual(obj, tt.want) {
				//     t.Errorf("${FUNC_NAME}() result = %v, want %v", obj, tt.want)
				// }
			}
		})
	}
}

EOF

# 5. 检查是否是删除函数（Delete/Remove开头）
elif echo "$FUNC_NAME" | grep -qE "^(Delete|Remove)"; then
    cat >> "$TEST_FILE" << EOF
func Test${FUNC_NAME}(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() // 准备测试数据
		id      string // TODO: 替换为实际ID类型
		wantErr bool
	}{
		{
			name: "delete existing",
			setup: func() {
				// TODO: 创建要删除的对象
			},
			id:      "test-id",
			wantErr: false,
		},
		{
			name:    "delete non-existing",
			setup:   func() {},
			id:      "non-existing-id",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			
			err := ${FUNC_NAME}(tt.id)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("${FUNC_NAME}() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

EOF

# 6. 默认通用测试模板
else
    cat >> "$TEST_FILE" << EOF
func Test${FUNC_NAME}(t *testing.T) {
	tests := []struct {
		name    string
		// TODO: 添加测试用例的输入字段
		wantErr bool
	}{
		{
			name:    "normal case",
			wantErr: false,
		},
		{
			name:    "error case",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: 准备测试数据
			
			// TODO: 调用 ${FUNC_NAME}
			
			// TODO: 验证结果
		})
	}
}

EOF
fi

echo "✅ 生成测试: Test${FUNC_NAME}"
