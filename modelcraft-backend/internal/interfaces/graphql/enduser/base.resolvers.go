package endusergraphql

import "modelcraft/internal/interfaces/graphql/enduser/generated"

// Query returns the QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver {
	return &queryResolver{r}
}

type queryResolver struct{ *Resolver }
