package models

import "github.com/google/uuid"

// UserInfo is the data we get back from the auth service
type UserInfo struct {
	Id            string `json:"id"`
	Email         string `json:"email"`
	Picture       string `json:"picture"`
	VerifiedEmail bool   `json:"verified_email"`
}

// User is an example model in the application (saved in the db)
type User struct {
	UUID    string  `db:"uuid"`
	Email   string  `db:"email"`
	Name    *string `db:"username"`
	Picture *string `db:"picture"`
	AuthId  string  `db:"authid"`
	Salt    *string `db:"salt"`
}

func NewUser(authid string, email string, picture string) *User {
	id := uuid.New()
	a := User{
		UUID:    id.String(),
		AuthId:  authid,
		Email:   email,
		Picture: &picture,
	}
	return &a
}
