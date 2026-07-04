package bizerrors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("创建简单错误", func(t *testing.T) {
		err := New("这是一个错误")
		assert.Error(t, err)
		assert.Equal(t, "这是一个错误", err.Error())
	})
}

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

func TestWrap(t *testing.T) {
	t.Run("包装基础错误", func(t *testing.T) {
		baseErr := New("基础错误")
		wrappedErr := Wrap(baseErr, "包装消息")

		assert.Error(t, wrappedErr)
		assert.Contains(t, wrappedErr.Error(), "包装消息")
		assert.Contains(t, wrappedErr.Error(), "基础错误")

		assert.Equal(t, baseErr, Cause(wrappedErr))
	})

	t.Run("包装 nil 错误", func(t *testing.T) {
		wrappedErr := Wrap(nil, "包装消息")
		assert.Nil(t, wrappedErr)
	})
}

func TestWrapf(t *testing.T) {
	t.Run("格式化包装错误", func(t *testing.T) {
		baseErr := New("数据库错误")
		userID := "12345"
		wrappedErr := Wrapf(baseErr, "用户 %s 查询失败", userID)

		assert.Error(t, wrappedErr)
		assert.Contains(t, wrappedErr.Error(), "用户 12345 查询失败")
		assert.Contains(t, wrappedErr.Error(), "数据库错误")

		assert.Equal(t, baseErr, Cause(wrappedErr))
	})
}

func TestWithStack(t *testing.T) {
	t.Run("添加堆栈跟踪", func(t *testing.T) {
		baseErr := errors.New("标准库错误")
		stackErr := WithStack(baseErr)

		assert.Error(t, stackErr)
		assert.Equal(t, baseErr.Error(), stackErr.Error())
		assert.NotEqual(t, baseErr, stackErr) // wrapped with stack frames
	})

	t.Run("nil 错误处理", func(t *testing.T) {
		stackErr := WithStack(nil)
		assert.Nil(t, stackErr)
	})
}

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

type CustomError struct {
	msg  string
	code int
}

func (e *CustomError) Error() string {
	return e.msg
}

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

func TestUnwrap(t *testing.T) {
	t.Run("解包包装错误", func(t *testing.T) {
		baseErr := New("基础错误")
		wrappedErr := Wrap(baseErr, "包装消息")

		// Wrap → withStack{wrapError{baseErr}}
		// Unwrap(withStack) → wrapError
		u1 := Unwrap(wrappedErr)
		assert.NotNil(t, u1)
		assert.Contains(t, u1.Error(), "包装消息")
		assert.Contains(t, u1.Error(), "基础错误")

		// Unwrap(wrapError) → baseErr
		u2 := Unwrap(u1)
		assert.Equal(t, baseErr, u2)
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

		// topErr = withStack{wrapError{msg:"顶层: 中间层: 根错误", err:middleErr}}
		// middleErr = withStack{wrapError{msg:"中间层: 根错误", err:rootErr}}

		u1 := Unwrap(topErr) // wrapError{msg:"顶层:...", err:middleErr}
		assert.NotNil(t, u1)
		assert.Contains(t, u1.Error(), "顶层")

		u2 := Unwrap(u1) // middleErr (withStack)
		assert.NotNil(t, u2)
		assert.Contains(t, u2.Error(), "中间层")

		u3 := Unwrap(u2) // wrapError{msg:"中间层:...", err:rootErr}
		assert.NotNil(t, u3)
		assert.Contains(t, u3.Error(), "中间层")

		u4 := Unwrap(u3) // rootErr
		assert.Equal(t, rootErr, u4)

		u5 := Unwrap(u4)
		assert.Nil(t, u5)
	})
}

func TestErrorChain(t *testing.T) {
	t.Run("完整错误链测试", func(t *testing.T) {
		dbErr := New("数据库连接失败")
		repoErr := Wrapf(dbErr, "用户仓库操作失败")
		serviceErr := Wrap(repoErr, "业务服务错误")
		handlerErr := Wrapf(serviceErr, "HTTP 处理失败")

		assert.Equal(t, "HTTP 处理失败: 业务服务错误: 用户仓库操作失败: 数据库连接失败",
			handlerErr.Error())

		rootCause := Cause(handlerErr)
		assert.Equal(t, dbErr, rootCause)
		assert.Equal(t, "数据库连接失败", rootCause.Error())

		assert.True(t, Is(handlerErr, repoErr))
		assert.True(t, Is(handlerErr, serviceErr))
		assert.True(t, Is(handlerErr, dbErr))
	})
}

func TestCompatibility(t *testing.T) {
	t.Run("与标准库 errors 兼容", func(t *testing.T) {
		stdErr := errors.New("标准错误")
		ourErr := New("我们的错误")

		assert.True(t, errors.Is(ourErr, ourErr))
		assert.False(t, errors.Is(ourErr, stdErr))

		assert.True(t, Is(stdErr, stdErr))
		assert.False(t, Is(stdErr, ourErr))
	})

	t.Run("错误消息格式", func(t *testing.T) {
		err := Errorf("格式化错误: %s", "测试")

		assert.Equal(t, "格式化错误: 测试", err.Error())
		assert.Equal(t, "格式化错误: 测试", fmt.Sprintf("%v", err))
	})
}

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
