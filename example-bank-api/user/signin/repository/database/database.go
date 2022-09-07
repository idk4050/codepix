package database

import (
	"codepix/example-bank-api/adapters/databaseclient"
	userdatabase "codepix/example-bank-api/user/repository/database"
	"codepix/example-bank-api/user/signin"
	"codepix/example-bank-api/user/signin/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Database struct {
	*databaseclient.Database
}

var _ repository.Repository = Database{}

func (db Database) Add(signIn signin.SignIn, userID uuid.UUID) error {
	new := NewSignIn(signIn, userID)
	return db.Transaction(func(tx *gorm.DB) error {
		step := tx.Delete(&SignIn{}, "Email = ?", new.Email)
		if step.Error != nil {
			return databaseclient.MapError(step)
		}
		step = tx.Create(new)
		return databaseclient.MapError(step)
	})
}

func (db Database) Find(token string) (*signin.SignIn, *repository.IDs, error) {
	var signIn SignIn
	tx := db.First(&signIn, "Token = ?", token)
	return SignInFromDB(signIn), SignInIDs(signIn), databaseclient.MapError(tx)
}

func (db Database) Remove(token string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var signIn SignIn
		step := tx.First(&signIn, "Token = ?", token)
		if step.Error != nil {
			return databaseclient.MapError(step)
		}
		step = tx.Delete(signIn)
		return databaseclient.MapError(step)
	})
}

type SignIn struct {
	databaseclient.BaseModel
	Email  string            `gorm:"uniqueIndex"`
	Token  string            `gorm:"uniqueIndex"`
	User   userdatabase.User `gorm:"<-:false;constraint:OnDelete:CASCADE"`
	UserID uuid.UUID         `gorm:"<-:create;not null"`
}

func NewSignIn(signIn signin.SignIn, userID uuid.UUID) *SignIn {
	return &SignIn{
		BaseModel: databaseclient.NewBaseModel(),
		Email:     signIn.Email,
		Token:     signIn.Token,
		UserID:    userID,
	}
}

func SignInFromDB(dbSignIn SignIn) *signin.SignIn {
	if dbSignIn == (SignIn{}) {
		return nil
	}
	return &signin.SignIn{
		Email: dbSignIn.Email,
		Token: dbSignIn.Token,
	}
}

func SignInIDs(dbSignIn SignIn) *repository.IDs {
	if dbSignIn == (SignIn{}) {
		return nil
	}
	return &repository.IDs{
		SignInID: dbSignIn.ID,
		UserID:   dbSignIn.UserID,
	}
}
