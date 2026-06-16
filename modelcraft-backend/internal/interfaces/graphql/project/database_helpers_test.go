package projectgraphql

import (
	"testing"

	domainmodeldatabase "modelcraft/internal/domain/modeldatabase"
	"modelcraft/internal/interfaces/graphql/project/generated"

	"github.com/stretchr/testify/assert"
)

func TestModelSyncJobToGQL_MapsStatusToGraphQLEnum(t *testing.T) {
	job := &domainmodeldatabase.ModelSyncJob{
		ID:     "job-1",
		Status: domainmodeldatabase.ModelSyncJobStatusSucceeded,
	}

	result := modelSyncJobToGQL(job)

	assert.Equal(t, generated.ModelSyncJobStatusSucceeded, result.Status)
}
