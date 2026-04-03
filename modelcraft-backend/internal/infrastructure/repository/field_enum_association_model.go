package repository

import (
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/project"
	"time"
)

// FieldEnumAssociationPO persistence object for model_field_enum_associations table.
type FieldEnumAssociationPO struct {
	ModelID      string
	FieldName    string
	ProjectSlug  string
	EnumName     string
	DatabaseName string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// TableName returns the table name for model_field_enum_associations.
func (FieldEnumAssociationPO) TableName() string {
	return "model_field_enum_associations"
}

// ToFieldEnumAssociation converts the persistence object to a domain entity.
func (po *FieldEnumAssociationPO) ToFieldEnumAssociation() *modeldesign.FieldEnumAssociation {
	return &modeldesign.FieldEnumAssociation{
		ModelID:   po.ModelID,
		FieldName: po.FieldName,
		ProjectScope: project.ProjectScope{
			OrgName:     "",
			ProjectSlug: po.ProjectSlug,
		},
		EnumName:     po.EnumName,
		DatabaseName: po.DatabaseName,
		CreatedAt:    po.CreatedAt,
		UpdatedAt:    po.UpdatedAt,
	}
}

// FromFieldEnumAssociation populates the persistence object from a domain entity.
func (po *FieldEnumAssociationPO) FromFieldEnumAssociation(assoc *modeldesign.FieldEnumAssociation) {
	po.ModelID = assoc.ModelID
	po.FieldName = assoc.FieldName
	po.ProjectSlug = assoc.ProjectSlug
	po.EnumName = assoc.EnumName
	po.DatabaseName = assoc.DatabaseName
	po.CreatedAt = assoc.CreatedAt
	po.UpdatedAt = assoc.UpdatedAt
}

// NewFieldEnumAssociationPOFromDomain creates a FieldEnumAssociationPO from a domain entity.
func NewFieldEnumAssociationPOFromDomain(assoc *modeldesign.FieldEnumAssociation) *FieldEnumAssociationPO {
	po := &FieldEnumAssociationPO{}
	po.FromFieldEnumAssociation(assoc)
	return po
}
