package database

import (
	"codepix/example-bank-api/adapters/databaseclient"
	"codepix/example-bank-api/user"
	"codepix/example-bank-api/user/repository"

	"github.com/google/uuid"
)

type Database struct {
	*databaseclient.Database
}

var _ repository.Repository = Database{}

func (db Database) Add(user user.User) (*uuid.UUID, error) {
	new := NewUser(user)
	tx := db.Create(new)
	return databaseclient.GetID(tx), databaseclient.MapError(tx)
}

func (db Database) FindByEmail(email string) (*user.User, *uuid.UUID, error) {
	var user User
	tx := db.First(&user, "Email = ?", email)
	return UserFromDB(user), databaseclient.GetID(tx), databaseclient.MapError(tx)
}

type User struct {
	databaseclient.BaseModel
	Email string `gorm:"uniqueIndex"`
}

func NewUser(user user.User) *User {
	return &User{
		BaseModel: databaseclient.NewBaseModel(),
		Email:     user.Email,
	}
}

func UserFromDB(dbUser User) *user.User {
	if dbUser == (User{}) {
		return nil
	}
	return &user.User{
		Email: dbUser.Email,
	}
}
