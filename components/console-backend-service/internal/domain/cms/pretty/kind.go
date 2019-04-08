package pretty

type Kind int

const (
	DocsTopic Kind = iota
	DocsTopics
	ClusterDocsTopic
	ClusterDocsTopics
)

func (k Kind) String() string {
	switch k {
	case DocsTopic:
		return "Docs Topic"
	case DocsTopics:
		return "Docs Topics"
	case ClusterDocsTopic:
		return "Cluster Docs Topic"
	case ClusterDocsTopics:
		return "Cluster Docs Topics"
	default:
		return ""
	}
}
