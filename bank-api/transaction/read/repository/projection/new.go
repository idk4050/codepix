package projection

import (
	"codepix/bank-api/adapters/projectionclient"
	"codepix/bank-api/transaction"
	"codepix/bank-api/transaction/read/repository"
	"fmt"

	"github.com/looplab/eventhorizon"
)

func New(client *projectionclient.StoreProjection) (*Projection, error) {
	projector := &Projector{}
	entityType := func() eventhorizon.Entity {
		return &repository.Transaction{}
	}
	projection, err := client.Setup(
		projector.ProjectorType(),
		entityType,
		projector,
		transaction.AggregateType,
	)
	if err != nil {
		return nil, fmt.Errorf("new Projection: %w", err)
	}
	return &Projection{projection}, nil
}
