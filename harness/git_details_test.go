package harness

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetEnvironmentFilePath_GitX(t *testing.T) {
	path := GetEnvironmentFilePath(true, "", Project{
		Identifier:    "pId",
		OrgIdentifier: "orgId",
	}, EnvironmentClass{
		Identifier: "envId",
		Type:       "PreProduction",
	})
	assert.Equal(t, ".harness/orgs/orgId/projects/pId/envs/pre_production/envId.yaml", path)
}

func Test_GetInfrastructureFilePath_GitX(t *testing.T) {
	path := GetInfrastructureFilePath(true, "", Project{
		Identifier:    "pId",
		OrgIdentifier: "orgId",
	}, EnvironmentClass{
		Identifier: "envId",
		Type:       "PreProduction",
	}, Infrastructure{
		Identifier: "infraId",
	})
	assert.Equal(t, ".harness/orgs/orgId/projects/pId/envs/pre_production/envId/infras/infraId.yaml", path)
}
