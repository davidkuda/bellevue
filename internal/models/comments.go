package models

import (
	"database/sql"
	"fmt"
	"time"
)

type CommentModel struct {
	DB *sql.DB
}

type Comment struct {
	ID        int
	UserID    int
	Date      time.Time
	Comment   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m *CommentModel) Insert(comment Comment) error {
	stmt := `
	insert into comments (user_id, date, comment)
	values ($1, $2, $3)
	on conflict (user_id, date)
	do update set comment = excluded.comment, updated_at = now();
	`
	if _, err := m.DB.Exec(
		stmt,
		comment.UserID,
		comment.Date,
		comment.Comment,
	); err != nil {
		return fmt.Errorf("failed inserting comment: %e", err)
	}

	return nil
}

func (m *CommentModel) GetByDateForUser(date time.Time, userID int) (Comment, error) {
	stmt := `
	select comment, created_at, updated_at
	from comments
	where date = $1
	and user_id = $2
	`

	row := m.DB.QueryRow(stmt, date, userID)

	c := Comment{}
	row.Scan(
		&c.Comment,
		&c.CreatedAt,
		&c.UpdatedAt,
	)

	return c, row.Err()
}
