package app_context

type WithApp interface {
	App() Context
}

type WithAppBase struct {
	AppInterface Context
}

func (w *WithAppBase) App() Context {
	return w.AppInterface
}
