package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"hingmy/database"
	"hingmy/database/sqlc"

	"github.com/pterm/pterm"
)

const (
	actionRead   = "Read yer Hingmies"
	actionCreate = "Add a new Hingmy"
	actionUpdate = "Update a Hingmy"
	actionDelete = "Delete a Hingmy"
	actionExit   = "Och, ah'm done"
)

func getAccessor() (*database.Accessor, error) {
	accessor, err := database.NewAccessor()
	if err != nil {
		pterm.Error.Println("Och nae! Couldnae connect tae the database:", err)
	}
	return accessor, err
}

func RunInteractiveMode() {
	accessor, err := database.NewAccessor()
	if err != nil {
		pterm.Error.Println("Och nae! Couldnae connect tae the database:", err)
		return
	}

	pterm.DefaultHeader.WithFullWidth().Println("Hingmy — Yer Scottish Todo Manager")

	for {
		pterm.Println()
		action, err := pterm.DefaultInteractiveSelect.
			WithOptions([]string{
				actionRead,
				actionCreate,
				actionUpdate,
				actionDelete,
				actionExit,
			}).
			Show("Whit wid ye like tae dae?")
		if err != nil {
			pterm.Error.Println("Och nae!", err)
			return
		}

		pterm.Println()

		switch action {
		case actionRead:
			doRead(accessor)
		case actionCreate:
			doCreate(accessor)
		case actionUpdate:
			doUpdate(accessor)
		case actionDelete:
			doDelete(accessor)
		case actionExit:
			pterm.Success.Println("Aye, cheerio then! Haste ye back!")
			return
		}
	}
}

func doRead(accessor *database.Accessor) {
	todos, err := accessor.GetAllTodos()
	if err != nil {
		pterm.Error.Println("Couldnae fetch yer hingmies:", err)
		return
	}

	if len(todos) == 0 {
		pterm.Warning.Println("Ye've got nae hingmies! Add one first, pal.")
		return
	}

	tableData := pterm.TableData{
		{"ID", "Title", "Description", "Due Date", "Done"},
	}
	for _, t := range todos {
		done := "Naw"
		if t.Completed.Bool {
			done = "Aye"
		}
		due := ""
		if t.DueDate.Valid {
			due = t.DueDate.Time.Format("2006-01-02")
		}
		tableData = append(tableData, []string{
			strconv.FormatInt(t.ID, 10),
			t.Title,
			t.Description.String,
			due,
			done,
		})
	}

	pterm.DefaultTable.WithHasHeader().WithBoxed().WithData(tableData).Render()
}

func doCreate(accessor *database.Accessor) {
	pterm.DefaultSection.Println("Add a New Hingmy")

	title, err := pterm.DefaultInteractiveTextInput.Show("Title (required)")
	if err != nil || strings.TrimSpace(title) == "" {
		pterm.Warning.Println("Och, ye need a title at least, pal!")
		return
	}

	description, err := pterm.DefaultInteractiveTextInput.Show("Description (optional, hit Enter tae skip)")
	if err != nil {
		description = ""
	}

	dueDate, err := pterm.DefaultInteractiveTextInput.WithDefaultText("").Show("Due date (YYYY-MM-DD, optional)")
	if err != nil {
		dueDate = ""
	}

	todo, err := accessor.CreateTodo(strings.TrimSpace(title), strings.TrimSpace(description), strings.TrimSpace(dueDate))
	if err != nil {
		pterm.Error.Println("Couldnae create yer hingmy:", err)
		return
	}

	pterm.Success.Printf("Weel done! Hingmy #%d '%s' added tae yer list!\n", todo.ID, todo.Title)
}

func doUpdate(accessor *database.Accessor) {
	todos, err := accessor.GetAllTodos()
	if err != nil {
		pterm.Error.Println("Couldnae fetch yer hingmies:", err)
		return
	}

	if len(todos) == 0 {
		pterm.Warning.Println("Ye've got nae hingmies tae update!")
		return
	}

	selected, err := selectTodo(todos, "Which hingmy wid ye like tae update?")
	if err != nil {
		return
	}

	pterm.DefaultSection.Printf("Updating Hingmy #%d — '%s'\n", selected.ID, selected.Title)
	pterm.Info.Println("(Hit Enter tae keep the current value)")
	pterm.Println()

	newTitle, err := pterm.DefaultInteractiveTextInput.WithDefaultText(selected.Title).Show("Title")
	if err != nil || strings.TrimSpace(newTitle) == "" {
		newTitle = selected.Title
	}

	newDesc, err := pterm.DefaultInteractiveTextInput.WithDefaultText(selected.Description.String).Show("Description")
	if err != nil {
		newDesc = selected.Description.String
	}

	currentDue := ""
	if selected.DueDate.Valid {
		currentDue = selected.DueDate.Time.Format("2006-01-02")
	}
	newDue, err := pterm.DefaultInteractiveTextInput.WithDefaultText(currentDue).Show("Due date (YYYY-MM-DD)")
	if err != nil {
		newDue = currentDue
	}

	completedStr := "Naw"
	if selected.Completed.Bool {
		completedStr = "Aye"
	}
	newCompletedStr, err := pterm.DefaultInteractiveSelect.
		WithOptions([]string{"Aye", "Naw"}).
		WithDefaultOption(completedStr).
		Show("Completed?")
	if err != nil {
		newCompletedStr = completedStr
	}

	completed := newCompletedStr == "Aye"

	err = accessor.UpdateTodo(selected.ID, strings.TrimSpace(newTitle), strings.TrimSpace(newDesc), strings.TrimSpace(newDue), completed)
	if err != nil {
		pterm.Error.Println("Couldnae update yer hingmy:", err)
		return
	}

	pterm.Success.Printf("Braw! Hingmy #%d updated nae bother.\n", selected.ID)
}

func doDelete(accessor *database.Accessor) {
	todos, err := accessor.GetAllTodos()
	if err != nil {
		pterm.Error.Println("Couldnae fetch yer hingmies:", err)
		return
	}

	if len(todos) == 0 {
		pterm.Warning.Println("Ye've got nae hingmies tae delete!")
		return
	}

	selected, err := selectTodo(todos, "Which hingmy wid ye like tae delete?")
	if err != nil {
		return
	}

	confirm, err := pterm.DefaultInteractiveConfirm.
		WithDefaultText(fmt.Sprintf("Delete hingmy #%d '%s'? This cannae be undone!", selected.ID, selected.Title)).
		Show()
	if err != nil || !confirm {
		pterm.Info.Println("Och, fair enough — leaving it be.")
		return
	}

	err = accessor.SoftDeleteTodo(selected.ID)
	if err != nil {
		pterm.Error.Println("Couldnae delete yer hingmy:", err)
		return
	}

	pterm.Success.Printf("Aye, hingmy #%d '%s' has been banished!\n", selected.ID, selected.Title)
}

func selectTodo(todos []sqlc.Todo, prompt string) (*sqlc.Todo, error) {
	options := make([]string, len(todos))
	for i, t := range todos {
		done := " "
		if t.Completed.Bool {
			done = "✓"
		}
		due := ""
		if t.DueDate.Valid {
			due = " (due: " + t.DueDate.Time.Format("2006-01-02") + ")"
		}
		options[i] = fmt.Sprintf("[%d] %s [%s]%s", t.ID, t.Title, done, due)
	}

	selectedStr, err := pterm.DefaultInteractiveSelect.WithOptions(options).Show(prompt)
	if err != nil {
		pterm.Warning.Println("Nae selection made.")
		return nil, err
	}

	// Parse the ID from the selected string "[ID] ..."
	idStr := strings.TrimPrefix(selectedStr, "[")
	idStr = idStr[:strings.Index(idStr, "]")]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		pterm.Error.Println("Couldnae parse the hingmy ID:", err)
		return nil, err
	}

	for i := range todos {
		if todos[i].ID == id {
			return &todos[i], nil
		}
	}

	pterm.Error.Println("Couldnae find that hingmy!")
	return nil, fmt.Errorf("todo not found")
}
