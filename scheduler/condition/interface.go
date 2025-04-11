package condition

type State struct {
	Enabled bool
	Final   bool
	Valid   bool
}

func (s State) GetState() State {
	return s
}

type StateTrigger interface {
	SetStateTrigger(func())
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

func SetStateTrigger(c Condition, t func()) {
	if c == nil {
		return
	}

	f := func(e Condition) bool {
		if s, ok := e.(StateTrigger); ok {
			s.SetStateTrigger(t)
		}
		return true
	}

	c.Walk(WalkerFunc(f))
}
