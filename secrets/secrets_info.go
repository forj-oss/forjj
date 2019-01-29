package secrets

type secretInfo struct {
	keyPath string
	value string
	source string
	env string
	found bool
}

