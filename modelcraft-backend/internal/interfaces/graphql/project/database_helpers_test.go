package projectgraphql

import (
	"modelcraft/internal/interfaces/graphql/project/generated"
	"testing"

	domainmodeldatabase "modelcraft/internal/domain/modeldatabase"

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
