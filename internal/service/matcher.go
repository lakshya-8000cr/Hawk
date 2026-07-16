package service

func selectorMatches(
	selector map[string]string,
	labels map[string]string,
) bool {
	if len(selector) == 0 {
		return false
	}

	for key, expectedValue := range selector {
		actualValue, exists := labels[key]

		if !exists || actualValue != expectedValue {
			return false
		}
	}

	return true
}
