package ui

import "fmt"

// PlainUI outputs plain text with no ANSI escape codes. Selected when CI=true
// or --no-color is passed so the output is safe for piped or log consumption.
type PlainUI struct{}

// NewPlainUI creates a PlainUI.
func NewPlainUI() *PlainUI { return &PlainUI{} }

type plainSpinner struct{ text string }

func (p *plainSpinner) Stop() error {
	fmt.Printf("done: %s\n", p.text)
	return nil
}

func (u *PlainUI) Spinner(text string) Spinner {
	fmt.Printf("... %s\n", text)
	return &plainSpinner{text: text}
}

func (u *PlainUI) Info(text string) {
	fmt.Printf("[info] %s\n", text)
}

func (u *PlainUI) Warning(text string) {
	fmt.Printf("[warn] %s\n", text)
}

func (u *PlainUI) Box(title, body string) {
	fmt.Printf("[%s]\n%s\n", title, body)
}

func (u *PlainUI) Table(rows [][]string) {
	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				fmt.Print("\t")
			}
			fmt.Print(cell)
		}
		fmt.Println()
	}
}

func (u *PlainUI) Section(title string) {
	fmt.Printf("=== %s ===\n", title)
}
