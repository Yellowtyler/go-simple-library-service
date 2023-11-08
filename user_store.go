package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db}
}

func (store *UserStore) GetUser(id uuid.UUID) (u User, e error) {
	statement, err := store.db.Prepare(`
		select u.* from users u where id=$1 
	`)

	if err != nil {
		log.Println("UserStore.GetUser() - received error from db", err)
		return u, err
	}

	rows := statement.QueryRow(id.String())

	if scanErr := rows.Scan(&u.Id, &u.Name, &u.Mail, &u.Password, &u.Role, &u.CreatedAt); scanErr != nil {
		log.Println("UserStore.GetUser() - received error from db", scanErr)
		return u, scanErr
	}

	log.Println("UserStore.GetUser() - received from db", u)
	return u, nil
}

func (store *UserStore) GetUserByName(name string) (u User, e error) {
	statement, err := store.db.Prepare(`
		select u.* from users u where name=$1 
	`)

	if err != nil {
		log.Println("UserStore.GetUserByName() - received error from db", err)
		return u, err
	}

	rows := statement.QueryRow(name)

	if scanErr := rows.Scan(&u.Id, &u.Name, &u.Mail, &u.Password, &u.Role, &u.CreatedAt); scanErr != nil {
		log.Println("UserStore.GetUserByName() - received error from db", scanErr)
		return u, scanErr
	}

	log.Println("UserStore.GetUserByName() - received from db", u)
	return u, nil
}

func (store *UserStore) GetUsers(m map[string]string) ([]User, error) {
	query := "select * from users u"
	params := make([]any, len(m))

	if len(m) != 0 {
		query += " where "
		var i = 1
		for k, v := range m {
			query += k + " like '%' || $" + fmt.Sprint(i) + " || '%' and "
			params[i-1] = v
			i++
		}
		query = strings.TrimSpace(query)
		queryArr := strings.Split(query, " ")
		query = strings.Join(queryArr[:len(queryArr)-1], " ")
	}

	log.Println("UserStore.GetUsers() - executing query", query, params)

	statement, err := store.db.Prepare(query)

	if err != nil {
		log.Println("UserStore.GetUsers() - received error from db", err)
		return nil, err
	}

	var queryRows *sql.Rows
	var queryError error
	if len(params) == 0 {
		queryRows, queryError = statement.Query()
	} else {
		queryRows, queryError = statement.Query(params...)
	}

	if queryError != nil {
		log.Println("UserStore.GetUsers() - received error from db", queryError)
		return nil, queryError
	}

	defer queryRows.Close()

	users := make([]User, 0)
	for queryRows.Next() {
		var user User
		if scanErr := queryRows.Scan(&user.Id, &user.Name, &user.Mail, &user.Password, &user.Role, &user.CreatedAt); scanErr != nil {
			log.Println("UserStore.GetUsers() - received error while scanning", err)
		}
	}

	if err := queryRows.Err(); err != nil {
		log.Println("UserStore.GetUsers() - received error from db", err)
		return nil, err
	}

	return users, nil
}

func (store *UserStore) CreateUser(user User) (savedUser User, err error) {
	statement, err := store.db.Prepare(`
		insert into Users(name, mail, password, role, created_at)
			values($1, $2, $3, $4, $5) 
			returning *
	`)

	if err != nil {
		log.Println("UserStore.CreateUser() - received error from db", err)
		return savedUser, err
	}

	password, err := hashAndSalt([]byte(user.Password))
	if err != nil {
		return savedUser, err
	}

	row := statement.QueryRow(&user.Name, &user.Mail, password, &user.Role, time.Now().UTC())

	scanError := row.Scan(&savedUser.Id, &savedUser.Name, &savedUser.Mail, &savedUser.Password, &savedUser.Role, &savedUser.CreatedAt)

	if scanError != nil {
		log.Println("UserStore.CreateUser() - received error from db", scanError)
		return savedUser, scanError
	}

	return savedUser, nil
}

func (store *UserStore) UpdateUser(user User) (updatedUser User, err error) {
	statement, err := store.db.Prepare(`
		update users set name=$1, mail=$2, password=$3, role=$4 where id=$5
		returning *
	`)

	if err != nil {
		log.Println("UserStore.UpdateUser() - received error from db", err)
		return updatedUser, err
	}

	password, err := hashAndSalt([]byte(updatedUser.Password))
	if err != nil {
		return updatedUser, err
	}

	row := statement.QueryRow(&user.Name, &user.Mail, password, &user.Role, &user.Id)

	if scanError := row.Scan(&updatedUser.Id, &updatedUser.Name, &updatedUser.CreatedAt); scanError != nil {
		log.Println("UserStore.UpdateUser() - received error from db", scanError)
		return updatedUser, scanError
	}

	return updatedUser, nil
}

func (store *UserStore) DeleteUser(id uuid.UUID) error {

	deleteStatement, deleteErr := store.db.Prepare(`delete from users where id=$1`)

	if deleteErr != nil {
		log.Println("UserStore.DeleteUser() - received error from db", deleteErr)
		return deleteErr
	}

	if _, execErr := deleteStatement.Exec(id); execErr != nil {
		log.Println("UserStore.DeleteUser() - received error from db", execErr)
		return execErr
	}

	return nil
}