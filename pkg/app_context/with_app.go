package app_context

type WithApp interface {
	App() Context
}

type WithAppBase struct {
	app Context
}

func (w *WithAppBase) Init(app Context) {
	w.app = app
}

func (w *WithAppBase) App() Context {
	return w.app
}
