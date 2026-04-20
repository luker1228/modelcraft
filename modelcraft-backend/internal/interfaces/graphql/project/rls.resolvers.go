package projectgraphql

import (
	"context"
	"errors"
	"fmt"
	"modelcraft/internal/app/rls"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/interfaces/graphql/project/generated"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/ctxutils"
	"strings"

	domainRLS "modelcraft/internal/domain/rls"
)

// SetModelRLSPolicy is the resolver for the setModelRLSPolicy field.
func (r *mutationResolver) SetModelRLSPolicy( //nolint:lll // generated resolver signature
	ctx context.Context, input generated.SetModelRLSPolicyInput,
) (*generated.SetModelRLSPolicyPayload, error) {
	orgName, projectSlug, err := getOrgAndProjectFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Convert GraphQL input to AppService DTO
	appInput := rls.SetModelRLSPolicyInput{
		ModelID:         input.ModelID,
		SelectPredicate: domainRLS.JsonExpr(input.SelectPredicate),
		InsertCheck:     domainRLS.JsonExpr(input.InsertCheck),
		UpdatePredicate: domainRLS.JsonExpr(input.UpdatePredicate),
		UpdateCheck:     domainRLS.JsonExpr(input.UpdateCheck),
		DeletePredicate: domainRLS.JsonExpr(input.DeletePredicate),
	}

	// Call AppService
	policy, err := r.RLSPolicyAppService.SetPolicy(ctx, orgName, projectSlug, appInput)
	if err != nil {
		return &generated.SetModelRLSPolicyPayload{
			Policy: nil,
			Error:  convertRLSErrorToGraphQLType(ctx, err),
		}, nil
	}

	// Convert domain model to GraphQL type
	graphqlPolicy := convertPolicyToGraphQL(policy)

	return &generated.SetModelRLSPolicyPayload{
		Policy: graphqlPolicy,
		Error:  nil,
	}, nil
}

// ValidateRLSExpr is the resolver for the validateRLSExpr field.
func (r *mutationResolver) ValidateRLSExpr(
	ctx context.Context, input generated.ValidateRLSExprInput,
) (*generated.ValidateRLSExprPayload, error) {
	orgName, projectSlug, err := getOrgAndProjectFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Convert exprType
	exprType := convertGraphQLExprTypeToDomain(input.ExprType)

	// Call AppService
	validationErrors := r.RLSPolicyAppService.ValidateExpr(
		ctx, orgName, projectSlug, input.ModelID,
		domainRLS.ExprType(exprType), domainRLS.JsonExpr(input.Expression),
	)

	// Convert validation errors
	graphqlErrors := make([]*generated.ValidationError, 0, len(validationErrors))
	for _, e := range validationErrors {
		graphqlErrors = append(graphqlErrors, &generated.ValidationError{
			Path:    e.Path,
			Message: e.Message,
			Code:    e.Code,
		})
	}

	return &generated.ValidateRLSExprPayload{
		Result: &generated.ValidationResult{
			Valid:  len(validationErrors) == 0,
			Errors: graphqlErrors,
		},
	}, nil
}

// ModelRLSPolicy is the resolver for the modelRLSPolicy field.
func (r *queryResolver) ModelRLSPolicy(ctx context.Context, modelID string) (*generated.ModelRLSPolicy, error) {
	orgName, projectSlug, err := getOrgAndProjectFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Call AppService
	policy, err := r.RLSPolicyAppService.GetPolicy(ctx, orgName, projectSlug, modelID)
	if err != nil {
		return nil, err
	}

	if policy == nil {
		return nil, nil //nolint:nilnil // nil policy with nil error means no policy exists
	}

	return convertPolicyToGraphQL(policy), nil
}

// Helper to extract orgName and projectSlug from context
func getOrgAndProjectFromContext(ctx context.Context) (orgName, projectSlug string, err error) {
	orgName, err = ctxutils.GetOrgNameFromContext(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to get organization name from context: %w", err)
	}

	projectSlug, err = ctxutils.GetProjectSlugFromContext(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to get project slug from context: %w", err)
	}

	return orgName, projectSlug, nil
}

// convertPolicyToGraphQL converts domain ModelRLSPolicy to GraphQL type
func convertPolicyToGraphQL(policy *modeldesign.ModelRLSPolicy) *generated.ModelRLSPolicy {
	if policy == nil {
		return nil
	}

	// Get preset
	preset := policy.GetPreset()
	var graphqlPreset *generated.RLSPreset
	if preset != nil {
		p := convertPresetToGraphQL(*preset)
		graphqlPreset = &p
	}

	return &generated.ModelRLSPolicy{
		ModelID:         policy.ModelID,
		SelectPredicate: string(policy.SelectPredicate),
		InsertCheck:     string(policy.InsertCheck),
		UpdatePredicate: string(policy.UpdatePredicate),
		UpdateCheck:     string(policy.UpdateCheck),
		DeletePredicate: string(policy.DeletePredicate),
		Preset:          graphqlPreset,
		CreatedAt:       policy.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       policy.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// convertPresetToGraphQL converts domain RLSPreset to GraphQL enum
func convertPresetToGraphQL(preset domainRLS.RLSPreset) generated.RLSPreset {
	switch preset {
	case domainRLS.RLSPresetReadWriteOwner:
		return generated.RLSPresetReadWriteOwner
	case domainRLS.RLSPresetReadAllWriteOwner:
		return generated.RLSPresetReadAllWriteOwner
	case domainRLS.RLSPresetReadAll:
		return generated.RLSPresetReadAll
	case domainRLS.RLSPresetReadWriteAll:
		return generated.RLSPresetReadWriteAll
	case domainRLS.RLSPresetNoAccess:
		return generated.RLSPresetNoAccess
	default:
		return generated.RLSPresetReadWriteOwner
	}
}

// convertGraphQLExprTypeToDomain converts GraphQL RLSExprType to domain ExprType
func convertGraphQLExprTypeToDomain(exprType generated.RLSExprType) string {
	switch exprType {
	case generated.RLSExprTypeSelectPredicate:
		return "SELECT_PREDICATE"
	case generated.RLSExprTypeInsertCheck:
		return "INSERT_CHECK"
	case generated.RLSExprTypeUpdatePredicate:
		return "UPDATE_PREDICATE"
	case generated.RLSExprTypeUpdateCheck:
		return "UPDATE_CHECK"
	case generated.RLSExprTypeDeletePredicate:
		return "DELETE_PREDICATE"
	default:
		return "SELECT_PREDICATE"
	}
}

// convertRLSErrorToGraphQLType converts error to GraphQL error type
func convertRLSErrorToGraphQLType(ctx context.Context, err error) generated.SetModelRLSPolicyError {
	if err == nil {
		return nil
	}

	var bizErr *bizerrors.BusinessError
	if !errors.As(err, &bizErr) {
		return nil
	}

	// Check error code and convert to appropriate GraphQL type
	code := bizErr.Info().GetCode()

	switch code {
	case bizerrors.ModelNotFound.GetCode():
		return &generated.ModelNotFound{
			Message: bizErr.Error(),
		}
	case bizerrors.ParamInvalid.GetCode():
		// Check if it's "no owner field" error
		if bizErr.Error() == "Model has no owner field (EndUserRef), cannot set RLS policy" {
			return &generated.ModelHasNoOwnerField{
				Message: bizErr.Error(),
				Suggestion: func() *string {
					s := "Please add an EndUserRef field named 'owner' to the model first"
					return &s
				}(),
			}
		}
		// Check if it's invalid RLS expression
		if contains(bizErr.Error(), "Invalid RLS expression") {
			return &generated.InvalidRLSExpression{
				Message: bizErr.Error(),
			}
		}
		// Check if it's invalid auth variable
		if contains(bizErr.Error(), "Invalid auth variable") {
			return &generated.InvalidAuthVariable{
				Message: bizErr.Error(),
			}
		}
		// For other ParamInvalid errors, return nil (not a specific GraphQL error type)
		return nil
	default:
		return nil
	}
}

// contains checks if string contains substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// PopulateModelRLSPolicy populates RLS policy for a model (called by ModelMapper)
func (r *Resolver) PopulateModelRLSPolicy(ctx context.Context, model *generated.Model) error {
	if model == nil || model.ID == "" {
		return nil
	}

	orgName, projectSlug, err := getOrgAndProjectFromContext(ctx)
	if err != nil {
		return err
	}

	// Call AppService to get policy
	policy, err := r.RLSPolicyAppService.GetPolicy(ctx, orgName, projectSlug, model.ID)
	if err != nil {
		// If no policy found, set to nil (which is valid)
		model.RlsPolicy = nil
		return nil
	}

	if policy == nil {
		model.RlsPolicy = nil
		return nil
	}

	model.RlsPolicy = convertPolicyToGraphQL(policy)
	return nil
}
