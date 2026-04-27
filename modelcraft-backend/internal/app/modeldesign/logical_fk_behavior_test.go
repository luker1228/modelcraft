package modeldesign

import (
	"context"
	"errors"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/project"
	"modelcraft/pkg/bizerrors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateSystemEndUserRefFK_CreatesUndeletableUnidirectionalFK(t *testing.T) {
	ctx := context.Background()
	fkRepo := new(MockLogicalForeignKeyRepository)
	modelRepo := new(MockModelRepository)

	svc := &LogicalFKAppService{fkRepo: fkRepo, modelRepo: modelRepo}
	ownerField := &modeldesign.FieldDefinition{
		Name: "owner",
		Type: modeldesign.GetFieldTypeByFormat(modeldesign.FormatEndUserRef),
	}
	m := &modeldesign.DataModel{
		ModelMeta: modeldesign.ModelMeta{
			ID: "m1",
			ModelLocator: modeldesign.ModelLocator{
				ProjectScope: project.ProjectScope{OrgName: "org1", ProjectSlug: "myproject"},
				ModelName:    "orders",
				DatabaseName: "private_myproject",
			},
		},
		Fields: []*modeldesign.FieldDefinition{ownerField},
	}

	fkRepo.On("Save", ctx, mock.MatchedBy(func(lf *modeldesign.LogicalForeignKey) bool {
		return lf.Direction == modeldesign.DirectionNormal &&
			lf.RefModelName == "end_user_users" &&
			lf.RefTableName == "end_user_users" &&
			lf.RefDatabaseName == "mc_meta" &&
			!lf.IsDeletable
	})).Return(nil)
	fkRepo.On("BindBelongsToFields", ctx, "org1", "m1", mock.AnythingOfType("string"), []string{"owner"}).Return(nil)

	err := svc.CreateSystemEndUserRefFK(ctx, "org1", m)
	assert.NoError(t, err)
	assert.NotNil(t, ownerField.BelongsToFKID)
	fkRepo.AssertExpectations(t)
}

func TestDeleteLogicalForeignKey_RejectsUndeletable(t *testing.T) {
	ctx := context.Background()
	fkRepo := new(MockLogicalForeignKeyRepository)
	modelRepo := new(MockModelRepository)
	svc := &LogicalFKAppService{fkRepo: fkRepo, modelRepo: modelRepo}

	fkRepo.On("FindByPairID", ctx, "org1", "pair-1").Return([]*modeldesign.LogicalForeignKey{{
		ID:           "lf1",
		PairID:       "pair-1",
		OrgName:      "org1",
		Direction:    modeldesign.DirectionNormal,
		ModelID:      "m1",
		ModelName:    "orders",
		RefModelName: "end_user_users",
		RefTableName: "end_user_users",
		SourceFields: []string{"owner"},
		TargetFields: []string{"id"},
		IsDeletable:  false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}}, nil)

	err := svc.DeleteLogicalForeignKey(ctx, DeleteLogicalForeignKeyCommand{OrgName: "org1", PairID: "pair-1"})
	assert.Error(t, err)
	var b *bizerrors.BusinessError
	assert.True(t, errors.As(err, &b))
	assert.Equal(t, bizerrors.FKNotDeletable.GetCode(), b.Info().GetCode())
}
