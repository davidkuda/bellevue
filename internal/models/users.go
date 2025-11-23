package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int
	Email     string
	FirstName string
	LastName  string
	Method    string
	// email signups / logins:
	HashedPassword []byte
	// OpenID Connect signups / logins:
	SUB       string
	CreatedAt time.Time
}

type UserModel struct {
	DB *sql.DB
}

// Creates a new user in the database
func (m *UserModel) InsertPassword(u User, password string) error {
	var err error

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 15)
	if err != nil {
		return err
	}

	var stmt string
	var result sql.Result

	stmt = `
	INSERT INTO users (
		first_name, last_name, email, method, hashed_password
	) VALUES (
		$1,         $2,        $3,    $4,     $5
	);`
	result, err = m.DB.Exec(
		stmt,
		u.FirstName,
		u.LastName,
		u.Email,
		"password",
		string(hashedPassword),
	)
	if err != nil {
		return fmt.Errorf("failed inserting user: %s", err)
	}
	// TODO: what to do with rows affected?
	result.RowsAffected()

	return nil
}

func (m *UserModel) InsertOIDC(u User) error {
	var err error
	var stmt string
	var result sql.Result

	stmt = `
	INSERT INTO users (
		first_name, last_name, email, method, sub
	) VALUES (
		$1,         $2,        $3,    $4,     $5
	);`
	result, err = m.DB.Exec(
		stmt,
		u.FirstName,
		u.LastName,
		u.Email,
		"openidconnect",
		u.SUB,
	)
	if err != nil {
		return fmt.Errorf("failed inserting user: %s", err)
	}
	// TODO: what to do with rows affected?
	result.RowsAffected()

	return nil
}


func (m *UserModel) Authenticate(email, password string) error {
	var hashedPassword []byte

	stmt := "SELECT hashed_password FROM users WHERE email = $1;"

	err := m.DB.QueryRow(stmt, email).Scan(&hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidCredentials
		} else {
			return fmt.Errorf("DB.QueryRow(): %v", err)
		}
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidCredentials
		} else {
			return fmt.Errorf("failed to compare password: %v", err)
		}
	}

	return nil
}

func (m *UserModel) Exists(email string) (bool, error) {
	return false, nil
}

func (m *UserModel) GetUserIDByEmail(email string) (int, error) {
	stmt := "SELECT id FROM users WHERE email = $1;"

	var userID int

	err := m.DB.QueryRow(stmt, email).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("DB.QueryRow(): %v", err)
	}

	return userID, nil
}

func (m *UserModel) GetAll() ([]User, error) {
	stmt := `
	SELECT id, first_name, last_name, email
	FROM users;
	`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, fmt.Errorf("DB.Query(stmt): %v", err)
	}

	defer rows.Close()

	var users []User

	for rows.Next() {
		var user User
		err = rows.Scan(
			&user.ID,
			&user.FirstName,
			&user.LastName,
			&user.Email,
		)
		if err != nil {
			return nil, fmt.Errorf("for rows.Next(): %v", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err(): %v", err)
	}

	return users, nil
}

func (m *UserModel) GetUserByID(id int) (User, error) {
	stmt := `
	SELECT id, first_name, last_name, email
	FROM users
	WHERE id = $1;
	`

	var u User
	err := m.DB.QueryRow(stmt, id).Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
	)
	if err != nil {
		return u, fmt.Errorf("failed getting user by id with id=%d: %s", id, err)
	}

	return u, nil
}


func (m *UserModel) GetUserByEmail(email string) (User, error) {
	stmt := `
	SELECT id, first_name, last_name
	FROM users
	WHERE email = $1;
	`

	var u User

	err := m.DB.QueryRow(stmt, email).Scan(&u.ID, &u.FirstName, &u.LastName)
	if err != nil {
		return u, fmt.Errorf("DB.QueryRow(): failed getting user by email %s: %v", email, err)
	}
	u.Email = email

	return u, nil
}

func (m *UserModel) GetUserIDBySUB(sub string) (int, error) {
	stmt := `
	SELECT id
	FROM users
	WHERE sub = $1;
	`

	var userID int

	err := m.DB.QueryRow(stmt, sub).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("DB.QueryRow(): failed getting user by sub %s: %v", sub, err)
	}

	return userID, nil
}
