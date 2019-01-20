package git

type GitCmdInterface interface {
	get(opts ...string) (string, error)
	getWithStatusCode(opts ...string) (string, int)
	do(opts ...string) int
}
