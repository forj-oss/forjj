package git

type GitCmdInterface interface {
	Get(opts ...string) (string, error)
	GetWithStatusCode(opts ...string) (string, int)
	Do(opts ...string) int
}
