package services

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	unique := make([]string, 0, len(values))

	for _, value := range values {
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		unique = append(unique, value)
	}

	return unique
}

/*
reconcileList Desired must be the list from Kubernetes manifest, the current must be the list of Quay
*/
func reconcileList(desired, current []string) (toAdd, toRemove []string) {
	currentSet := make(map[string]struct{}, len(current))

	for _, item := range current {
		currentSet[item] = struct{}{}
	}

	desiredSet := make(map[string]struct{}, len(desired))

	for _, item := range desired {
		desiredSet[item] = struct{}{}
	}

	// Present in desired but not in current
	for item := range desiredSet {
		if _, exists := currentSet[item]; !exists {
			toAdd = append(toAdd, item)
		}
	}

	// Present in current but not in desired
	for item := range currentSet {
		if _, exists := desiredSet[item]; !exists {
			toRemove = append(toRemove, item)
		}
	}
	return
}
