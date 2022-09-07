package database

import (
	"codepix/example-bank-api/adapters/databaseclient"
	"codepix/example-bank-api/customer/account"
	"codepix/example-bank-api/customer/account/repository"
	customerdatabase "codepix/example-bank-api/customer/repository/database"

	"github.com/google/uuid"
)

type Database struct {
	*databaseclient.Database
}

var _ repository.Repository = Database{}

func (db Database) Add(account account.Account, customerID uuid.UUID) (*uuid.UUID, error) {
	new := NewAccount(account, customerID)
	tx := db.Create(new)
	return databaseclient.GetID(tx), databaseclient.MapError(tx)
}

func (db Database) Find(ID uuid.UUID) (*account.Account, error) {
	var account Account
	tx := db.First(&account, ID)
	return AccountFromDB(account), databaseclient.MapError(tx)
}

func (db Database) Remove(ID uuid.UUID) error {
	tx := db.Delete(&Account{}, ID)
	return databaseclient.MapError(tx)
}

func (db Database) List(customerID uuid.UUID) ([]repository.AccountListItem, error) {
	var accounts []Account
	tx := db.DB.Find(&accounts, "customer_id = ?", customerID)
	return AccountListFromDB(accounts), databaseclient.MapError(tx)
}

func (db Database) ExistsWithCustomerID(ID uuid.UUID, customerID uuid.UUID) error {
	tx := db.Model(&Account{}).
		Select("1").
		First(new(int), "ID = ? and customer_id = ?", ID, customerID)
	return databaseclient.MapError(tx)
}

type Account struct {
	databaseclient.BaseModel
	Number     string                    `gorm:"unique"`
	Customer   customerdatabase.Customer `gorm:"<-:false;constraint:OnDelete:RESTRICT"`
	CustomerID uuid.UUID                 `gorm:"<-:create;index;not null"`
}

func NewAccount(account account.Account, customerID uuid.UUID) *Account {
	return &Account{
		BaseModel:  databaseclient.NewBaseModel(),
		Number:     account.Number,
		CustomerID: customerID,
	}
}

func AccountFromDB(dbAccount Account) *account.Account {
	if dbAccount == (Account{}) {
		return nil
	}
	return &account.Account{
		Number: dbAccount.Number,
	}
}

func AccountListFromDB(list []Account) []repository.AccountListItem {
	accounts := []repository.AccountListItem{}
	for _, dbAccount := range list {
		account := repository.AccountListItem{
			ID:     dbAccount.ID,
			Number: dbAccount.Number,
		}
		accounts = append(accounts, account)
	}
	return accounts
}
