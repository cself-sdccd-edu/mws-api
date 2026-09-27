package api

func validTerm(term string) bool {
	if len(term) != 4 {
		return false
	}

	for _, char := range term {
		if char < '0' || char > '9' {
			return false
		}
	}
	// only use this if we truly want to validate terms as 3/5/7 (spring, summer, fall)
	switch term[3] {
	case '3', '5', '7':
		return true
	default:
		return false
	}
}
