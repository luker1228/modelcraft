package orggraphql

import "modelcraft/internal/interfaces/graphql/org/generated"

func emptyProjectAuthSchema() *generated.ProjectAuthSchema {
	return &generated.ProjectAuthSchema{Variables: []*generated.AuthVariable{}}
}
