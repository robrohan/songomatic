package repository

import (
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/robrohan/go-web-template/internals/models"
)

type DataRepository struct {
	Db                  *sqlx.DB
	upsertUserQuery     *sqlx.Stmt
	getUserByEmailQuery *sqlx.Stmt
	getUserByIdQuery    *sqlx.Stmt
}

func prepareQuery(query string, db *sqlx.DB) *sqlx.Stmt {
	stmt, err := db.Preparex(query)
	if err != nil {
		log.Fatal(err)
	}
	return stmt
}

// Attach creates a new repository and sets up prepared statements
func Attach(schema string, db *sqlx.DB, driver string) *DataRepository {
	a := DataRepository{
		Db: db,
	}

	a.upsertUserQuery = prepareQuery(`
		INSERT INTO users (uuid, authid, email, picture, salt)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (email) DO UPDATE
			SET picture = $4,
			salt = $5;
	`, db)

	a.getUserByEmailQuery = prepareQuery(`
		SELECT uuid, email, username, picture, authid, salt
		FROM users
		WHERE email = $1
	`, db)

	a.getUserByIdQuery = prepareQuery(`
		SELECT uuid, email, username, picture, authid, salt
		FROM users
		WHERE uuid = $1
	`, db)

	return &a
}

func (r *DataRepository) Begin() (*sqlx.Tx, error) {
	return r.Db.Beginx()
}

func (r *DataRepository) UpsertUser(user *models.User, salt string) error {
	_, err := r.upsertUserQuery.Exec(
		user.UUID, user.AuthId, user.Email, user.Picture, salt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *DataRepository) GetUserById(uuid uuid.UUID) (*models.User, error) {
	rows, err := r.getUserByIdQuery.Queryx(uuid)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return nil, errors.New("no rows")
	}

	user := models.User{}
	for rows.Next() {
		err = rows.StructScan(&user)
		if err != nil {
			return nil, err
		}
	}

	return &user, nil
}

func (r *DataRepository) GetUser(email string) (*models.User, error) {
	rows, err := r.getUserByEmailQuery.Queryx(email)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return nil, errors.New("no rows")
	}

	user := models.User{}
	for rows.Next() {
		err = rows.StructScan(&user)
		if err != nil {
			return nil, err
		}
	}

	return &user, nil
}
