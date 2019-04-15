package fixture

import "github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared"

func ClusterDocsTopic(name string) shared.ClusterDocsTopic {
	return shared.ClusterDocsTopic{
		Name:        name,
		GroupName:   DocsTopicGroupName,
		DisplayName: DocsTopicDisplayName,
		Description: DocsTopicDescription,
	}
}
