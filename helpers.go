package goht

// If returns the trueValue if the condition is true, otherwise falseValue.
func If(condition bool, trueValue, falseValue string) string {
	if condition {
		return trueValue
	}
	return falseValue
}
