package main

func StringInSlice(a string, s []string) bool {
	for _, v := range s {
		if v == a {
			return true
		}
	}
	return false
}
