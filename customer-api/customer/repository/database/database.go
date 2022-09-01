package database

import (
	"codepix/customer-api/adapters/databaseclient"
	"codepix/customer-api/customer"
	"codepix/customer-api/customer/repository"
	userdatabase "codepix/customer-api/user/repository/database"

	"github.com/google/uuid"
)

type Database struct {
	*databaseclient.Database
}

var _ repository.Repository = Database{}

func (db Database) Find(ID uuid.UUID) (*customer.Customer, error) {
	var customer Customer
	tx := db.First(&customer, ID)
	return CustomerFromDB(customer), databaseclient.MapError(tx)
}

func (db Database) FindByUserID(ID uuid.UUID) (*customer.Customer, *uuid.UUID, error) {
	var customer Customer
	tx := db.First(&customer, "user_id = ?", ID)
	return CustomerFromDB(customer), databaseclient.GetID(tx), databaseclient.MapError(tx)
}

type Customer struct {
	databaseclient.BaseModel
	Name   string
	User   userdatabase.User `gorm:"<-:false;constraint:OnDelete:CASCADE"`
	UserID uuid.UUID         `gorm:"<-:create;uniqueIndex;not null"`
}

func CustomerFromDB(dbCustomer Customer) *customer.Customer {
	if dbCustomer == (Customer{}) {
		return nil
	}
	return &customer.Customer{
		Name: dbCustomer.Name,
	}
}
