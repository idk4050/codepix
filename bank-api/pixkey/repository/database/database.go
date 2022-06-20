package database

import (
	"codepix/bank-api/adapters/databaseclient"
	"codepix/bank-api/pixkey"
	"codepix/bank-api/pixkey/repository"

	"github.com/google/uuid"
)

type Database struct {
	*databaseclient.Database
}

var _ repository.Repository = Database{}

func (db Database) Add(pixKey pixkey.PixKey, accountID, bankID uuid.UUID) (*uuid.UUID, error) {
	new := NewPixKey(pixKey, accountID, bankID)
	tx := db.Create(new)
	return databaseclient.GetID(tx), databaseclient.MapError(tx)
}

func (db Database) Find(ID uuid.UUID) (*pixkey.PixKey, *repository.IDs, error) {
	var pixKey PixKey
	tx := db.First(&pixKey, "ID = ?", ID)
	return PixKeyFromDB(pixKey), PixKeyIDs(pixKey), databaseclient.MapError(tx)
}

func (db Database) FindByKey(key pixkey.Key) (*pixkey.PixKey, *repository.IDs, error) {
	var pixKey PixKey
	tx := db.First(&pixKey, "key = ?", key)
	return PixKeyFromDB(pixKey), PixKeyIDs(pixKey), databaseclient.MapError(tx)
}

func (db Database) List(options repository.ListOptions) ([]repository.ListItem, error) {
	var pixKeys []PixKey
	tx := db.DB.Find(&pixKeys, "account_id = ? and bank_id = ?", options.AccountID, options.BankID)
	return PixKeysFromDB(pixKeys), databaseclient.MapError(tx)
}

type PixKey struct {
	databaseclient.BaseModel
	Type      pixkey.Type `gorm:"<-:create;"`
	Key       pixkey.Key  `gorm:"<-:create;uniqueIndex"`
	AccountID uuid.UUID   `gorm:"<-:create;index"`
	BankID    uuid.UUID   `gorm:"<-:create;index"`
}

func NewPixKey(pixKey pixkey.PixKey, accountID, bankID uuid.UUID) *PixKey {
	return &PixKey{
		BaseModel: databaseclient.NewBaseModel(),
		Type:      pixKey.Type,
		Key:       pixKey.Key,
		AccountID: accountID,
		BankID:    bankID,
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
		BankID:    dbPixKey.BankID,
	}
}

func PixKeysFromDB(dbPixKeys []PixKey) []repository.ListItem {
	if dbPixKeys == nil {
		return nil
	}
	pixKeys := []repository.ListItem{}
	for _, pixKey := range dbPixKeys {
		pixKeys = append(pixKeys, repository.ListItem{
			ID:   pixKey.ID,
			Type: pixKey.Type,
			Key:  pixKey.Key,
		})
	}
	return pixKeys
}
