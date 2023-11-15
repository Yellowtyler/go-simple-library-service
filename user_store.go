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
		select id, name, mail, role, created_at from users where id=$1 
	`)

	if err != nil {
		log.Println("UserStore.GetUser() - received error from db", err)
		return u, err
	}

	rows := statement.QueryRow(id.String())

	if scanErr := rows.Scan(&u.Id, &u.Name, &u.Mail, &u.Role, &u.CreatedAt); scanErr != nil {
		log.Println("UserStore.GetUser() - received error from db", scanErr)
		return u, scanErr
	}

	log.Println("UserStore.GetUser() - received from db", u)
	return u, nil
}

func (store *UserStore) ExistsWithNameOrMail(name string, mail string) (bool, error) {
	statement, err := store.db.Prepare(`
		select count(*) from users where name=$1 or mail=$2 
	`)

	if err != nil {
		log.Println("UserStore.ExistsWithNameOrMail() - received error from db", err)
		return true, err
	}

	row := statement.QueryRow(name, mail)
	var count int
	if scanErr := row.Scan(&count); scanErr != nil {
		log.Println("UserStore.ExistsWithNameOrMail() - received error from db", scanErr)
		return true, scanErr
	}

	log.Println("UserStore.ExistsWithNameOrMail() - received from db", count)
	if count > 0 {
		return true, nil
	}
	return false, nil
}

func (store *UserStore) GetUserByIdAndRole(id uuid.UUID, role int) (u User, err error) {
	statement, err := store.db.Prepare(`
		select u.id, u.name, u.mail, u.role, u.created_at, u.password from users u where u.id=$1 and u.role=$2 and u.token is not null 
	`)

	if err != nil {
		log.Println("UserStore.GetUserByIdAndRole() - received error from db", err)
		return u, err
	}

	row := statement.QueryRow(id, role)

	if scanErr := row.Scan(&u.Id, &u.Name, &u.Mail, &u.Role, &u.CreatedAt, &u.Password); scanErr != nil {
		log.Println("UserStore.GetUserByIdAndRole() - received error from db", scanErr)
		return u, scanErr
	}

	return u, nil
}

func (store *UserStore) GetUserByName(name string) (u User, e error) {
	statement, err := store.db.Prepare(`
		select u.id, u.name, u.mail, u.role, u.created_at, u.password from users u where u.name=$1 
	`)

	if err != nil {
		log.Println("UserStore.GetUserByName() - received error from db", err)
		return u, err
	}

	rows := statement.QueryRow(name)

	if scanErr := rows.Scan(&u.Id, &u.Name, &u.Mail, &u.Role, &u.CreatedAt, &u.Password); scanErr != nil {
		log.Println("UserStore.GetUserByName() - received error from db", scanErr)
		return u, scanErr
	}

	log.Println("UserStore.GetUserByName() - received from db", u.Id, u.Name)
	return u, nil
}

func (store *UserStore) GetUsers(m map[string]string) ([]User, error) {
	query := "select id, name, mail, role, created_at from users"
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
		if scanErr := queryRows.Scan(&user.Id, &user.Name, &user.Mail, &user.Role, &user.CreatedAt); scanErr != nil {
			log.Println("UserStore.GetUsers() - received error while scanning", err)
		}
		users = append(users, user)
	}

	if err := queryRows.Err(); err != nil {
		log.Println("UserStore.GetUsers() - received error from db", err)
		return nil, err
	}

	return users, nil
}

func (store *UserStore) CreateUser(user ReqisterRequest) error {
	statement, err := store.db.Prepare(`
		insert into users(name, mail, password, role, created_at)
			values($1, $2, $3, $4, $5) 
	`)

	if err != nil {
		log.Println("UserStore.CreateUser() - received error from db", err)
		return err
	}

	_, err = statement.Exec(&user.Name, &user.Mail, &user.Password, &user.Role, time.Now().UTC())

	if err != nil {
		log.Println("UserStore.CreateUser() - received error from db", err)
		return err
	}

	return nil
}

func (store *UserStore) UpdateUser(user User) (updatedUser User, err error) {
	statement, err := store.db.Prepare(`
		update users set name=$1, mail=$2, role=$3 where id=$4
		returning id, name, mail, role, created_at
	`)

	if err != nil {
		log.Println("UserStore.UpdateUser() - received error from db", err)
		return updatedUser, err
	}

	// password, err := HashAndSalt([]byte(updatedUser.Password))
	// if err != nil {
	// 	return updatedUser, err
	// }

	row := statement.QueryRow(&user.Name, &user.Mail, &user.Role, &user.Id)

	if scanError := row.Scan(&updatedUser.Id, &updatedUser.Name, &updatedUser.Mail, &updatedUser.Role, &updatedUser.CreatedAt); scanError != nil {
		log.Println("UserStore.UpdateUser() - received error from db", scanError)
		return updatedUser, scanError
	}

	return updatedUser, nil
}

func (store *UserStore) UpdateToken(user User) error {
	statement, err := store.db.Prepare(`
		update users set token=$1 where id=$2
		returning *
	`)

	if err != nil {
		log.Println("UserStore.UpdateToken() - received error from db", err)
		return err
	}

	_, execErr := statement.Exec(&user.Token, &user.Id)

	if execErr != nil {
		log.Println("UserStore.UpdateToken() - received error from db", execErr)
		return execErr
	}

	return nil
}

func (store *UserStore) DeleteToken(id uuid.UUID) (err error) {
	statement, err := store.db.Prepare(`
		update users set token=null where id=$1
	`)

	if err != nil {
		log.Println("UserStore.DeleteToken() - received error from db", err)
		return err
	}
	_, err = statement.Exec(id)

	if err != nil {
		log.Println("UserStore.DeleteToken() - received error from db", err)
		return err
	}

	return nil
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
