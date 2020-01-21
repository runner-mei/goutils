package util

func IntsExist(values []int64, n int64) bool {
	for _, value := range values {
		if n == value {
			return true
		}
	}
	return false
}

func IntsDiff(oldValues, newValues []int64) (add, updated, removed []int64) {
	if len(oldValues) == 0 {
		return IntsUnique(newValues), nil, nil
	}
	if len(newValues) == 0 {
		return nil, nil, IntsUnique(oldValues)
	}

	for _, old := range oldValues {
		found := false
		for _, n := range newValues {
			if n == old {
				found = true
				break
			}
		}

		if found {
			if !IntsExist(updated, old) {
				updated = append(updated, old)
			}
		} else {
			if !IntsExist(removed, old) {
				removed = append(removed, old)
			}
		}
	}

	for _, n := range newValues {
		found := false
		for _, old := range oldValues {
			if n == old {
				found = true
				break
			}
		}

		if found {
			continue
		}

		//found = false
		if !IntsExist(add, n) {
			add = append(add, n)
		}
	}

	return
}

func IntsUnique(values []int64) []int64 {
	if len(values) <= 1 {
		return values
	}

	offset := 1
	for i := 1; i < len(values); i++ {
		if IntsExist(values[:i], values[i]) {
			continue
		}

		if i != offset {
			values[offset] = values[i]
		}
		offset++
	}
	return values[:offset]
}
