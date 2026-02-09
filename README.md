# Hingmy - Quick Todo CLI 📝

A Scottish-themed todo CLI application built with Go and Cobra, featuring a charming interface and robust database management.

## Features ✨

- **Scottish Charm**: Enjoy delightful Scottish expressions throughout the interface
- **Animated Welcome**: Beautiful ASCII art welcome sequence
- **Database Management**: Automatic SQLite database creation and migration
- **CRUD Operations**: Full todo management (Create, Read, Update, Delete)
- **Tags & Notes**: Organize todos with tags and add detailed notes
- **Migration System**: Robust database versioning with golang-migrate
- **Environment Config**: Easy configuration via `.env` files

## Installation 🚀

### Prerequisites

- Go 1.21 or later
- Git

### Build from Source

```bash
git clone https://github.com/JacobKurtzbergsNickname/quick_todo.git
cd quick_todo
go mod tidy
go build -o taedae.exe
```

## Configuration ⚙️

Create a `.env` file in the project root:

```env
DB_PATH=.local/share/taedae/database.db
```

The database will be created automatically in your user directory at the specified path.

## Usage 🎯

### Basic Commands

```bash
# Run the application
./taedae

# Create a new todo
./taedae create "Buy groceries" --description "Milk, bread, eggs"

# List all todos  
./taedae read

# Update a todo
./taedae update 1 --title "Buy organic groceries" --completed

# Delete a todo
./taedae delete 1
```

### Database Features

The application automatically:

- ✅ Checks for database existence
- ✅ Creates database if missing
- ✅ Copies migration files to the database location
- ✅ Runs pending migrations
- ✅ Displays progress with animated status boxes

## Project Structure 📁

```folder-structure
quick_todo/
├── cmd/                    # Cobra commands
│   ├── root.go            # Main command & welcome animation
│   ├── create.go          # Create todo command
│   ├── read.go            # Read todos command
│   ├── update.go          # Update todo command
│   └── delete.go          # Delete todo command
├── database/              # Database layer
│   ├── init.db.go         # Database initialization
│   ├── migration.db.go    # Migration management
│   ├── db_utils.go        # Database utilities
│   └── sqlc/              # Generated SQLC code
├── models/                # SQL schema files
│   ├── schema.sql         # Complete database schema
│   ├── todos.sql          # Todos table
│   ├── tags.sql           # Tags table
│   ├── notes.sql          # Notes table
│   └── tag_entities.sql   # Many-to-many relationships
├── queries/               # SQLC query definitions
│   ├── todos.sql          # Todo CRUD operations
│   ├── tags.sql           # Tag CRUD operations
│   ├── notes.sql          # Note CRUD operations
│   └── tag_entities.sql   # Relationship operations
├── migrations/            # Database migration files
│   ├── 001_create_todos_table.up.sql
│   └── 001_create_todos_table.down.sql
├── .env                   # Environment configuration
├── sqlc.yaml             # SQLC configuration
└── go.mod                # Go module definition
```

## Database Schema 🗄️

The application uses SQLite with the following tables:

- **`todos`**: Main todo items with title, description, completion status
- **`tags`**: Reusable labels for organizing todos  
- **`notes`**: Additional notes that can be attached to todos
- **`tag_entities`**: Many-to-many relationship between todos and tags

All tables support soft deletion with `deleted_at` timestamps.

## Development 🛠️

### Generate Database Code

```bash
# Generate Go code from SQL queries
sqlc generate

# Run database migrations
go run . # Migrations run automatically on startup
```

### Dependencies

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [PTerm](https://github.com/pterm/pterm) - Terminal UI library
- [SQLC](https://sqlc.dev/) - Type-safe SQL code generation
- [golang-migrate](https://github.com/golang-migrate/migrate) - Database migration tool
- [godotenv](https://github.com/joho/godotenv) - Environment variable loading

## Contributing 🤝

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Commit changes: `git commit -am 'Add new feature'`
4. Push to branch: `git push origin feature-name`
5. Submit a pull request

## License 📄

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**"Och, yer gonnae love this wee todo app!"** 🏴󐁧󐁢󐁳󐁣󐁴󐁿
