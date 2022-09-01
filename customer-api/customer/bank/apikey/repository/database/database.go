package database

import (
	"codepix/customer-api/adapters/databaseclient"
	"codepix/customer-api/customer/bank/apikey"
	"codepix/customer-api/customer/bank/apikey/repository"
	bankdatabase "codepix/customer-api/customer/bank/repository/database"

	"github.com/google/uuid"
)

type Database struct {
	*databaseclient.Database
}

var _ repository.Repository = Database{}

func (db Database) Add(apiKey apikey.APIKey, bankID uuid.UUID) (*uuid.UUID, error) {
	new := NewAPIKey(apiKey, bankID)
	tx := db.Create(new)
	return databaseclient.GetID(tx), databaseclient.MapError(tx)
}

func (db Database) Remove(ID uuid.UUID) error {
	tx := db.Delete(&APIKey{}, ID)
	return databaseclient.MapError(tx)
}

func (db Database) List(bankID uuid.UUID) ([]repository.APIKeyListItem, error) {
	var apiKeys []APIKey
	tx := db.Find(&apiKeys, "bank_id = ?", bankID)
	return APIKeysFromDB(apiKeys), databaseclient.MapError(tx)
}

func (db Database) FindBankID(hash apikey.Hash) (*uuid.UUID, error) {
	var apiKey APIKey
	tx := db.First(&apiKey, "Hash = ?", hash)
	if tx.Error != nil {
		return nil, databaseclient.MapError(tx)
	}
	return &apiKey.BankID, nil
}

type APIKey struct {
	databaseclient.BaseModel
	Name   string
	Hash   apikey.Hash       `gorm:"uniqueIndex"`
	Bank   bankdatabase.Bank `gorm:"<-:false;constraint:OnDelete:CASCADE"`
	BankID uuid.UUID         `gorm:"<-:create;index;not null"`
}

func NewAPIKey(apiKey apikey.APIKey, bankID uuid.UUID) *APIKey {
	baseModel := databaseclient.NewBaseModel()
	return &APIKey{
		BaseModel: baseModel,
		Name:      apiKey.Name,
		Hash:      apiKey.Hash,
		BankID:    bankID,
	}
}

func APIKeysFromDB(dbAPIKeys []APIKey) []repository.APIKeyListItem {
	apiKeys := []repository.APIKeyListItem{}
	for _, key := range dbAPIKeys {
		apiKey := repository.APIKeyListItem{
			ID:        key.ID,
			Name:      key.Name,
			CreatedAt: key.CreatedAt,
		}
		apiKeys = append(apiKeys, apiKey)
	}
	return apiKeys
}
