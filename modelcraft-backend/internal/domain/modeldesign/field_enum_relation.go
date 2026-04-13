package modeldesign

import (
	"time"

	"modelcraft/internal/domain/project"
	bizerrors "modelcraft/pkg/bizerrors"
)

// FieldEnumRelation 表示 ENUM_LABEL 到 ENUM 源字段的稳定绑定关系。
type FieldEnumRelation struct {
	ID              string `json:"id"`
	ModelID         string `json:"modelId"`
	LabelFieldName  string `json:"labelFieldName"`
	SourceFieldName string `json:"sourceFieldName"`
	project.ProjectScope
	EnumName  string    `json:"enumName"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (r *FieldEnumRelation) Validate() error {
	if r.ID == "" {
		return bizerrors.NewError(bizerrors.ParamInvalid, "field enum relation id cannot be empty")
	}
	if r.ModelID == "" {
		return bizerrors.NewError(bizerrors.ParamInvalid, "field enum relation modelId cannot be empty")
	}
	if r.LabelFieldName == "" {
		return bizerrors.NewError(bizerrors.ParamInvalid, "field enum relation labelFieldName cannot be empty")
	}
	if r.SourceFieldName == "" {
		return bizerrors.NewError(bizerrors.ParamInvalid, "field enum relation sourceFieldName cannot be empty")
	}
	if err := r.ProjectScope.Validate(); err != nil {
		return err
	}
	if r.EnumName == "" {
		return bizerrors.NewError(bizerrors.ParamInvalid, "field enum relation enumName cannot be empty")
	}
	return nil
}
