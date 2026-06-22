package soargs

type Cmd struct {
	Lines   int
	Columns int
	IsAtty  bool
	Env     map[string]string
	Args    []string
}
