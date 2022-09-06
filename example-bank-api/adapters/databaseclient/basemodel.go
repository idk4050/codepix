package databaseclient

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uuid.UUID `gorm:"<-:create;primarykey;type:uuid;not null"`
	CreatedAt time.Time `gorm:"<-:create;not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func NewBaseModel() BaseModel {
	return BaseModel{ID: uuid.New()}
}

func (b BaseModel) GetID() uuid.UUID {
	return b.ID
}

type Identifiable interface {
	GetID() uuid.UUID
}

func GetID(tx *gorm.DB) *uuid.UUID {
	if tx.Error != nil {
		return nil
	}
	model := tx.Statement.Dest
	if model == nil {
		return nil
	}
	ID := model.(Identifiable).GetID()
	return &ID
}
