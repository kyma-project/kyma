package controller

type labelled interface {
	Labels() map[string]string
}

func detectLabelsConflicts(source labelled, labels map[string]string) ([]string, bool) {
	dLabels := source.Labels()
	if dLabels == nil {
		return nil, false
	}
	var conflicts []string
	for k := range labels {
		if _, exists := dLabels[k]; exists {
			conflicts = append(conflicts, k)
		}
	}

	return conflicts, len(conflicts) != 0
}
