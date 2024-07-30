package user

import (
	"context"
	"database/sql"
)

type UserRequest struct {
	Email   string
	Name    string
	Picture string
}

func (s *Service) GetOrInsert(ctx context.Context, r UserRequest, db *sql.DB) (*User, error) {

	tx, err := db.BeginTx(ctx, &sql.TxOptions{
		ReadOnly:  false,
		Isolation: sql.LevelReadCommitted,
	})
	defer tx.Rollback()

	if err != nil {
		return nil, err
	}

	queries := New(tx)

	var picture sql.NullString
	if r.Picture != "" {
		picture.Valid = true
		picture.String = r.Picture
	}

	user, _ := queries.GetUserByEmail(ctx, r.Email)
	if user.ID > 0 {
		if picture.Valid {
			_, err := queries.updateUserPicture(ctx, updateUserPictureParams{
				Picture: picture,
				ID:      user.ID,
			})
			if err == nil {
				tx.Commit()
			}
		}
		return &user, nil
	}

	if r.Name == "" {
		r.Name = r.Email
	}

	result, err := queries.AddUser(ctx, AddUserParams{
		Name:    r.Name,
		Email:   r.Email,
		Picture: picture,
		Status:  "VALID",
	})
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	user, err = queries.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &user, err

}
