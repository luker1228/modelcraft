package projectgraphql

import (
	domainmodeldesign "modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/interfaces/graphql/project/generated"
)

// convertFKDirection converts a domain LogicalFKDirection to the generated FKDirection enum.
func convertFKDirection(d domainmodeldesign.LogicalFKDirection) generated.FKDirection {
	if d == domainmodeldesign.DirectionReverse {
		return generated.FKDirectionReverse
	}
	return generated.FKDirectionNormal
}
