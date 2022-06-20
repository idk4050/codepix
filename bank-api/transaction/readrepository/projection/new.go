package projection

import (
	"codepix/bank-api/adapters/projectionclient"
	"codepix/bank-api/transaction"
	"codepix/bank-api/transaction/readrepository"
	"fmt"

	"github.com/looplab/eventhorizon"
)

func New(client *projectionclient.StoreProjection) (*Projection, error) {
	projector := &Projector{}
	entityType := func() eventhorizon.Entity {
		return &readrepository.Transaction{}
	}
	projection, err := client.Connect(
		readrepository.RepositoryType,
		entityType,
		projector,
		eventhorizon.MatchAggregates{transaction.AggregateType},
	)
	if err != nil {
		return nil, fmt.Errorf("new Projection: %w", err)
	}
	return &Projection{projection}, nil
}
