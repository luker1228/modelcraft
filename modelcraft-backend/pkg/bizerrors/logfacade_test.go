package bizerrors

import (
	"context"
	"fmt"
	"modelcraft/pkg/logfacade"
	"os"
	"strings"
	"testing"
)

func captureLogToFile(t *testing.T, fn func(l logfacade.Logger)) string {
	t.Helper()

	logDir := "./test_logs_biz"
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	defer os.RemoveAll(logDir)

	logFile := logDir + "/biz.log"

	logger, err := logfacade.New(logfacade.Config{
		Level:      logfacade.DebugLevel,
		OutputPath: logFile,
	})
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

func TestErrorfWithBizError(t *testing.T) {
	t.Run("logs bizerrors.BusinessError fields", func(t *testing.T) {
		bizErr := NewError(NotFound)

		content := captureLogToFile(t, func(l logfacade.Logger) {
			l.Errorf(context.Background(), bizErr, "operation failed")
		})

		if !strings.Contains(content, "operation failed") {
			t.Errorf("expected msg in log, got:\n%s", content)
		}

		if !strings.Contains(content, bizErr.Error()) {
			t.Errorf("expected bizErr.Error() in log, got:\n%s", content)
		}

		if !strings.Contains(content, `"error"`) {
			t.Errorf("expected 'error' field in log, got:\n%s", content)
		}
	})

	t.Run("stdout visual check", func(t *testing.T) {
		logger, err := logfacade.New(logfacade.Config{
			Level:      logfacade.DebugLevel,
			OutputPath: "stdout",
		})
		if err != nil {
			t.Fatal(err)
		}

		fmt.Println("=== stdout visual check: bizerrors + logfacade integration ===")

		bizErr := NewError(NotFound)
		logger.Errorf(context.Background(), bizErr, "operation failed")

		inner := Wrap(fmt.Errorf("db connection timeout"), "query failed")
		bizErrWithStack := WrapError(inner, SystemError)
		logger.Errorf(context.Background(), bizErrWithStack, "business error with inner stack")

		bizErrWithParams := NewError(ParamInvalid, "user_id", "must be positive")
		logger.Errorf(context.Background(), bizErrWithParams, "validation error")

		fmt.Println("=== end ===")

		_ = logger.Sync()
	})
}
