// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package review

import (
	"database/sql"
	"time"
)

type Capture struct {
	ID          int64
	UserID      sql.NullInt64
	Name        string
	Description sql.NullString
	Type        string
	HasCab      sql.NullBool
	Am2Hash     string
	DataHash    string
	Data        []byte
	Downloads   int64
	DemoLink    sql.NullString
	CreatedAt   time.Time
	UpdatedAt   sql.NullTime
}

type ProtectedAm2 struct {
	Am2Hash   string
	Ref       string
	CreatedAt time.Time
}

type Review struct {
	ID        int64
	UserID    sql.NullInt64
	CaptureID int64
	Rate      int64
	Comment   sql.NullString
	CreatedAt time.Time
	UpdatedAt sql.NullTime
}

type User struct {
	ID        int64
	Login     string
	Email     string
	Pass      string
	Status    string
	CreatedAt time.Time
	UpdatedAt sql.NullTime
}

type UserFavorite struct {
	UserID    int64
	CaptureID int64
	CreatedAt time.Time
}
