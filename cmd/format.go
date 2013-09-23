package cmd

// Bold generates ansi-escaped bold text.
func Bold(str string) string {
	return "\033[1m" + str + "\033[0m"
}

// Blue generates ansi-escaped blue text.
func Blue(str string) string {
	return Bold("\033[34m" + str + "\033[0m")
}
