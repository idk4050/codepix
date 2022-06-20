package database

import (
	accountdatabase "codepix/bank-api/account/repository/database"
	"codepix/bank-api/adapters/databaseclient"
	"codepix/bank-api/pixkey"
	"codepix/bank-api/pixkey/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Database struct {
	*gorm.DB
}

var _ repository.Repository = Database{}

func (db Database) Add(pixKey pixkey.PixKey, accountID uuid.UUID) (*uuid.UUID, error) {
	new := NewPixKey(pixKey, accountID)
	tx := db.Create(new)
	return databaseclient.GetID(tx), databaseclient.MapError(tx)
}

func (db Database) Find(ID uuid.UUID) (*pixkey.PixKey, *repository.IDs, error) {
	var pixKey PixKey
	tx := db.Preload("Account").First(&pixKey, "ID = ?", ID)
	return PixKeyFromDB(pixKey), PixKeyIDs(pixKey), databaseclient.MapError(tx)
}

func (db Database) FindByKey(key pixkey.Key) (*pixkey.PixKey, *repository.IDs, error) {
	var pixKey PixKey
	tx := db.Preload("Account").First(&pixKey, "key = ?", key)
	return PixKeyFromDB(pixKey), PixKeyIDs(pixKey), databaseclient.MapError(tx)
}

type PixKey struct {
	databaseclient.BaseModel
	Type      pixkey.Type             `gorm:"<-:create;"`
	Key       pixkey.Key              `gorm:"<-:create;uniqueIndex"`
	Account   accountdatabase.Account `gorm:"<-:false;constraint:OnDelete:CASCADE"`
	AccountID uuid.UUID               `gorm:"<-:create;index"`
}

func NewPixKey(pixKey pixkey.PixKey, accountID uuid.UUID) *PixKey {
	return &PixKey{
		BaseModel: databaseclient.NewBaseModel(),
		Type:      pixKey.Type,
		Key:       pixKey.Key,
		AccountID: accountID,
	}
}

func PixKeyFromDB(dbPixKey PixKey) *pixkey.PixKey {
	if dbPixKey == (PixKey{}) {
		return nil
	}
	return &pixkey.PixKey{
		Type: dbPixKey.Type,
		Key:  dbPixKey.Key,
	}
}

func PixKeyIDs(dbPixKey PixKey) *repository.IDs {
	if dbPixKey == (PixKey{}) {
		return nil
	}
	return &repository.IDs{
		PixKeyID:  dbPixKey.ID,
		AccountID: dbPixKey.AccountID,
		BankID:    dbPixKey.Account.BankID,
	}
}
