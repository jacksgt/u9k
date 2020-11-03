package misc

// extrated from https://github.com/emirozer/go-helpers/
func StringInSlice(str string, slice []string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}
