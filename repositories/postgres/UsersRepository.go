package postgres

import (
	// client "avito-user-balance/db/postgres"
	"avito-user-balance/models"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type User models.User

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(d *sql.DB) *UserRepository {
	return &(UserRepository{d})
}

func (ur *UserRepository) FindUserByID(id int) (*models.User, error) {
	u := models.User{}
	row := ur.db.QueryRow("SELECT * FROM users WHERE id=$1;", id)
	err := row.Scan(&u.ID, &u.Balance, &u.Reserve)
	switch err {
	case sql.ErrNoRows:
		return nil, fmt.Errorf("No User with ID %d found", id) // 404 Not found with err struct
	case nil:
		return &u, nil
	default:
		return nil, err
	}
}

func (ur *UserRepository) FindAllUsers() (models.Users, error) {
	var users models.Users

	rows, err := ur.db.Query("SELECT * FROM users;")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("No Users in the database")
		}
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var user = models.User{}
		err = rows.Scan(&user.ID, &user.Balance, &user.Reserve)
		users = append(users, &user)
		if err != nil {
			return nil, err
		}
	}
	return users, nil
}

func (ur *UserRepository) AddUser(newValue *models.User) error {
	_, err := ur.db.Exec(
		"INSERT INTO users VALUES($1, $2, $3);",
		newValue.ID, newValue.Balance, newValue.Reserve)
	return err
}

func (ur *UserRepository) UpdateUser(newValue *models.User) error {
	// var queryString string
	// fmt.Sprintf()
	_, err := ur.db.Exec(
		"UPDATE users SET (balance, reserve) = ($1, $2) WHERE id=$3;",
		newValue.Balance, newValue.Reserve, newValue.ID)
	return err
}
