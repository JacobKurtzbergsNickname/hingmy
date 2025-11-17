-- Todos CRUD operations

-- name: GetTodo :one
SELECT * FROM todos WHERE id = ? AND deleted_at IS NULL;

-- name: ListTodos :many
SELECT * FROM todos WHERE deleted_at IS NULL ORDER BY created_at DESC;

-- name: ListActiveTodos :many
SELECT * FROM todos WHERE completed = FALSE AND deleted_at IS NULL ORDER BY created_at DESC;

-- name: ListCompletedTodos :many
SELECT * FROM todos WHERE completed = TRUE AND deleted_at IS NULL ORDER BY updated_at DESC;

-- name: CreateTodo :one
INSERT INTO todos (title, description, due_date, completed, created_at, updated_at, deleted_at) 
VALUES (?, ?, ?, ?, ?, ?, ?) 
RETURNING *;

-- name: UpdateTodo :exec
UPDATE todos 
SET title = ?, description = ?, completed = ?, due_date = ?, updated_at = CURRENT_TIMESTAMP 
WHERE id = ? AND deleted_at IS NULL;

-- name: CompleteTodo :exec
UPDATE todos 
SET completed = TRUE, updated_at = CURRENT_TIMESTAMP 
WHERE id = ? AND deleted_at IS NULL;

-- name: UncompleteTodo :exec
UPDATE todos 
SET completed = FALSE, updated_at = CURRENT_TIMESTAMP 
WHERE id = ? AND deleted_at IS NULL;

-- name: SoftDeleteTodo :exec
UPDATE todos 
SET deleted_at = CURRENT_TIMESTAMP 
WHERE id = ?;

-- name: DeleteTodo :exec
DELETE FROM todos WHERE id = ?;