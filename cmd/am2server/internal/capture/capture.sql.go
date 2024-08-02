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
SELECT id, user_id, name, description, type, has_cab, am2_hash, data_hash, data, downloads, demo_link, created_at, updated_at FROM capture WHERE id = ?
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
		&i.DemoLink,
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

const searchCaptures = `-- name: SearchCaptures :many
SELECT c.id, c.name, c.description, c.downloads, c.has_cab, c.type, c.created_at, c.demo_link, AVG(r.rate) rate
FROM capture c LEFT OUTER JOIN review r ON (c.id = r.capture_id)
WHERE c.description LIKE '%'||?1||'%' OR c.name LIKE '%'||?1||'%' 
OR c.data_hash = ?1 OR c.am2_hash = ?1
GROUP BY c.id
ORDER BY rate DESC, c.downloads DESC
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
	HasCab      sql.NullBool
	Type        string
	CreatedAt   time.Time
	DemoLink    sql.NullString
	Rate        sql.NullFloat64
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
			&i.HasCab,
			&i.Type,
			&i.CreatedAt,
			&i.DemoLink,
			&i.Rate,
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
INSERT INTO capture(user_id, name, description, type, has_cab, data, am2_hash, data_hash, demo_link)
VALUES(?,?,?,?,?,?,?,?,?)
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
	DemoLink    sql.NullString
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
		arg.DemoLink,
	)
}

const listReviewsByCapture = `-- name: listReviewsByCapture :many
SELECT id, user_id, capture_id, rate, comment, created_at, updated_at FROM review
WHERE capture_id = ?
`

func (q *Queries) listReviewsByCapture(ctx context.Context, captureID int64) ([]Review, error) {
	rows, err := q.db.QueryContext(ctx, listReviewsByCapture, captureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Review
	for rows.Next() {
		var i Review
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.CaptureID,
			&i.Rate,
			&i.Comment,
			&i.CreatedAt,
			&i.UpdatedAt,
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

const mostDownloadedCaptures = `-- name: mostDownloadedCaptures :many
SELECT c.id, c.name, c.description, c.downloads, c.has_cab, c.type, c.created_at 
FROM capture c
ORDER BY c.downloads DESC
LIMIT 5
`

type mostDownloadedCapturesRow struct {
	ID          int64
	Name        string
	Description sql.NullString
	Downloads   int64
	HasCab      sql.NullBool
	Type        string
	CreatedAt   time.Time
}

func (q *Queries) mostDownloadedCaptures(ctx context.Context) ([]mostDownloadedCapturesRow, error) {
	rows, err := q.db.QueryContext(ctx, mostDownloadedCaptures)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []mostDownloadedCapturesRow
	for rows.Next() {
		var i mostDownloadedCapturesRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.Downloads,
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

const mostRecentCaptures = `-- name: mostRecentCaptures :many
SELECT c.id, c.name, c.description, c.downloads, c.has_cab, c.type, c.created_at 
FROM capture c
ORDER BY c.created_at DESC
LIMIT 5
`

type mostRecentCapturesRow struct {
	ID          int64
	Name        string
	Description sql.NullString
	Downloads   int64
	HasCab      sql.NullBool
	Type        string
	CreatedAt   time.Time
}

func (q *Queries) mostRecentCaptures(ctx context.Context) ([]mostRecentCapturesRow, error) {
	rows, err := q.db.QueryContext(ctx, mostRecentCaptures)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []mostRecentCapturesRow
	for rows.Next() {
		var i mostRecentCapturesRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.Downloads,
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

const protectedTrainer = `-- name: protectedTrainer :one
SELECT am2_hash, ref, created_at FROM protected_am2 WHERE am2_hash = ? LIMIT 1
`

func (q *Queries) protectedTrainer(ctx context.Context, am2Hash string) (ProtectedAm2, error) {
	row := q.db.QueryRowContext(ctx, protectedTrainer, am2Hash)
	var i ProtectedAm2
	err := row.Scan(&i.Am2Hash, &i.Ref, &i.CreatedAt)
	return i, err
}

const rateByCapture = `-- name: rateByCapture :one
SELECT AVG(rate) FROM review WHERE capture_id = ?
`

func (q *Queries) rateByCapture(ctx context.Context, captureID int64) (sql.NullFloat64, error) {
	row := q.db.QueryRowContext(ctx, rateByCapture, captureID)
	var avg sql.NullFloat64
	err := row.Scan(&avg)
	return avg, err
}

const totalSearchCaptures = `-- name: totalSearchCaptures :one
SELECT count(*)
FROM capture c
WHERE c.description LIKE '%'||?1||'%' OR c.name LIKE '%'||?1||'%' 
OR c.data_hash = ?1 OR c.am2_hash = ?1
`

func (q *Queries) totalSearchCaptures(ctx context.Context, arg sql.NullString) (int64, error) {
	row := q.db.QueryRowContext(ctx, totalSearchCaptures, arg)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const updateCapture = `-- name: updateCapture :execresult
UPDATE capture SET name = ?, 
description = ?, type = ?, has_cab = ?, demo_link = ?
WHERE id = ?
`

type updateCaptureParams struct {
	Name        string
	Description sql.NullString
	Type        string
	HasCab      sql.NullBool
	DemoLink    sql.NullString
	ID          int64
}

// http: PUT /captures/{id}
func (q *Queries) updateCapture(ctx context.Context, arg updateCaptureParams) (sql.Result, error) {
	return q.db.ExecContext(ctx, updateCapture,
		arg.Name,
		arg.Description,
		arg.Type,
		arg.HasCab,
		arg.DemoLink,
		arg.ID,
	)
}
