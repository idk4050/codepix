package database

import (
	"codepix/customer-api/adapters/databaseclient"
	"codepix/customer-api/lib/outboxes"
	"codepix/customer-api/user/signin"
	"codepix/customer-api/user/signin/eventhandler"
	"codepix/customer-api/user/signin/repository"

	"gorm.io/gorm"
)

type Database struct {
	*databaseclient.Database
	Outbox outboxes.Outbox
}

var _ repository.Repository = Database{}

func (db Database) Start(signIn signin.SignIn) error {
	new := NewSignIn(signIn)
	return db.Transaction(func(tx *gorm.DB) error {
		step := tx.Delete(&SignIn{}, "Email = ?", new.Email)
		if step.Error != nil {
			return databaseclient.MapError(step)
		}
		step = tx.Create(new)
		if step.Error != nil {
			return databaseclient.MapError(step)
		}
		event := eventhandler.Started{
			Email: new.Email,
			Token: new.Token,
		}
		return db.Outbox.Write(tx, event)
	})
}

func (db Database) Find(token string) (*signin.SignIn, error) {
	var signIn SignIn
	tx := db.First(&signIn, "Token = ?", token)
	return SignInFromDB(signIn), databaseclient.MapError(tx)
}

func (db Database) Finish(token string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var signIn SignIn
		step := tx.First(&signIn, "Token = ?", token)
		if step.Error != nil {
			return databaseclient.MapError(step)
		}
		step = tx.Delete(signIn)
		if step.Error != nil {
			return databaseclient.MapError(step)
		}
		return nil
	})
}

type SignIn struct {
	databaseclient.BaseModel
	Email string `gorm:"uniqueIndex"`
	Token string `gorm:"uniqueIndex"`
}

func NewSignIn(signIn signin.SignIn) *SignIn {
	return &SignIn{
		BaseModel: databaseclient.NewBaseModel(),
		Email:     signIn.Email,
		Token:     signIn.Token,
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
