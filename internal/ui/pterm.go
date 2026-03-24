package ui

import (
	"github.com/pterm/pterm"
)

// PtermUI is the rich terminal implementation of UI using pterm components.
type PtermUI struct{}

// NewPtermUI creates a PtermUI.
func NewPtermUI() *PtermUI { return &PtermUI{} }

type ptermSpinner struct {
	s *pterm.SpinnerPrinter
}

func (p *ptermSpinner) Stop() error {
	return p.s.Stop()
}

// Spinner starts a pterm spinner with a braille sequence.
func (u *PtermUI) Spinner(text string) Spinner {
	s, _ := pterm.DefaultSpinner.
		WithSequence("⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷").
		Start(text)
	return &ptermSpinner{s: s}
}

// Info prints an informational message.
func (u *PtermUI) Info(text string) {
	pterm.Info.Println(text)
}

// Warning prints a warning message.
func (u *PtermUI) Warning(text string) {
	pterm.Warning.Println(text)
}

// Box renders a persistent titled box. Unlike spinners it does not self-clear.
func (u *PtermUI) Box(title, body string) {
	pterm.DefaultBox.WithTitle(title).Println(body)
}

// Table renders a table from a slice of rows. The first row is treated as the header.
func (u *PtermUI) Table(rows [][]string) {
	if len(rows) == 0 {
		return
	}
	data := make(pterm.TableData, len(rows))
	for i, r := range rows {
		data[i] = r
	}
	pterm.DefaultTable.WithHasHeader().WithBoxed().WithData(data).Render() //nolint:errcheck
}

// Section prints a section header.
func (u *PtermUI) Section(title string) {
	pterm.DefaultSection.Println(title)
}
