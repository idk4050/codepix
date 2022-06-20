package database

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"codepix/bank-api/account"
	"codepix/bank-api/account/repository"
	"codepix/bank-api/adapters/databaseclient"
)

type Database struct {
	*gorm.DB
}

var _ repository.Repository = Database{}

func (db Database) Add(account account.Account, bankID uuid.UUID) (*uuid.UUID, error) {
	new := NewAccount(account, bankID)
	tx := db.Create(new)
	return databaseclient.GetID(tx), databaseclient.MapError(tx)
}

func (db Database) Find(ID uuid.UUID) (*account.Account, *repository.IDs, error) {
	var account Account
	tx := db.First(&account, "ID = ?", ID)
	return AccountFromDB(account), AccountIDs(account), databaseclient.MapError(tx)
}

func (db Database) FindByNumber(number account.Number) (*account.Account, *repository.IDs, error) {
	var account Account
	tx := db.First(&account, "number = ?", number)
	return AccountFromDB(account), AccountIDs(account), databaseclient.MapError(tx)
}

func (db Database) ExistsWithBankID(ID uuid.UUID, bankID uuid.UUID) error {
	tx := db.Model(&Account{}).
		Select("1").
		First(new(int), "ID = ? and bank_id = ?", ID, bankID)
	return databaseclient.MapError(tx)
}

type Account struct {
	databaseclient.BaseModel
	Number    account.Number `gorm:"<-:create;index:bank_account,unique"`
	OwnerName string         `gorm:"<-:create"`
	BankID    uuid.UUID      `gorm:"<-:create;index:bank_account,unique"`
}

func NewAccount(account account.Account, bankID uuid.UUID) *Account {
	return &Account{
		BaseModel: databaseclient.NewBaseModel(),
		Number:    account.Number,
		OwnerName: account.OwnerName,
		BankID:    bankID,
	}
}

func AccountFromDB(dbAccount Account) *account.Account {
	if dbAccount == (Account{}) {
		return nil
	}
	return &account.Account{
		Number:    dbAccount.Number,
		OwnerName: dbAccount.OwnerName,
	}
}

func AccountIDs(dbAccount Account) *repository.IDs {
	if dbAccount == (Account{}) {
		return nil
	}
	return &repository.IDs{
		AccountID: dbAccount.ID,
		BankID:    dbAccount.BankID,
	}
}
