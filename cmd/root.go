/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"time"

	"taedae/database"

	"github.com/joho/godotenv"
	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "taedae",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,

	Run: func(cmd *cobra.Command, args []string) {
		WelcomeTaeTaeD()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Load environment variables from .env file
	godotenv.Load()

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.taedae.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// Define consistent box dimensions
const boxWidth = 100

// IntroScroll plays a sequence of big text animations
func IntroScroll(area *pterm.AreaPrinter, texts []string, delay time.Duration) error {
	for _, text := range texts {
		updateHandle, err := Big(text)
		if err != nil {
			return err
		}
		area.Update(updateHandle)
		time.Sleep(delay)
	}
	return nil
}

func WelcomeTaeTaeD() {
	area, err := pterm.DefaultArea.Start()
	if err != nil {
		pterm.Error.Println("Failed to start area:", err)
		return
	}

	// Welcome sequence
	welcomeTexts := []string{
		"Welcome tae TaeDae!",
		"And noo",
		"And noo.",
		"And noo..",
		"And noo...",
		"...yer gonnae",
		"have tae dae it",
		"Now fae the DB",
	}

	// Play welcome animation (first text gets 2 seconds, rest get 500ms)
	updateHandle, err := Big(welcomeTexts[0])
	if err != nil {
		pterm.Error.Println("Failed to render big text:", err)
		return
	}
	area.Update(updateHandle)
	time.Sleep(2 * time.Second)

	// Play the rest of the sequence
	if err := IntroScroll(area, welcomeTexts[1:], 500*time.Millisecond); err != nil {
		pterm.Error.Println("Failed to play animation:", err)
		return
	}

	time.Sleep(1 * time.Second) // Final pause
	area.Clear()

	// Blue box for database check
	boxContent := createBox("Areet, checking yer database status...", pterm.FgBlue, boxWidth)
	area.Update(boxContent)

	// Database check logic
	wasDatabaseCreated, err := database.CreateIfNotExists("DB_PATH")
	if err != nil {
		errorBox := createBox("Och nae! Database error: "+err.Error(), pterm.FgRed, boxWidth, "Error")
		area.Update(errorBox)
		time.Sleep(2 * time.Second)
		area.Stop()
		return
	}

	if wasDatabaseCreated {
		// Yellow box for database creation
		boxContent = createBox(
			"Aye, there's nocht here, fuck all...",
			pterm.FgYellow,
			boxWidth,
		)
		time.Sleep(1 * time.Second)
		area.Update(boxContent)

		// Of course, you're getting flag from a Scottish guy...
		boxContent = createBox(
			"Ah suppose, ah'll dae it, then...",
			pterm.FgYellow,
			boxWidth,
		)
	} else {
		// Green box for database found
		boxContent = createBox(
			"Database found, nae worries.",
			pterm.FgGreen, boxWidth,
		)
	}
	// And a dramatic... PAUSE!
	time.Sleep(1 * time.Second)
	area.Update(boxContent)

	// And another dramatic... PAUSE!
	time.Sleep(1 * time.Second)

	// Cyan box for migration setup
	boxContent = createBox("Trying tae get the DB reddit up", pterm.FgCyan, boxWidth)
	area.Update(boxContent)

	ranMigrations, err := database.RunManualMigrations()
	if err != nil {
		errorBox := createBox("Och nae! Migration error: "+err.Error(), pterm.FgRed, boxWidth, "Error")
		area.Update(errorBox)
		time.Sleep(2 * time.Second)
		area.Stop()
		return
	}

	if ranMigrations {
		// Yellow box for migrations run
		boxContent = createBox(
			"Aye, had tae update a few things, nae problem...",
			pterm.FgYellow,
			boxWidth,
		)
	} else {
		// Green box for no migrations needed
		boxContent = createBox(
			"Database is up tae date, bonnie and braw.",
			pterm.FgGreen,
			boxWidth,
		)
	}
	time.Sleep(1 * time.Second)
	area.Update(boxContent)

	// Green box for success
	time.Sleep(1 * time.Second)
	boxContent = createBox("Aye, ye're aw set!", pterm.FgGreen, boxWidth, "Success")
	area.Update(boxContent)
	time.Sleep(1 * time.Second)

	area.Stop()
}

// Big creates and returns big text render handle
func Big(text string) (string, error) {
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
	// If message fits on one line, pad it normally
	if len(msg) <= messageWidth {
		for len(msg) < messageWidth {
			msg += " "
		}
		return msg
	}

	// Break long messages into multiple lines
	var lines []string
	for len(msg) > messageWidth {
		// Find the best place to break (prefer spaces)
		breakPoint := messageWidth
		for i := messageWidth - 1; i >= messageWidth-20 && i > 0; i-- {
			if msg[i] == ' ' {
				breakPoint = i
				break
			}
		}

		lines = append(lines, padLine(msg[:breakPoint], messageWidth))
		msg = msg[breakPoint:]

		// Remove leading space if we broke at a space
		if len(msg) > 0 && msg[0] == ' ' {
			msg = msg[1:]
		}
	}
	// Add the remaining text as the last line
	if len(msg) > 0 {
		lines = append(lines, padLine(msg, messageWidth))
	}
	return joinLines(lines)
}

func padLine(line string, width int) string {
	// Pad a single line to the specified width
	for len(line) < width {
		line += " "
	}
	return line
}

func joinLines(lines []string) string {
	// Join lines with newlines
	result := ""
	for i, line := range lines {
		if i > 0 {
			result += "\n"
		}
		result += line
	}
	return result
}
