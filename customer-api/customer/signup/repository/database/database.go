package database

import (
	"codepix/customer-api/adapters/databaseclient"
	customerdatabase "codepix/customer-api/customer/repository/database"
	"codepix/customer-api/customer/signup"
	"codepix/customer-api/customer/signup/eventhandler"
	"codepix/customer-api/customer/signup/repository"
	"codepix/customer-api/lib/outboxes"
	"codepix/customer-api/lib/repositories"
	userdatabase "codepix/customer-api/user/repository/database"

	"gorm.io/gorm"
)

type Database struct {
	*databaseclient.Database
	Outbox outboxes.Outbox
}

var _ repository.Repository = Database{}

func (db Database) Start(signUp signup.SignUp) error {
	new := NewSignUp(signUp)
	return db.Transaction(func(tx *gorm.DB) error {
		var alreadyExists bool
		step := tx.Model(&userdatabase.User{}).
			Select("count(*) > 0").Find(&alreadyExists, "Email = ?", new.Email)
		if alreadyExists {
			return &repositories.AlreadyExistsError{databaseclient.GetSchemaName(step)}
		}
		step = tx.Delete(&SignUp{}, "Email = ?", new.Email)
		if step.Error != nil {
			return databaseclient.MapError(step)
		}
		step = tx.Create(new)
		if step.Error != nil {
			return databaseclient.MapError(step)
		}
		event := eventhandler.Started{
			Name:  new.Name,
			Email: new.Email,
			Token: new.Token,
		}
		return db.Outbox.Write(tx, event)
	})
}

func (db Database) Finish(token string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var signUp SignUp
		step := tx.First(&signUp, "Token = ?", token)
		if step.Error != nil {
			return databaseclient.MapError(step)
		}
		var alreadyExists bool
		step = tx.Model(&userdatabase.User{}).
			Select("count(*) > 0").Find(&alreadyExists, "Email = ?", signUp.Email)
		if alreadyExists {
			return &repositories.AlreadyExistsError{databaseclient.GetSchemaName(step)}
		}
		step = tx.Delete(signUp)
		if step.Error != nil {
			return databaseclient.MapError(step)
		}

		newUser := &userdatabase.User{
			BaseModel: databaseclient.NewBaseModel(),
			Email:     signUp.Email,
		}
		step = tx.Create(newUser)
		if step.Error != nil {
			return databaseclient.MapError(step)
		}

		newCustomer := &customerdatabase.Customer{
			BaseModel: databaseclient.NewBaseModel(),
			Name:      signUp.Name,
			UserID:    newUser.ID,
		}
		step = tx.Create(newCustomer)
		if step.Error != nil {
			return databaseclient.MapError(step)
		}

		event := eventhandler.Finished{
			Email: signUp.Email,
		}
		return db.Outbox.Write(tx, event)
	})
}

type SignUp struct {
	databaseclient.BaseModel
	Name  string
	Email string `gorm:"uniqueIndex"`
	Token string `gorm:"uniqueIndex"`
}

func NewSignUp(signUp signup.SignUp) *SignUp {
	return &SignUp{
		BaseModel: databaseclient.NewBaseModel(),
		Name:      signUp.Name,
		Email:     signUp.Email,
		Token:     signUp.Token,
	}
}

func SignUpFromDB(dbSignUp SignUp) *signup.SignUp {
	if dbSignUp == (SignUp{}) {
		return nil
	}
	return &signup.SignUp{
		Name:  dbSignUp.Name,
		Email: dbSignUp.Email,
		Token: dbSignUp.Token,
	}
}
