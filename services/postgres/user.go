package main

import (
	"errors"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func CreateUser(name, email string) (int, error) {
	var id int
	err := db.QueryRow(
		"INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
		name, email,
	).Scan(&id)
	return id, err
}

func GetAllUsers() ([]User, error) {
	rows, err := db.Query("SELECT id, name, email FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func UpdateUser(id int, name, email string) error {
	res, err := db.Exec(
		"UPDATE users SET name=$1, email=$2 WHERE id=$3",
		name, email, id,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("user not found")
	}
	return nil
}

func DeleteUser(id int) error {
	res, err := db.Exec("DELETE FROM users WHERE id=$1", id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("user not found")
	}
	return nil
}
