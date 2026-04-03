package bizerrors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNew 测试 New 函数
func TestNew(t *testing.T) {
	t.Run("创建简单错误", func(t *testing.T) {
		err := New("这是一个错误")
		assert.Error(t, err)
		assert.Equal(t, "这是一个错误", err.Error())
	})

	t.Run("错误包含堆栈信息", func(t *testing.T) {
		err := New("带堆栈的错误")

		// 验证错误可以格式化为包含堆栈信息
		errStr := err.Error()
		fmt.Printf("err = %+v", err)
		assert.Contains(t, errStr, "带堆栈的错误")

		// 使用 %+v 格式化应该包含堆栈跟踪
		detailedErr := fmt.Sprintf("%+v", err)
		assert.Contains(t, detailedErr, "带堆栈的错误")
		assert.Contains(t, detailedErr, "TestNew") // 应该包含测试函数名
	})
}

// TestErrorf 测试 Errorf 函数
func TestErrorf(t *testing.T) {
	t.Run("格式化错误消息", func(t *testing.T) {
		name := "张三"
		age := 25
		err := Errorf("用户 %s 的年龄 %d 无效", name, age)

		assert.Error(t, err)
		assert.Equal(t, "用户 张三 的年龄 25 无效", err.Error())
	})

	t.Run("复杂格式化", func(t *testing.T) {
		err := Errorf("操作失败: %s, 代码: %d", "权限不足", 403)
		assert.Equal(t, "操作失败: 权限不足, 代码: 403", err.Error())
	})
}

// TestWrap 测试 Wrap 函数
func TestWrap(t *testing.T) {
	t.Run("包装基础错误", func(t *testing.T) {
		baseErr := New("基础错误")
		wrappedErr := Wrap(baseErr, "包装消息")

		assert.Error(t, wrappedErr)
		assert.Contains(t, wrappedErr.Error(), "包装消息")
		assert.Contains(t, wrappedErr.Error(), "基础错误")

		// 验证错误链
		assert.Equal(t, baseErr, Cause(wrappedErr))
	})

	t.Run("包装 nil 错误", func(t *testing.T) {
		wrappedErr := Wrap(nil, "包装消息")
		assert.Nil(t, wrappedErr)
	})
}

// TestWrapf 测试 Wrapf 函数
func TestWrapf(t *testing.T) {
	t.Run("格式化包装错误", func(t *testing.T) {
		baseErr := New("数据库错误")
		userID := "12345"
		wrappedErr := Wrapf(baseErr, "用户 %s 查询失败", userID)

		assert.Error(t, wrappedErr)
		assert.Contains(t, wrappedErr.Error(), "用户 12345 查询失败")
		assert.Contains(t, wrappedErr.Error(), "数据库错误")

		// 验证错误链
		assert.Equal(t, baseErr, Cause(wrappedErr))
	})
}

// TestWithStack 测试 WithStack 函数
func TestWithStack(t *testing.T) {
	t.Run("添加堆栈跟踪", func(t *testing.T) {
		baseErr := errors.New("标准库错误")
		stackErr := WithStack(baseErr)

		assert.Error(t, stackErr)
		assert.Equal(t, baseErr.Error(), stackErr.Error())

		// 验证堆栈信息
		detailedErr := fmt.Sprintf("%+v", stackErr)
		assert.Contains(t, detailedErr, "标准库错误")
		assert.Contains(t, detailedErr, "TestWithStack") // 应该包含测试函数名
	})

	t.Run("nil 错误处理", func(t *testing.T) {
		stackErr := WithStack(nil)
		assert.Nil(t, stackErr)
	})
}

// TestCause 测试 Cause 函数
func TestCause(t *testing.T) {
	t.Run("获取根错误", func(t *testing.T) {
		rootErr := New("根错误")
		wrapped1 := Wrap(rootErr, "第一层包装")
		wrapped2 := Wrap(wrapped1, "第二层包装")

		cause := Cause(wrapped2)
		assert.Equal(t, rootErr, cause)
		assert.Equal(t, "根错误", cause.Error())
	})

	t.Run("未包装的错误", func(t *testing.T) {
		simpleErr := New("简单错误")
		cause := Cause(simpleErr)
		assert.Equal(t, simpleErr, cause)
	})

	t.Run("nil 错误", func(t *testing.T) {
		cause := Cause(nil)
		assert.Nil(t, cause)
	})
}

// TestIs 测试 Is 函数
func TestIs(t *testing.T) {
	t.Run("直接比较", func(t *testing.T) {
		targetErr := New("目标错误")
		assert.True(t, Is(targetErr, targetErr))
	})

	t.Run("错误链比较", func(t *testing.T) {
		targetErr := New("目标错误")
		wrappedErr := Wrap(targetErr, "包装")

		assert.True(t, Is(wrappedErr, targetErr))
		assert.True(t, Is(wrappedErr, wrappedErr))
	})

	t.Run("不同类型错误", func(t *testing.T) {
		err1 := New("错误1")
		err2 := New("错误2")

		assert.False(t, Is(err1, err2))
		assert.True(t, Is(err1, err1))
	})

	t.Run("与标准库错误兼容", func(t *testing.T) {
		stdErr := errors.New("标准错误")
		wrappedStdErr := Wrap(stdErr, "包装标准错误")

		assert.True(t, Is(wrappedStdErr, stdErr))
	})
}

// 定义自定义错误类型
type CustomError struct {
	msg  string
	code int
}

func (e *CustomError) Error() string {
	return e.msg
}

// TestAs 测试 As 函数
func TestAs(t *testing.T) {
	t.Run("类型转换成功", func(t *testing.T) {
		customErr := &CustomError{msg: "自定义错误", code: 1001}
		wrappedErr := Wrap(customErr, "包装自定义错误")

		var target *CustomError
		assert.True(t, As(wrappedErr, &target))
		assert.Equal(t, customErr, target)
		assert.Equal(t, "自定义错误", target.Error())
		assert.Equal(t, 1001, target.code)
	})

	t.Run("类型转换失败", func(t *testing.T) {
		simpleErr := New("简单错误")
		var target *CustomError
		assert.False(t, As(simpleErr, &target))
		assert.Nil(t, target)
	})

	t.Run("nil 错误处理", func(t *testing.T) {
		var target *CustomError
		assert.False(t, As(nil, &target))
		assert.Nil(t, target)
	})
}

// TestUnwrap 测试 Unwrap 函数
func TestUnwrap(t *testing.T) {
	t.Run("解包包装错误", func(t *testing.T) {
		baseErr := New("基础错误")
		wrappedErr := Wrap(baseErr, "包装消息")

		// github.com/pkg/errors 的 Wrap 会形成 withStack -> withMessage -> cause 的链
		unwrapped1 := Unwrap(wrappedErr)
		assert.NotNil(t, unwrapped1)
		assert.Equal(t, "包装消息: 基础错误", unwrapped1.Error())

		unwrapped2 := Unwrap(unwrapped1)
		assert.Equal(t, baseErr, unwrapped2)
	})

	t.Run("解包未包装错误", func(t *testing.T) {
		simpleErr := New("简单错误")
		unwrapped := Unwrap(simpleErr)
		assert.Nil(t, unwrapped)
	})

	t.Run("多层解包", func(t *testing.T) {
		rootErr := New("根错误")
		middleErr := Wrap(rootErr, "中间层")
		topErr := Wrap(middleErr, "顶层")

		// topErr(withStack) -> withMessage("顶层") -> middleErr(withStack) -> withMessage("中间层") -> rootErr
		u1 := Unwrap(topErr)
		assert.NotNil(t, u1)
		assert.Equal(t, "顶层: 中间层: 根错误", u1.Error())

		u2 := Unwrap(u1)
		assert.Equal(t, middleErr, u2)

		u3 := Unwrap(u2)
		assert.NotNil(t, u3)
		assert.Equal(t, "中间层: 根错误", u3.Error())

		u4 := Unwrap(u3)
		assert.Equal(t, rootErr, u4)

		u5 := Unwrap(u4)
		assert.Nil(t, u5)
	})
}

// TestErrorChain 测试错误链功能
func TestErrorChain(t *testing.T) {
	t.Run("完整错误链测试", func(t *testing.T) {
		// 创建错误链
		dbErr := New("数据库连接失败")
		repoErr := Wrapf(dbErr, "用户仓库操作失败")
		serviceErr := Wrap(repoErr, "业务服务错误")
		handlerErr := Wrapf(serviceErr, "HTTP 处理失败")

		// 验证错误链
		assert.Equal(t, "HTTP 处理失败: 业务服务错误: 用户仓库操作失败: 数据库连接失败",
			handlerErr.Error())

		// 验证根错误
		rootCause := Cause(handlerErr)
		assert.Equal(t, dbErr, rootCause)
		assert.Equal(t, "数据库连接失败", rootCause.Error())

		// 验证中间错误
		assert.True(t, Is(handlerErr, repoErr))
		assert.True(t, Is(handlerErr, serviceErr))
		assert.True(t, Is(handlerErr, dbErr))

		// 验证堆栈信息
		detailedErr := fmt.Sprintf("%+v", handlerErr)
		assert.Contains(t, detailedErr, "数据库连接失败")
		assert.Contains(t, detailedErr, "用户仓库操作失败")
		assert.Contains(t, detailedErr, "业务服务错误")
		assert.Contains(t, detailedErr, "HTTP 处理失败")
	})
}

// TestCompatibility 测试与标准库的兼容性
func TestCompatibility(t *testing.T) {
	t.Run("与标准库 errors 兼容", func(t *testing.T) {
		// 验证我们的错误包与标准库函数兼容
		stdErr := errors.New("标准错误")
		ourErr := New("我们的错误")

		// 标准库的 Is 应该能处理我们的错误
		assert.True(t, errors.Is(ourErr, ourErr))
		assert.False(t, errors.Is(ourErr, stdErr))

		// 我们的 Is 应该能处理标准库错误
		assert.True(t, Is(stdErr, stdErr))
		assert.False(t, Is(stdErr, ourErr))
	})

	t.Run("错误消息格式", func(t *testing.T) {
		err := Errorf("格式化错误: %s", "测试")

		// 基本错误消息应该一致
		assert.Equal(t, "格式化错误: 测试", err.Error())

		// %v 格式化应该与标准库一致
		assert.Equal(t, "格式化错误: 测试", fmt.Sprintf("%v", err))

		// %+v 格式化应该包含堆栈信息
		detailed := fmt.Sprintf("%+v", err)
		assert.Contains(t, detailed, "格式化错误: 测试")
		assert.Contains(t, detailed, "TestCompatibility")
	})
}

// BenchmarkErrorCreation 性能基准测试
func BenchmarkErrorCreation(b *testing.B) {
	b.Run("New 函数", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = New("基准测试错误")
		}
	})

	b.Run("Errorf 函数", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Errorf("格式化错误: %d", i)
		}
	})

	b.Run("Wrap 函数", func(b *testing.B) {
		baseErr := New("基础错误")
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = Wrap(baseErr, "包装错误")
		}
	})
}
