package projection

import (
	"codepix/bank-api/adapters/projectionclient"
	"codepix/bank-api/transaction/readrepository"
	"context"

	"github.com/google/uuid"
	"github.com/looplab/eventhorizon"
)

type Projection struct {
	eventhorizon.ReadWriteRepo
}

var _ readrepository.ReadRepository = Projection{}

func (p Projection) Find(ID uuid.UUID) (*readrepository.Transaction, error) {
	entity, err := p.ReadWriteRepo.Find(context.Background(), ID)
	transaction, _ := entity.(*readrepository.Transaction)
	return transaction, projectionclient.MapError(err, readrepository.EntityType)
}
