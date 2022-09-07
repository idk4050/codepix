package database

import (
	"codepix/example-bank-api/adapters/databaseclient"
	customerdatabase "codepix/example-bank-api/customer/repository/database"
	"codepix/example-bank-api/customer/signup"
	"codepix/example-bank-api/customer/signup/repository"
	"codepix/example-bank-api/lib/repositories"
	userdatabase "codepix/example-bank-api/user/repository/database"

	"gorm.io/gorm"
)

type Database struct {
	*databaseclient.Database
}

var _ repository.Repository = Database{}

func (db Database) Add(signUp signup.SignUp) error {
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
		return databaseclient.MapError(step)
	})
}

func (db Database) Remove(token string) error {
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
		return databaseclient.MapError(step)
	})
}

type SignUp struct {
	databaseclient.BaseModel
	Name  string
	Email string `gorm:"unique"`
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
