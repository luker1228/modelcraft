package logfacade

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
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

func makePlainErr() error {
	return errors.New("plain std error")
}

// TestErrorfWithErr 验证：传入 err 时输出 "error" 结构化字段
func TestErrorfWithErr(t *testing.T) {
	err := fmt.Errorf("db connection timeout: failed to init db")

	content := captureLog(t, func(l Logger) {
		l.Errorf(context.Background(), err, "op failed: %v", err)
	})

	if !strings.Contains(content, "op failed") {
		t.Errorf("expected 'op failed' in log, got:\n%s", content)
	}

	if !strings.Contains(content, `"error"`) {
		t.Errorf("expected structured 'error' field, got:\n%s", content)
	}
}

// TestErrorfAutoStackOnlyErrorField 验证：err 只输出 error 字段
func TestErrorfAutoStackOnlyErrorField(t *testing.T) {
	err := fmt.Errorf("db connection timeout: failed to init db")

	content := captureLog(t, func(l Logger) {
		l.Errorf(context.Background(), err, "op failed")
	})

	if !strings.Contains(content, `"error"`) {
		t.Errorf("expected 'error' field, got:\n%s", content)
	}
}

// TestErrorfAutoStackPlainErr 验证：传入标准库 error，只输出 error 字段，不输出 stack 字段
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
