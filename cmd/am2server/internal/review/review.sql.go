// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: review.sql

package review

import (
	"context"
	"database/sql"
)

const listReviewsByUser = `-- name: ListReviewsByUser :many
SELECT id, user_id, capture_id, rate, comment, created_at, updated_at FROM review
WHERE user_id = ?
`

// http: GET /users/{user_id}/reviews
func (q *Queries) ListReviewsByUser(ctx context.Context, userID sql.NullInt64) ([]Review, error) {
	rows, err := q.db.QueryContext(ctx, listReviewsByUser, userID)
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

const addReview = `-- name: addReview :execresult
INSERT INTO review(user_id, capture_id, rate, comment) 
VALUES(?,?,?,?)
`

type addReviewParams struct {
	UserID    sql.NullInt64
	CaptureID int64
	Rate      int64
	Comment   sql.NullString
}

// http: POST /reviews
func (q *Queries) addReview(ctx context.Context, arg addReviewParams) (sql.Result, error) {
	return q.db.ExecContext(ctx, addReview,
		arg.UserID,
		arg.CaptureID,
		arg.Rate,
		arg.Comment,
	)
}

const existsReviewByUserCapture = `-- name: existsReviewByUserCapture :one
SELECT COUNT(*) FROM review WHERE user_id = ? AND capture_id = ?
`

type existsReviewByUserCaptureParams struct {
	UserID    sql.NullInt64
	CaptureID int64
}

func (q *Queries) existsReviewByUserCapture(ctx context.Context, arg existsReviewByUserCaptureParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, existsReviewByUserCapture, arg.UserID, arg.CaptureID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const getReview = `-- name: getReview :one
SELECT id, user_id, capture_id, rate, comment, created_at, updated_at FROM review WHERE id = ?
`

func (q *Queries) getReview(ctx context.Context, id int64) (Review, error) {
	row := q.db.QueryRowContext(ctx, getReview, id)
	var i Review
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.CaptureID,
		&i.Rate,
		&i.Comment,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const removeReview = `-- name: removeReview :execresult
DELETE FROM review WHERE id = ?
`

// http: DELETE /reviews/{id}
func (q *Queries) removeReview(ctx context.Context, id int64) (sql.Result, error) {
	return q.db.ExecContext(ctx, removeReview, id)
}
