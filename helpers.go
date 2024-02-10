package goht

// If returns trueValue if condition is true, otherwise falseValue.
func If(condition bool, trueValue, falseValue string) string {
	if condition {
		return trueValue
	}
	return falseValue
}
