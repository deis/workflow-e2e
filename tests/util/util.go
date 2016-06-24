package util

// PrependError adds 'Error: ' to an expected error, like the CLI does to error messages.
func PrependError(expected error) string {
	return "Error: " + expected.Error()
}
