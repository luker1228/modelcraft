package bizutils

import (
	"fmt"

	"github.com/gofrs/uuid"
)

// GenerateUUIDV7 生成UUIDV7（天然有序的分布式ID），如果数据库有特殊需求采用其它id，需要特殊标记
func GenerateUUIDV7() (string, error) {
	v7, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("GenerateUUIDV7 Fail %w", err)
	}
	return v7.String(), nil
}
