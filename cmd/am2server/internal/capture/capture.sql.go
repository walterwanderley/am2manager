// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: capture.sql

package capture

import (
	"context"
	"database/sql"
	"time"
)

const getCapture = `-- name: GetCapture :one
SELECT id, user_id, name, description, type, has_cab, am2_hash, data_hash, data, downloads, created_at, updated_at FROM capture WHERE id = ?
`

// http: GET /captures/{id}
func (q *Queries) GetCapture(ctx context.Context, id int64) (Capture, error) {
	row := q.db.QueryRowContext(ctx, getCapture, id)
	var i Capture
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Name,
		&i.Description,
		&i.Type,
		&i.HasCab,
		&i.Am2Hash,
		&i.DataHash,
		&i.Data,
		&i.Downloads,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getCaptureFile = `-- name: GetCaptureFile :one
UPDATE capture SET downloads = downloads + 1 WHERE id = ?
RETURNING data, name
`

type GetCaptureFileRow struct {
	Data []byte
	Name string
}

// http: GET /captures/{id}/file
func (q *Queries) GetCaptureFile(ctx context.Context, id int64) (GetCaptureFileRow, error) {
	row := q.db.QueryRowContext(ctx, getCaptureFile, id)
	var i GetCaptureFileRow
	err := row.Scan(&i.Data, &i.Name)
	return i, err
}

const removeCapture = `-- name: RemoveCapture :execresult
DELETE FROM capture WHERE id = ?
`

// http: DELETE /captures/{id}
func (q *Queries) RemoveCapture(ctx context.Context, id int64) (sql.Result, error) {
	return q.db.ExecContext(ctx, removeCapture, id)
}

const searchCaptures = `-- name: SearchCaptures :many
SELECT c.id, c.name, c.description, c.downloads, count(f.capture_id) AS fav, c.has_cab, c.type, c.created_at 
FROM capture c LEFT OUTER JOIN user_favorite f ON c.id = f.capture_id
WHERE c.description LIKE '%'||?1||'%' OR c.name LIKE '%'||?1||'%' 
GROUP BY f.capture_id
ORDER BY c.downloads, c.created_at DESC
LIMIT ?3 OFFSET ?2
`

type SearchCapturesParams struct {
	Arg    sql.NullString
	Offset int64
	Limit  int64
}

type SearchCapturesRow struct {
	ID          int64
	Name        string
	Description sql.NullString
	Downloads   int64
	Fav         int64
	HasCab      sql.NullBool
	Type        string
	CreatedAt   time.Time
}

// http: GET /captures
func (q *Queries) SearchCaptures(ctx context.Context, arg SearchCapturesParams) ([]SearchCapturesRow, error) {
	rows, err := q.db.QueryContext(ctx, searchCaptures, arg.Arg, arg.Offset, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SearchCapturesRow
	for rows.Next() {
		var i SearchCapturesRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.Downloads,
			&i.Fav,
			&i.HasCab,
			&i.Type,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const addCapture = `-- name: addCapture :execresult
INSERT INTO capture(user_id, name, description, type, has_cab, data, am2_hash, data_hash)
VALUES(?,?,?,?,?,?,?,?)
`

type addCaptureParams struct {
	UserID      sql.NullInt64
	Name        string
	Description sql.NullString
	Type        string
	HasCab      sql.NullBool
	Data        []byte
	Am2Hash     string
	DataHash    string
}

func (q *Queries) addCapture(ctx context.Context, arg addCaptureParams) (sql.Result, error) {
	return q.db.ExecContext(ctx, addCapture,
		arg.UserID,
		arg.Name,
		arg.Description,
		arg.Type,
		arg.HasCab,
		arg.Data,
		arg.Am2Hash,
		arg.DataHash,
	)
}