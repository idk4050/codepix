package database

import (
	"codepix/customer-api/adapters/databaseclient"
	"codepix/customer-api/customer/bank"
	"codepix/customer-api/customer/bank/repository"
	customerdatabase "codepix/customer-api/customer/repository/database"

	"github.com/google/uuid"
)

type Database struct {
	*databaseclient.Database
}

var _ repository.Repository = Database{}

func (db Database) Add(bank bank.Bank, customerID uuid.UUID) (*uuid.UUID, error) {
	new := NewBank(bank, customerID)
	tx := db.Create(new)
	return databaseclient.GetID(tx), databaseclient.MapError(tx)
}

func (db Database) Find(ID uuid.UUID) (*bank.Bank, error) {
	var bank Bank
	tx := db.First(&bank, ID)
	return BankFromDB(bank), databaseclient.MapError(tx)
}

func (db Database) Remove(ID uuid.UUID) error {
	tx := db.Delete(&Bank{}, ID)
	return databaseclient.MapError(tx)
}

func (db Database) List(customerID uuid.UUID) ([]repository.BankListItem, error) {
	var banks []Bank
	tx := db.DB.Find(&banks, "customer_id = ?", customerID)
	return BankListFromDB(banks), databaseclient.MapError(tx)
}

func (db Database) ExistsWithCustomerID(ID uuid.UUID, customerID uuid.UUID) error {
	tx := db.Model(&Bank{}).
		Select("1").
		First(new(int), "ID = ? and customer_id = ?", ID, customerID)
	return databaseclient.MapError(tx)
}

type Bank struct {
	databaseclient.BaseModel
	Code       uint32
	Name       string
	Customer   customerdatabase.Customer `gorm:"<-:false;constraint:OnDelete:RESTRICT"`
	CustomerID uuid.UUID                 `gorm:"<-:create;index;not null"`
}

func NewBank(bank bank.Bank, customerID uuid.UUID) *Bank {
	return &Bank{
		BaseModel:  databaseclient.NewBaseModel(),
		Code:       bank.Code,
		Name:       bank.Name,
		CustomerID: customerID,
	}
}

func BankFromDB(dbBank Bank) *bank.Bank {
	if dbBank == (Bank{}) {
		return nil
	}
	return &bank.Bank{
		Code: dbBank.Code,
		Name: dbBank.Name,
	}
}

func BankListFromDB(list []Bank) []repository.BankListItem {
	banks := []repository.BankListItem{}
	for _, dbBank := range list {
		bank := repository.BankListItem{
			ID:   dbBank.ID,
			Code: dbBank.Code,
			Name: dbBank.Name,
		}
		banks = append(banks, bank)
	}
	return banks
}
