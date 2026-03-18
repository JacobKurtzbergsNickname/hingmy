package cmd

import (
	"os"
	"strings"
	"time"

	"hingmy/database"

	"github.com/joho/godotenv"
	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "hingmy",
	Short: "A Scottish-themed todo manager",
	Long:  `Hingmy — yer wee Scottish todo manager. Run it bare tae get the full interactive experience.`,

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if _, err := database.CreateIfNotExists("DB_PATH"); err != nil {
			pterm.Error.Printf("Database setup failed: %v\n", err)
			os.Exit(1)
		}
		if _, err := database.RunManualMigrations(); err != nil {
			pterm.Error.Printf("Migration failed: %v\n", err)
			os.Exit(1)
		}
	},

	Run: func(cmd *cobra.Command, args []string) {
		WelcomeHingmy()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	godotenv.Load()
}

const boxWidth = 100

func introScroll(area *pterm.AreaPrinter, texts []string, delay time.Duration) error {
	for _, text := range texts {
		updateHandle, err := big(text)
		if err != nil {
			return err
		}
		area.Update(updateHandle)
		time.Sleep(delay)
	}
	return nil
}

func WelcomeHingmy() {
	area, err := pterm.DefaultArea.Start()
	if err != nil {
		pterm.Error.Println("Failed to start area:", err)
		return
	}

	welcomeTexts := []string{
		"Welcome tae Hingmy!",
		"And noo",
		"And noo.",
		"And noo..",
		"And noo...",
		"...yer gonnae",
		"have tae dae it",
		"Now fae the DB",
	}

	updateHandle, err := big(welcomeTexts[0])
	if err != nil {
		pterm.Error.Println("Failed to render big text:", err)
		return
	}
	area.Update(updateHandle)
	time.Sleep(2 * time.Second)

	if err := introScroll(area, welcomeTexts[1:], 500*time.Millisecond); err != nil {
		pterm.Error.Println("Failed to play animation:", err)
		return
	}

	time.Sleep(1 * time.Second)
	area.Clear()

	boxContent := createBox("Areet, checking yer database status...", pterm.FgBlue, boxWidth)
	area.Update(boxContent)

	wasDatabaseCreated, err := database.CreateIfNotExists("DB_PATH")
	if err != nil {
		area.Update(createBox("Och nae! Database error: "+err.Error(), pterm.FgRed, boxWidth, "Error"))
		time.Sleep(2 * time.Second)
		area.Stop()
		return
	}

	if wasDatabaseCreated {
		boxContent = createBox("Aye, there's nocht here, fuck all...", pterm.FgYellow, boxWidth)
		time.Sleep(1 * time.Second)
		area.Update(boxContent)
		boxContent = createBox("Ah suppose, ah'll dae it, then...", pterm.FgYellow, boxWidth)
	} else {
		boxContent = createBox("Database found, nae worries.", pterm.FgGreen, boxWidth)
	}
	time.Sleep(1 * time.Second)
	area.Update(boxContent)

	time.Sleep(1 * time.Second)

	area.Update(createBox("Trying tae get the DB reddit up", pterm.FgCyan, boxWidth))

	ranMigrations, err := database.RunManualMigrations()
	if err != nil {
		area.Update(createBox("Och nae! Migration error: "+err.Error(), pterm.FgRed, boxWidth, "Error"))
		time.Sleep(2 * time.Second)
		area.Stop()
		return
	}

	if ranMigrations {
		boxContent = createBox("Aye, had tae update a few things, nae problem...", pterm.FgYellow, boxWidth)
	} else {
		boxContent = createBox("Database is up tae date, bonnie and braw.", pterm.FgGreen, boxWidth)
	}
	time.Sleep(1 * time.Second)
	area.Update(boxContent)

	time.Sleep(1 * time.Second)
	area.Update(createBox("Aye, ye're aw set!", pterm.FgGreen, boxWidth, "Success"))
	time.Sleep(1 * time.Second)

	area.Stop()

	RunInteractiveMode()
}

func big(text string) (string, error) {
	return pterm.DefaultBigText.WithLetters(putils.LettersFromString(text)).Srender()
}

func createBox(message string, color pterm.Color, boxWidth int, title ...string) string {
	box := pterm.DefaultBox.
		WithRightPadding(2).
		WithLeftPadding(2).
		WithTopPadding(1).
		WithBottomPadding(1).
		WithBoxStyle(pterm.NewStyle(color))

	if len(title) > 0 {
		box = box.WithTitle(title[0])
	}

	return box.Sprint(padMessage(message, boxWidth-4))
}

func padMessage(msg string, messageWidth int) string {
	if len(msg) <= messageWidth {
		return msg + strings.Repeat(" ", messageWidth-len(msg))
	}

	var lines []string
	for len(msg) > messageWidth {
		breakPoint := messageWidth
		for i := messageWidth - 1; i >= messageWidth-20 && i > 0; i-- {
			if msg[i] == ' ' {
				breakPoint = i
				break
			}
		}
		lines = append(lines, padLine(msg[:breakPoint], messageWidth))
		msg = msg[breakPoint:]
		if len(msg) > 0 && msg[0] == ' ' {
			msg = msg[1:]
		}
	}
	if len(msg) > 0 {
		lines = append(lines, padLine(msg, messageWidth))
	}
	return strings.Join(lines, "\n")
}

func padLine(line string, width int) string {
	if len(line) >= width {
		return line
	}
	return line + strings.Repeat(" ", width-len(line))
}
