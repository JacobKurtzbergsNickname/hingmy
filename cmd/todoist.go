package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"hingmy/internal/todoist"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// ---- parent command -------------------------------------------------------

var todoistCmd = &cobra.Command{
	Use:   "todoist",
	Short: "Sync wi' yer Todoist account",
	Long:  `Fetch, create, complete, or delete tasks in yer Todoist account usin' the Sync API.`,
}

// ---- shared helpers -------------------------------------------------------

// buildTodoistService resolves an API token and constructs a Service.
// It prefers TODOIST_API_TOKEN from the environment; if absent it falls back
// to a stored OAuth access token. Returns a user-facing error if neither exists.
func buildTodoistService() (*todoist.Service, error) {
	apiToken := os.Getenv("TODOIST_API_TOKEN")

	if apiToken == "" {
		store, err := todoist.NewAccessTokenStore("")
		if err == nil {
			apiToken, _ = store.Load()
		}
	}

	if apiToken == "" {
		return nil, fmt.Errorf(
			"nae Todoist token found — run `hingmy todoist auth login` or set TODOIST_API_TOKEN in .env",
		)
	}

	client := todoist.NewClient(apiToken, nil)
	tokenStore, err := todoist.NewFileSyncTokenStore("")
	if err != nil {
		return nil, fmt.Errorf("todoist: setting up sync token store: %w", err)
	}
	return todoist.NewService(client, tokenStore), nil
}

// selectTodoistItem presents an interactive selector and returns the chosen item.
// It mirrors selectTodo() in interactive.go: option strings are "[<id>] Content (due: date)".
func selectTodoistItem(items []todoist.Item, prompt string) (*todoist.Item, error) {
	options := make([]string, len(items))
	for i, item := range items {
		due := ""
		if item.Due != nil && item.Due.Date != "" {
			due = " (due: " + item.Due.Date + ")"
		}
		options[i] = fmt.Sprintf("[%s] %s%s", item.ID, item.Content, due)
	}

	selectedStr, err := pterm.DefaultInteractiveSelect.WithOptions(options).Show(prompt)
	if err != nil {
		pterm.Warning.Println("Nae selection made.")
		return nil, err
	}

	// Extract the ID from "[<id>] ..."
	trimmed := strings.TrimPrefix(selectedStr, "[")
	id := trimmed[:strings.Index(trimmed, "]")]
	for i := range items {
		if items[i].ID == id {
			return &items[i], nil
		}
	}
	return nil, fmt.Errorf("todoist: selected item not found")
}

// ---- todoist list ---------------------------------------------------------

var todoistListCmd = &cobra.Command{
	Use:   "list",
	Short: "List yer incomplete Todoist tasks",
	Long:  `Fetch incomplete tasks fae Todoist and display them in a braw wee table.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := buildTodoistService()
		if err != nil {
			pterm.Error.Println(err)
			return err
		}

		spinner, _ := pterm.DefaultSpinner.Start("Fetchin' yer tasks fae Todoist...")
		items, err := svc.ListItems(context.Background())
		spinner.Stop() //nolint:errcheck
		if err != nil {
			pterm.Error.Println("Och nae! Couldnae fetch yer Todoist tasks:", err)
			return err
		}

		if len(items) == 0 {
			pterm.Info.Println("Nae incomplete tasks — ye're aw caught up, ya legend!")
			return nil
		}

		tableData := pterm.TableData{{"ID", "Content", "Description", "Due"}}
		for _, item := range items {
			due := ""
			if item.Due != nil {
				due = item.Due.Date
			}
			tableData = append(tableData, []string{item.ID, item.Content, item.Description, due})
		}
		return pterm.DefaultTable.WithHasHeader().WithBoxed().WithData(tableData).Render()
	},
}

// ---- todoist create -------------------------------------------------------

var (
	todoistCreateTitle string
	todoistCreateDesc  string
	todoistCreateDue   string
)

var todoistCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new Todoist task",
	Long:  `Send a new task tae Todoist. Use -t for the title, -d for a description, -u for a due date.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if strings.TrimSpace(todoistCreateTitle) == "" {
			return fmt.Errorf("ye need tae pass a title wi' -t")
		}

		svc, err := buildTodoistService()
		if err != nil {
			pterm.Error.Println(err)
			return err
		}

		spinner, _ := pterm.DefaultSpinner.Start("Sendin' yer task tae Todoist...")
		id, err := svc.CreateItem(
			context.Background(),
			strings.TrimSpace(todoistCreateTitle),
			strings.TrimSpace(todoistCreateDesc),
			strings.TrimSpace(todoistCreateDue),
		)
		spinner.Stop() //nolint:errcheck
		if err != nil {
			pterm.Error.Println("Och nae! Couldnae create yer Todoist task:", err)
			return err
		}

		pterm.Success.Printf("Weel done! Task '%s' added tae Todoist (ID: %s)\n", todoistCreateTitle, id)
		return nil
	},
}

// ---- todoist complete -----------------------------------------------------

var todoistCompleteCmd = &cobra.Command{
	Use:   "complete",
	Short: "Mark a Todoist task as done",
	Long:  `Interactively pick an incomplete task fae Todoist and mark it as done.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := buildTodoistService()
		if err != nil {
			pterm.Error.Println(err)
			return err
		}

		spinner, _ := pterm.DefaultSpinner.Start("Fetchin' yer tasks fae Todoist...")
		items, err := svc.ListItems(context.Background())
		spinner.Stop() //nolint:errcheck
		if err != nil {
			pterm.Error.Println("Couldnae fetch tasks:", err)
			return err
		}

		if len(items) == 0 {
			pterm.Info.Println("Nae incomplete tasks tae complete — job's a good 'un already!")
			return nil
		}

		selected, err := selectTodoistItem(items, "Which task wid ye like tae complete?")
		if err != nil {
			return nil // user cancelled the selector
		}

		spinner2, _ := pterm.DefaultSpinner.Start("Marking it done...")
		err = svc.CompleteItem(context.Background(), selected.ID)
		spinner2.Stop() //nolint:errcheck
		if err != nil {
			pterm.Error.Println("Couldnae complete the task:", err)
			return err
		}

		pterm.Success.Printf("Aye, '%s' is done and dusted!\n", selected.Content)
		return nil
	},
}

// ---- todoist delete -------------------------------------------------------

var todoistDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a Todoist task",
	Long:  `Interactively pick a task fae Todoist and delete it permanently.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := buildTodoistService()
		if err != nil {
			pterm.Error.Println(err)
			return err
		}

		spinner, _ := pterm.DefaultSpinner.Start("Fetchin' yer tasks fae Todoist...")
		items, err := svc.ListItems(context.Background())
		spinner.Stop() //nolint:errcheck
		if err != nil {
			pterm.Error.Println("Couldnae fetch tasks:", err)
			return err
		}

		if len(items) == 0 {
			pterm.Info.Println("Nae tasks tae delete — yer Todoist is spotless!")
			return nil
		}

		selected, err := selectTodoistItem(items, "Which task wid ye like tae delete?")
		if err != nil {
			return nil // user cancelled
		}

		confirm, _ := pterm.DefaultInteractiveConfirm.
			WithDefaultText(fmt.Sprintf("Delete '%s'? This cannae be undone!", selected.Content)).
			Show()
		if !confirm {
			pterm.Info.Println("Och, fair enough — leavin' it be.")
			return nil
		}

		spinner2, _ := pterm.DefaultSpinner.Start("Deletin' the task...")
		err = svc.DeleteItem(context.Background(), selected.ID)
		spinner2.Stop() //nolint:errcheck
		if err != nil {
			pterm.Error.Println("Couldnae delete the task:", err)
			return err
		}

		pterm.Success.Printf("Aye, '%s' has been banished fae Todoist!\n", selected.Content)
		return nil
	},
}

// ---- todoist auth (sub-group) ---------------------------------------------

var todoistAuthCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage yer Todoist authentication",
	Long:  `Log intae Todoist via OAuth, or remove yer stored credentials.`,
}

var todoistAuthLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log intae Todoist via OAuth",
	Long: `Opens yer browser tae authorise hingmy wi' Todoist.
Requires TODOIST_CLIENT_ID and TODOIST_CLIENT_SECRET in .env.
Alternatively, skip this and set TODOIST_API_TOKEN in .env directly.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := todoist.OAuthConfig{
			ClientID:     os.Getenv("TODOIST_CLIENT_ID"),
			ClientSecret: os.Getenv("TODOIST_CLIENT_SECRET"),
		}

		store, err := todoist.NewAccessTokenStore("")
		if err != nil {
			pterm.Error.Println("Couldnae set up the token store:", err)
			return err
		}

		if err := todoist.RunOAuthFlow(cfg, store); err != nil {
			pterm.Error.Println("Och nae! OAuth login failed:", err)
			return err
		}

		pterm.Success.Println("Aye, ye're connected tae Todoist! Run `hingmy todoist list` tae get started.")
		return nil
	},
}

var todoistAuthLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove yer stored Todoist token",
	Long:  `Deletes the locally stored Todoist OAuth token. Doesnae revoke it on Todoist's side.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := todoist.NewAccessTokenStore("")
		if err != nil {
			pterm.Error.Println("Couldnae set up the token store:", err)
			return err
		}

		if err := store.Delete(); err != nil {
			pterm.Error.Println("Couldnae remove the token:", err)
			return err
		}

		pterm.Info.Println("Todoist token removed. Ye'll need tae log in again tae use `hingmy todoist`.")
		return nil
	},
}

// ---- self-registration (mirrors cmd/auth.go) ------------------------------

func init() {
	rootCmd.AddCommand(todoistCmd)

	todoistCmd.AddCommand(todoistListCmd)
	todoistCmd.AddCommand(todoistCreateCmd)
	todoistCmd.AddCommand(todoistCompleteCmd)
	todoistCmd.AddCommand(todoistDeleteCmd)

	todoistCmd.AddCommand(todoistAuthCmd)
	todoistAuthCmd.AddCommand(todoistAuthLoginCmd)
	todoistAuthCmd.AddCommand(todoistAuthLogoutCmd)

	todoistCreateCmd.Flags().StringVarP(&todoistCreateTitle, "title", "t", "", "Task title (required)")
	todoistCreateCmd.Flags().StringVarP(&todoistCreateDesc, "description", "d", "", "Task description (optional)")
	todoistCreateCmd.Flags().StringVarP(&todoistCreateDue, "due", "u", "", "Due date, e.g. 'tomorrow' or '2026-05-20' (optional)")
}
