package logfacade

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	pkgerrors "github.com/pkg/errors"
)

// 用临时文件承载日志输出，返回文件内容字符串
func captureLog(t *testing.T, fn func(l Logger)) string {
	t.Helper()

	logDir := "./test_logs_stack"
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	defer os.RemoveAll(logDir)

	logFile := logDir + "/stack.log"

	logger, err := newZapLogger(Config{
		Level:      DebugLevel,
		OutputPath: logFile,
	}, 1)
	if err != nil {
		t.Fatalf("new logger failed: %v", err)
	}

	fn(logger)

	_ = logger.Sync()
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read log file failed: %v", err)
	}
	return string(content)
}

// makeStackErr 用 pkg/errors 构造带堆栈的 error
func makeStackErr() error {
	return pkgerrors.Wrap(pkgerrors.New("db connection timeout"), "failed to init db")
}

// makePlainErr 构造不带堆栈的标准库 error（不是 pkg/errors）
func makePlainErr() error {
	return errors.New("plain std error")
}

// TestErrorfAutoStackWithStackErr 验证：传入 pkg/errors 包装的 err，
// 即使用 %v 也自动输出 "error" 结构化字段和 "stack" 字段（含堆栈帧）
func TestErrorfAutoStackWithStackErr(t *testing.T) {
	err := makeStackErr()

	content := captureLog(t, func(l Logger) {
		l.Errorf(context.Background(), err, "op failed: %v", err)
	})

	if !strings.Contains(content, "op failed") {
		t.Errorf("expected 'op failed' in log, got:\n%s", content)
	}

	if !strings.Contains(content, `"error"`) {
		t.Errorf("expected structured 'error' field, got:\n%s", content)
	}

	if !strings.Contains(content, `"stack"`) {
		t.Errorf("expected structured 'stack' field, got:\n%s", content)
	}

	// stack 字段应包含多个 .go: 堆栈帧
	stackLineCount := strings.Count(content, ".go:")
	if stackLineCount < 3 {
		t.Errorf("expected >=3 stack frames (.go:), got %d:\n%s", stackLineCount, content)
	}
}

// TestErrorfAutoStackOnlyErrorField 验证：传入带堆栈的 err，
// msg 里不写任何 %v，堆栈仍能输出
func TestErrorfAutoStackOnlyErrorField(t *testing.T) {
	err := makeStackErr()

	content := captureLog(t, func(l Logger) {
		l.Errorf(context.Background(), err, "op failed")
	})

	if !strings.Contains(content, `"error"`) {
		t.Errorf("expected 'error' field, got:\n%s", content)
	}
	if !strings.Contains(content, `"stack"`) {
		t.Errorf("expected 'stack' field, got:\n%s", content)
	}
}

// TestErrorfAutoStackPlainErr 验证：传入标准库 error（无 StackTrace），
// 只输出 error 字段，不输出 stack 字段
func TestErrorfAutoStackPlainErr(t *testing.T) {
	err := makePlainErr()

	content := captureLog(t, func(l Logger) {
		l.Errorf(context.Background(), err, "op failed")
	})

	if !strings.Contains(content, `"error"`) {
		t.Errorf("expected 'error' field, got:\n%s", content)
	}

	if strings.Contains(content, `"stack"`) {
		t.Errorf("plain error should NOT produce 'stack' field, got:\n%s", content)
	}
}

// TestErrorfAutoStackNilErr 验证：err=nil 时既不输出 error 也不输出 stack
func TestErrorfAutoStackNilErr(t *testing.T) {
	content := captureLog(t, func(l Logger) {
		l.Errorf(context.Background(), nil, "just a message: %d", 42)
	})

	if !strings.Contains(content, "just a message: 42") {
		t.Errorf("expected formatted message, got:\n%s", content)
	}

	if strings.Contains(content, `"error"`) {
		t.Errorf("nil err should NOT produce 'error' field, got:\n%s", content)
	}

	if strings.Contains(content, `"stack"`) {
		t.Errorf("nil err should NOT produce 'stack' field, got:\n%s", content)
	}
}
