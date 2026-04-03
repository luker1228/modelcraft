package bizerrors

import (
	stderrors "errors" // 标准库重命名

	pkgerrors "github.com/pkg/errors"
)

// 完全替换标准库 errors
// 对外暴露的接口保持与标准库一致
var (
	New       = pkgerrors.New
	Errorf    = pkgerrors.Errorf
	Wrap      = pkgerrors.Wrap
	Wrapf     = pkgerrors.Wrapf
	WithStack = pkgerrors.WithStack
	Cause     = pkgerrors.Cause
	Is        = pkgerrors.Is
	As        = pkgerrors.As
	Unwrap    = pkgerrors.Unwrap
)

// 禁止使用标准库
var _ = stderrors.New // 仅用于类型检查
