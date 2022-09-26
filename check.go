package main

func checkMatch(s string) (m bool, c int) {
	for _, val := range compiledPatterns {
		if val.MatchString(s) {
			c++
		}
	}
	if c < 1 {
		return false, c
	}
	return true, c
}
