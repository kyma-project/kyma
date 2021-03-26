package app

type App struct {
	busolaURL string
}

func New(busolaURL string) App {
	return App{
		busolaURL: busolaURL,
	}
}
