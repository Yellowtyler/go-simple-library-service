package user

import (
	"database/sql"
	"example/library-service/internal/entity"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
)

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db}
}

func (store *UserStore) GetUser(id uuid.UUID) (u entity.User, e error) {
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

func (store *UserStore) GetUsers(m map[string]string) ([]entity.User, error) {
	query := "select id, name, mail, role, created_at from users"
	params := make([]any, len(m))

	if len(m) != 0 {
		query += " where "
		var i = 1
		for k, v := range m {
			if k == "role" {
				query += k + "=$" + fmt.Sprint(i) + " and "
			} else {
				query += k + " like '%' || $" + fmt.Sprint(i) + " || '%' and "
			}

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

	users := make([]entity.User, 0)
	for queryRows.Next() {
		var user entity.User
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

func (store *UserStore) UpdateUser(user entity.User) (updatedUser entity.User, err error) {
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
