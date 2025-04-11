package processors

func NotFunc(f func() bool) func() bool {
	return func() bool { return !f() }
}
