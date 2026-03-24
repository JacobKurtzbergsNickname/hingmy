package ui

// Spinner is a running progress indicator that can be stopped.
type Spinner interface {
	Stop() error
}

// UI is the interface all terminal output flows through. Auth and other
// packages call only these methods — they import no pterm packages directly.
type UI interface {
	Spinner(text string) Spinner
	Info(text string)
	Warning(text string)
	Box(title, body string)
	Table(rows [][]string)
	Section(title string)
}
