-- Tag Entities CRUD operations (Many-to-Many relationship between Todos and Tags)

-- name: GetTagEntity :one
SELECT * FROM tag_entities WHERE id = ? AND deleted_at IS NULL;

-- name: GetTagEntityByTodoAndTag :one
SELECT * FROM tag_entities WHERE todo_id = ? AND tag_id = ? AND deleted_at IS NULL;

-- name: ListTagEntities :many
SELECT * FROM tag_entities WHERE deleted_at IS NULL ORDER BY created_at DESC;

-- name: ListTagEntitiesForTodo :many
SELECT * FROM tag_entities WHERE todo_id = ? AND deleted_at IS NULL ORDER BY created_at;

-- name: ListTagEntitiesForTag :many
SELECT * FROM tag_entities WHERE tag_id = ? AND deleted_at IS NULL ORDER BY created_at;

-- name: CreateTagEntity :one
INSERT INTO tag_entities (todo_id, tag_id) 
VALUES (?, ?) 
RETURNING *;

-- name: SoftDeleteTagEntity :exec
UPDATE tag_entities 
SET deleted_at = CURRENT_TIMESTAMP 
WHERE id = ?;

-- name: DeleteTagEntity :exec
DELETE FROM tag_entities WHERE id = ?;

-- name: DeleteTagEntityByTodoAndTag :exec
DELETE FROM tag_entities WHERE todo_id = ? AND tag_id = ?;

-- name: DeleteTagEntitiesForTodo :exec
DELETE FROM tag_entities WHERE todo_id = ?;

-- name: DeleteTagEntitiesForTag :exec
DELETE FROM tag_entities WHERE tag_id = ?;

-- name: AddTagToTodo :one
INSERT OR IGNORE INTO tag_entities (todo_id, tag_id) 
VALUES (?, ?) 
RETURNING *;

-- name: RemoveTagFromTodo :exec
UPDATE tag_entities 
SET deleted_at = CURRENT_TIMESTAMP 
WHERE todo_id = ? AND tag_id = ? AND deleted_at IS NULL;