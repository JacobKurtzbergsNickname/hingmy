-- Tags CRUD operations

-- name: GetTag :one
SELECT * FROM tags WHERE id = ? AND deleted_at IS NULL;

-- name: GetTagByName :one
SELECT * FROM tags WHERE name = ? AND deleted_at IS NULL;

-- name: ListTags :many
SELECT * FROM tags WHERE deleted_at IS NULL ORDER BY name;

-- name: CreateTag :one
INSERT INTO tags (name) 
VALUES (?) 
RETURNING *;

-- name: UpdateTag :exec
UPDATE tags 
SET name = ?, updated_at = CURRENT_TIMESTAMP 
WHERE id = ? AND deleted_at IS NULL;

-- name: SoftDeleteTag :exec
UPDATE tags 
SET deleted_at = CURRENT_TIMESTAMP 
WHERE id = ?;

-- name: DeleteTag :exec
DELETE FROM tags WHERE id = ?;

-- name: GetTagsForTodo :many
SELECT t.* FROM tags t
JOIN tag_entities te ON t.id = te.tag_id
WHERE te.todo_id = ? AND t.deleted_at IS NULL AND te.deleted_at IS NULL
ORDER BY t.name;