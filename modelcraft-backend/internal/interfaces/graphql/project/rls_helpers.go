package projectgraphql

import (
	"context"
	"errors"
	"fmt"
	rls "modelcraft/internal/app/rls"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/interfaces/graphql/project/generated"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/ctxutils"
	"strings"

	domainRLS "modelcraft/internal/domain/rls"
)

// getOrgAndProjectFromContext extracts orgName and projectSlug from GraphQL context.
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

func convertPolicyToGraphQL(policy *modeldesign.ModelRLSPolicy) *generated.ModelRLSPolicy {
	if policy == nil {
		return nil
	}

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

func convertRLSErrorToGraphQLType(ctx context.Context, err error) generated.SetModelRLSPolicyError {
	if err == nil {
		return nil
	}

	var bizErr *bizerrors.BusinessError
	if !errors.As(err, &bizErr) {
		return nil
	}

	code := bizErr.Info().GetCode()

	switch code {
	case bizerrors.ModelNotFound.GetCode():
		return &generated.ModelNotFound{Message: bizErr.Error()}
	case bizerrors.ParamInvalid.GetCode():
		if bizErr.Error() == "Model has no owner field (EndUserRef), cannot set RLS policy" {
			suggestion := "Please add an EndUserRef field named 'owner' to the model first"
			return &generated.ModelHasNoOwnerField{
				Message:    bizErr.Error(),
				Suggestion: &suggestion,
			}
		}
		if contains(bizErr.Error(), "Invalid RLS expression") {
			return &generated.InvalidRLSExpression{Message: bizErr.Error()}
		}
		if contains(bizErr.Error(), "Invalid auth variable") {
			return &generated.InvalidAuthVariable{Message: bizErr.Error()}
		}
		return nil
	default:
		return nil
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func toDomainAuthVarType(t generated.AuthVariableType) domainRLS.AuthVarType {
	switch t {
	case generated.AuthVariableTypeUUID:
		return domainRLS.AuthVarTypeUUID
	case generated.AuthVariableTypeInteger:
		return domainRLS.AuthVarTypeInteger
	case generated.AuthVariableTypeString:
		fallthrough
	default:
		return domainRLS.AuthVarTypeString
	}
}

func toGraphQLAuthVarType(t domainRLS.AuthVarType) generated.AuthVariableType {
	switch t {
	case domainRLS.AuthVarTypeUUID:
		return generated.AuthVariableTypeUUID
	case domainRLS.AuthVarTypeInteger:
		return generated.AuthVariableTypeInteger
	case domainRLS.AuthVarTypeString:
		fallthrough
	default:
		return generated.AuthVariableTypeString
	}
}

func toGraphQLProjectAuthSchema(authSchema *domainRLS.AuthSchema) *generated.ProjectAuthSchema {
	if authSchema == nil {
		return &generated.ProjectAuthSchema{Variables: []*generated.AuthVariable{}}
	}

	variables := make([]*generated.AuthVariable, 0, len(authSchema.Variables))
	for _, v := range authSchema.Variables {
		variables = append(variables, &generated.AuthVariable{
			Name:   v.Name,
			Source: v.Source,
			Type:   toGraphQLAuthVarType(v.Type),
		})
	}

	return &generated.ProjectAuthSchema{Variables: variables}
}

func attachModelRLSPolicy(
	ctx context.Context,
	policySvc *rls.ModelRLSPolicyAppService,
	model *generated.Model,
) error {
	if model == nil || policySvc == nil {
		return nil
	}

	orgName, projectSlug, err := getOrgAndProjectFromContext(ctx)
	if err != nil {
		return err
	}

	policy, err := policySvc.GetPolicy(ctx, orgName, projectSlug, model.ID)
	if err != nil {
		return err
	}
	model.RlsPolicy = convertPolicyToGraphQL(policy)
	return nil
}
