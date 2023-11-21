package auth

import (
	"database/sql"
	"example/library-service/internal/entity"
	"log"
	"time"

	"github.com/google/uuid"
)

type AuthStore struct {
	db *sql.DB
}

func NewAuthStore(db *sql.DB) *AuthStore {
	return &AuthStore{db}
}

func (store *AuthStore) ExistsWithNameOrMail(name string, mail string) (bool, error) {
	statement, err := store.db.Prepare(`
		select count(*) from users where name=$1 or mail=$2 
	`)

	if err != nil {
		log.Println("AuthStore.ExistsWithNameOrMail() - received error from db", err)
		return true, err
	}

	row := statement.QueryRow(name, mail)
	var count int
	if scanErr := row.Scan(&count); scanErr != nil {
		log.Println("AuthStore.ExistsWithNameOrMail() - received error from db", scanErr)
		return true, scanErr
	}

	log.Println("AuthStore.ExistsWithNameOrMail() - received from db", count)
	if count > 0 {
		return true, nil
	}
	return false, nil
}

func (store *AuthStore) GetUserByIdAndRole(id uuid.UUID, role int) (u entity.User, err error) {
	statement, err := store.db.Prepare(`
		select u.id, u.name, u.mail, u.role, u.created_at, u.password from users u where u.id=$1 and u.role=$2 and u.token is not null 
	`)

	if err != nil {
		log.Println("AuthStore.GetUserByIdAndRole() - received error from db", err)
		return u, err
	}

	row := statement.QueryRow(id, role)

	if scanErr := row.Scan(&u.Id, &u.Name, &u.Mail, &u.Role, &u.CreatedAt, &u.Password); scanErr != nil {
		log.Println("AuthStore.GetUserByIdAndRole() - received error from db", scanErr)
		return u, scanErr
	}

	return u, nil
}

func (store *AuthStore) GetUserByName(name string) (u entity.User, e error) {
	statement, err := store.db.Prepare(`
		select u.id, u.name, u.mail, u.role, u.created_at, u.password from users u where u.name=$1 
	`)

	if err != nil {
		log.Println("AuthStore.GetUserByName() - received error from db", err)
		return u, err
	}

	rows := statement.QueryRow(name)

	if scanErr := rows.Scan(&u.Id, &u.Name, &u.Mail, &u.Role, &u.CreatedAt, &u.Password); scanErr != nil {
		log.Println("AuthStore.GetUserByName() - received error from db", scanErr)
		return u, scanErr
	}

	log.Println("AuthStore.GetUserByName() - received from db", u.Id, u.Name)
	return u, nil
}

func (store *AuthStore) CreateUser(user ReqisterRequest) error {
	statement, err := store.db.Prepare(`
		insert into users(name, mail, password, role, created_at)
			values($1, $2, $3, $4, $5) 
	`)

	if err != nil {
		log.Println("AuthStore.CreateUser() - received error from db", err)
		return err
	}

	_, err = statement.Exec(&user.Name, &user.Mail, &user.Password, &user.Role, time.Now().UTC())

	if err != nil {
		log.Println("AuthStore.CreateUser() - received error from db", err)
		return err
	}

	return nil
}

func (store *AuthStore) UpdateToken(user entity.User) error {
	statement, err := store.db.Prepare(`
		update users set token=$1 where id=$2
		returning *
	`)

	if err != nil {
		log.Println("AuthStore.UpdateToken() - received error from db", err)
		return err
	}

	_, execErr := statement.Exec(&user.Token, &user.Id)

	if execErr != nil {
		log.Println("AuthStore.UpdateToken() - received error from db", execErr)
		return execErr
	}

	return nil
}

func (store *AuthStore) DeleteToken(id uuid.UUID) (err error) {
	statement, err := store.db.Prepare(`
		update users set token=null where id=$1
	`)

	if err != nil {
		log.Println("AuthStore.DeleteToken() - received error from db", err)
		return err
	}
	_, err = statement.Exec(id)

	if err != nil {
		log.Println("AuthStore.DeleteToken() - received error from db", err)
		return err
	}

	return nil
}
