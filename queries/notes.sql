-- Notes CRUD operations

-- name: GetNote :one
SELECT * FROM notes WHERE id = ? AND deleted_at IS NULL;

-- name: ListNotes :many
SELECT * FROM notes WHERE deleted_at IS NULL ORDER BY created_at DESC;

-- name: ListNotesForTodo :many
SELECT * FROM notes WHERE todo_id = ? AND deleted_at IS NULL ORDER BY created_at DESC;

-- name: CreateNote :one
INSERT INTO notes (todo_id, content) 
VALUES (?, ?) 
RETURNING *;

-- name: UpdateNote :exec
UPDATE notes 
SET content = ?, updated_at = CURRENT_TIMESTAMP 
WHERE id = ? AND deleted_at IS NULL;

-- name: SoftDeleteNote :exec
UPDATE notes 
SET deleted_at = CURRENT_TIMESTAMP 
WHERE id = ?;

-- name: DeleteNote :exec
DELETE FROM notes WHERE id = ?;

-- name: DeleteNotesForTodo :exec
DELETE FROM notes WHERE todo_id = ?;