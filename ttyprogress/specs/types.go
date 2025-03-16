package specs

// DecoratorFunc is a function that can be prepended and appended to the progress bar
type DecoratorFunc func(b ElementInterface) string
