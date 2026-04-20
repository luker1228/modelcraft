package modeldesign

import (
	"context"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/project"
	"modelcraft/pkg/ctxutils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Mock: EnumRepository for IsArray inference tests
// ============================================================================

// MockEnumRepo is a mock implementation of EnumRepository for testing.
type MockEnumRepo struct {
	mock.Mock
}

func (m *MockEnumRepo) Create(enum *modeldesign.EnumDefinition) error {
	args := m.Called(enum)
	return args.Error(0)
}

func (m *MockEnumRepo) Update(enum *modeldesign.EnumDefinition) error {
	args := m.Called(enum)
	return args.Error(0)
}

func (m *MockEnumRepo) Delete(orgName, projectSlug, name string) error {
	args := m.Called(orgName, projectSlug, name)
	return args.Error(0)
}

func (m *MockEnumRepo) FindByName(orgName, projectSlug, name string) (*modeldesign.EnumDefinition, error) {
	args := m.Called(orgName, projectSlug, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*modeldesign.EnumDefinition), args.Error(1)
}

func (m *MockEnumRepo) FindByID(id string) (*modeldesign.EnumDefinition, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*modeldesign.EnumDefinition), args.Error(1)
}

func (m *MockEnumRepo) List(orgName, projectSlug string) ([]*modeldesign.EnumDefinition, error) {
	args := m.Called(orgName, projectSlug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*modeldesign.EnumDefinition), args.Error(1)
}

func (m *MockEnumRepo) IsReferencedByFields(orgName, projectSlug, name string) (bool, []string, error) {
	args := m.Called(orgName, projectSlug, name)
	return args.Bool(0), args.Get(1).([]string), args.Error(2)
}

func (m *MockEnumRepo) ExistsByName(orgName, projectSlug, name string) (bool, error) {
	args := m.Called(orgName, projectSlug, name)
	return args.Bool(0), args.Error(1)
}

// ============================================================================
// Helper Functions for IsArray Tests
// ============================================================================

// newTestContextForIsArray creates a context with required values for testing.
func newTestContextForIsArray() context.Context {
	ctx := context.Background()
	ctx = ctxutils.NewHttpContext(ctx, &ctxutils.HttpRequestContext{
		RequestId: "test-req-id",
		Lang:      "en",
	})
	ctx = ctxutils.SetContextValue(ctx, ctxutils.ContextKeyOrgName, "test-org")
	return ctx
}

// newTestEnumDefinition creates an EnumDefinition for testing.
func newTestEnumDefinition(name string, isMultiSelect bool) *modeldesign.EnumDefinition {
	return &modeldesign.EnumDefinition{
		ID: "enum-" + name,
		ProjectScope: project.ProjectScope{
			OrgName:     "test-org",
			ProjectSlug: "project-1",
		},
		Name:          name,
		DisplayName:   name,
		Description:   "Test enum",
		IsMultiSelect: isMultiSelect,
		Options: []modeldesign.EnumOption{
			{Code: "A", Label: "Option A", Order: 1},
			{Code: "B", Label: "Option B", Order: 2},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// newTestEnumField creates an ENUM format field for testing.
func newTestEnumField(
	modelID, name, enumName string,
	isArray bool,
	locator *modeldesign.ModelLocator,
) *modeldesign.FieldDefinition {
	fieldType := modeldesign.GetFieldTypeByFormat(modeldesign.FormatEnum)
	return &modeldesign.FieldDefinition{
		ModelID:      modelID,
		ModelLocator: locator,
		Name:         name,
		Title:        name,
		Type:         fieldType,
		EnumName:     enumName,
		IsArray:      isArray,
		Status:       modeldesign.FieldStatusInit,
		Metadata:     map[string]any{},
	}
}

// newTestServiceWithEnumRepo creates a service with enum repository injected.
func newTestServiceWithEnumRepo(
	modelRepo modeldesign.ModelRepository,
	deployRepo modeldesign.DeployRepo,
	enumRepo modeldesign.EnumRepository,
	enumAssocRepo modeldesign.FieldEnumAssociationRepository,
) *ModelDesignAppService {
	return NewModelDesignAppService(ModelDesignAppServiceDeps{
		ModelRepo:     modelRepo,
		DeployRepo:    deployRepo,
		EnumRepo:      enumRepo,
		EnumAssocRepo: enumAssocRepo,
	})
}

// ============================================================================
// Tests: IsArray Inference for ENUM fields
// ============================================================================

// TestAddFieldSync_EnumField_InfersIsArrayFromMultiSelect verifies that when adding
// an ENUM field with an associated enum that has IsMultiSelect=true, the field's
// IsArray is automatically set to true.
func TestAddFieldSync_EnumField_InfersIsArrayFromMultiSelect(t *testing.T) {
	ctx := newTestContextForIsArray()
	mockModelRepo := new(MockModelRepository)
	mockDeployRepo := new(MockDeployRepo)
	mockEnumRepo := new(MockEnumRepo)
	mockEnumAssocRepo := new(MockEnumAssocRepo)

	svc := newTestServiceWithEnumRepo(mockModelRepo, mockDeployRepo, mockEnumRepo, mockEnumAssocRepo)

	locator, _ := modeldesign.NewModelLocator("test-org", "project-1", "db_1", "test_model")
	existingModel := newTestModel("model-1", "project-1", "test_model", "db_1")

	// Create an ENUM field without setting IsArray (default false)
	// The enum has IsMultiSelect=true, so IsArray should be inferred as true
	enumField := newTestEnumField("model-1", "status", "status_enum", false, locator)

	// Enum definition with IsMultiSelect=true
	multiSelectEnum := newTestEnumDefinition("status_enum", true)

	// Setup mocks
	mockModelRepo.On("GetByID", ctx, "model-1", mock.Anything).Return(existingModel, nil)
	mockEnumRepo.On("FindByName", "test-org", "project-1", "status_enum").Return(multiSelectEnum, nil)
	mockModelRepo.On("GetTailFieldDisplayOrder", ctx, "model-1").Return("P", nil)

	// Capture the field passed to AddFields to verify IsArray was set
	var capturedFields []*modeldesign.FieldDefinition
	mockModelRepo.On("AddFields", ctx, "test-org", mock.MatchedBy(func(fields []*modeldesign.FieldDefinition) bool {
		capturedFields = fields
		return len(fields) == 1
	})).Return(nil)
	mockDeployRepo.On("DeployModelToAddFields", ctx, existingModel, mock.Anything).Return(nil)
	mockModelRepo.On("UpdateFieldsStatus", ctx, mock.Anything).Return(nil)
	mockEnumAssocRepo.On("Create", ctx, mock.Anything).Return(nil)

	err := svc.AddFieldSync(ctx, AddFieldCommand{
		ModelID: "model-1",
		Fields:  []*modeldesign.FieldDefinition{enumField},
	})

	require.NoError(t, err)
	require.Len(t, capturedFields, 1)
	// The field's IsArray should be inferred from enum.IsMultiSelect
	assert.True(t, capturedFields[0].IsArray, "IsArray should be inferred as true from enum.IsMultiSelect")
	mockEnumRepo.AssertExpectations(t)
}

// TestAddFieldSync_EnumField_SingleSelectEnumKeepsIsArrayFalse verifies that when
// adding an ENUM field with an associated single-select enum, IsArray remains false.
func TestAddFieldSync_EnumField_SingleSelectEnumKeepsIsArrayFalse(t *testing.T) {
	ctx := newTestContextForIsArray()
	mockModelRepo := new(MockModelRepository)
	mockDeployRepo := new(MockDeployRepo)
	mockEnumRepo := new(MockEnumRepo)
	mockEnumAssocRepo := new(MockEnumAssocRepo)

	svc := newTestServiceWithEnumRepo(mockModelRepo, mockDeployRepo, mockEnumRepo, mockEnumAssocRepo)

	locator, _ := modeldesign.NewModelLocator("test-org", "project-1", "db_1", "test_model")
	existingModel := newTestModel("model-1", "project-1", "test_model", "db_1")

	// Create an ENUM field with IsArray=false
	enumField := newTestEnumField("model-1", "priority", "priority_enum", false, locator)

	// Enum definition with IsMultiSelect=false
	singleSelectEnum := newTestEnumDefinition("priority_enum", false)

	// Setup mocks
	mockModelRepo.On("GetByID", ctx, "model-1", mock.Anything).Return(existingModel, nil)
	mockEnumRepo.On("FindByName", "test-org", "project-1", "priority_enum").Return(singleSelectEnum, nil)
	mockModelRepo.On("GetTailFieldDisplayOrder", ctx, "model-1").Return("P", nil)

	var capturedFields []*modeldesign.FieldDefinition
	mockModelRepo.On("AddFields", ctx, "test-org", mock.MatchedBy(func(fields []*modeldesign.FieldDefinition) bool {
		capturedFields = fields
		return len(fields) == 1
	})).Return(nil)
	mockDeployRepo.On("DeployModelToAddFields", ctx, existingModel, mock.Anything).Return(nil)
	mockModelRepo.On("UpdateFieldsStatus", ctx, mock.Anything).Return(nil)
	mockEnumAssocRepo.On("Create", ctx, mock.Anything).Return(nil)

	err := svc.AddFieldSync(ctx, AddFieldCommand{
		ModelID: "model-1",
		Fields:  []*modeldesign.FieldDefinition{enumField},
	})

	require.NoError(t, err)
	require.Len(t, capturedFields, 1)
	// The field's IsArray should remain false for single-select enum
	assert.False(t, capturedFields[0].IsArray, "IsArray should be false for single-select enum")
	mockEnumRepo.AssertExpectations(t)
}

// TestAddFieldSync_EnumField_IsArrayTrueButEnumNotMultiSelect_ReturnsError verifies
// that when a field requests IsArray=true but the enum is not multi-select, an error
// is returned.
func TestAddFieldSync_EnumField_IsArrayTrueButEnumNotMultiSelect_ReturnsError(t *testing.T) {
	ctx := newTestContextForIsArray()
	mockModelRepo := new(MockModelRepository)
	mockDeployRepo := new(MockDeployRepo)
	mockEnumRepo := new(MockEnumRepo)
	mockEnumAssocRepo := new(MockEnumAssocRepo)

	svc := newTestServiceWithEnumRepo(mockModelRepo, mockDeployRepo, mockEnumRepo, mockEnumAssocRepo)

	locator, _ := modeldesign.NewModelLocator("test-org", "project-1", "db_1", "test_model")
	existingModel := newTestModel("model-1", "project-1", "test_model", "db_1")

	// Create an ENUM field with IsArray=true (explicitly requesting multi-select)
	enumField := newTestEnumField("model-1", "category", "category_enum", true, locator)

	// Enum definition with IsMultiSelect=false (single-select only)
	singleSelectEnum := newTestEnumDefinition("category_enum", false)

	// Setup mocks
	mockModelRepo.On("GetByID", ctx, "model-1", mock.Anything).Return(existingModel, nil)
	mockEnumRepo.On("FindByName", "test-org", "project-1", "category_enum").Return(singleSelectEnum, nil)

	err := svc.AddFieldSync(ctx, AddFieldCommand{
		ModelID: "model-1",
		Fields:  []*modeldesign.FieldDefinition{enumField},
	})

	// Should return an error because enum does not support multi-select
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not support multi-select",
		"Error message should indicate enum doesn't support multi-select")
	mockEnumRepo.AssertExpectations(t)
}

// TestAddFieldSync_NonEnumField_IsArrayNotAffected verifies that non-ENUM fields
// preserve their IsArray setting without any modification.
func TestAddFieldSync_NonEnumField_IsArrayNotAffected(t *testing.T) {
	ctx := newTestContextForIsArray()
	mockModelRepo := new(MockModelRepository)
	mockDeployRepo := new(MockDeployRepo)
	mockEnumRepo := new(MockEnumRepo)
	mockEnumAssocRepo := new(MockEnumAssocRepo)

	svc := newTestServiceWithEnumRepo(mockModelRepo, mockDeployRepo, mockEnumRepo, mockEnumAssocRepo)

	locator, _ := modeldesign.NewModelLocator("test-org", "project-1", "db_1", "test_model")
	existingModel := newTestModel("model-1", "project-1", "test_model", "db_1")

	// Create a STRING field (non-ENUM) with IsArray=true
	stringField := newTestField("model-1", "tags", locator)
	stringField.IsArray = true

	// Setup mocks
	mockModelRepo.On("GetByID", ctx, "model-1", mock.Anything).Return(existingModel, nil)
	mockModelRepo.On("GetTailFieldDisplayOrder", ctx, "model-1").Return("P", nil)

	var capturedFields []*modeldesign.FieldDefinition
	mockModelRepo.On("AddFields", ctx, "test-org", mock.MatchedBy(func(fields []*modeldesign.FieldDefinition) bool {
		capturedFields = fields
		return len(fields) == 1
	})).Return(nil)
	mockDeployRepo.On("DeployModelToAddFields", ctx, existingModel, mock.Anything).Return(nil)
	mockModelRepo.On("UpdateFieldsStatus", ctx, mock.Anything).Return(nil)

	err := svc.AddFieldSync(ctx, AddFieldCommand{
		ModelID: "model-1",
		Fields:  []*modeldesign.FieldDefinition{stringField},
	})

	require.NoError(t, err)
	require.Len(t, capturedFields, 1)
	// Non-ENUM field's IsArray should not be modified
	assert.True(t, capturedFields[0].IsArray, "Non-ENUM field's IsArray should not be changed")
	// EnumRepo should NOT be called for non-ENUM fields
	mockEnumRepo.AssertNotCalled(t, "FindByName")
}

// TestAddFieldSync_EnumField_NoEnumName_ReturnsError verifies that ENUM fields
// without EnumName now fail validation per current business rules.
// (Previously this test expected the field to pass without enumName - legacy behavior has been removed.)
func TestAddFieldSync_EnumField_NoEnumName_ReturnsError(t *testing.T) {
	ctx := newTestContextForIsArray()
	mockModelRepo := new(MockModelRepository)
	mockDeployRepo := new(MockDeployRepo)
	mockEnumRepo := new(MockEnumRepo)
	mockEnumAssocRepo := new(MockEnumAssocRepo)

	svc := newTestServiceWithEnumRepo(mockModelRepo, mockDeployRepo, mockEnumRepo, mockEnumAssocRepo)

	locator, _ := modeldesign.NewModelLocator("test-org", "project-1", "db_1", "test_model")
	existingModel := newTestModel("model-1", "project-1", "test_model", "db_1")

	// Create an ENUM field without EnumName
	enumField := newTestEnumField("model-1", "inline_enum", "", false, locator)

	// Setup mocks
	mockModelRepo.On("GetByID", ctx, "model-1", mock.Anything).Return(existingModel, nil)

	err := svc.AddFieldSync(ctx, AddFieldCommand{
		ModelID: "model-1",
		Fields:  []*modeldesign.FieldDefinition{enumField},
	})

	// Current business rule: ENUM fields must have an associated enum
	require.Error(t, err)
	assert.Contains(t, err.Error(), "relateEnumName is required when format=ENUM",
		"Error message should indicate ENUM fields require enumName")
	// EnumRepo should NOT be called since validation fails first
	mockEnumRepo.AssertNotCalled(t, "FindByName")
}
