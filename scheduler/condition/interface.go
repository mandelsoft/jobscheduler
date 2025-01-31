package condition

type State struct {
	Enabled bool
	Final   bool
	Valid   bool
}

func (s State) GetState() State {
	return s
}

type Condition interface {
	GetState() State

	// IsEnabled returns whether the condition enabled successive elements.
	IsEnabled() bool

	Evaluate(Event)

	Walk(Walker) bool
}

type Walker interface {
	Walk(t Condition) bool
}

type WalkerFunc func(t Condition) bool

func (f WalkerFunc) Walk(t Condition) bool {
	return f(t)
}

type Event interface {
}
