package projection

import (
	"codepix/bank-api/adapters/projectionclient"
	"codepix/bank-api/transaction/read/repository"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/looplab/eventhorizon/repo/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	opts "go.mongodb.org/mongo-driver/mongo/options"
)

type Projection struct {
	Repo *mongodb.Repo
}

var _ repository.Repository = Projection{}

func (p Projection) Find(ctx context.Context, ID uuid.UUID) (*repository.Transaction, error) {
	entity, err := p.Repo.Find(ctx, ID)
	transaction, _ := entity.(*repository.Transaction)
	return transaction, projectionclient.MapError(err, repository.EntityType)
}

func (p Projection) List(ctx context.Context, options repository.ListOptions,
) ([]repository.ListItem, error) {
	entities, err := p.Repo.FindCustom(ctx, func(ctx context.Context, c *mongo.Collection,
	) (*mongo.Cursor, error) {
		filter := bson.D{
			{"created_at", bson.D{{"$gte", options.CreatedAfter.Truncate(time.Millisecond)}}},
		}
		if options.SenderID != uuid.Nil {
			filter = append(filter, bson.E{"sender", options.SenderID.String()})
		}
		if options.ReceiverID != uuid.Nil {
			filter = append(filter, bson.E{"receiver", options.ReceiverID.String()})
		}
		opts := opts.Find().
			SetSort(bson.D{{"created_at", -1}}).
			SetLimit(int64(options.Limit)).
			SetSkip(int64(options.Skip))
		return c.Find(ctx, filter, opts)
	})

	transactions := []repository.ListItem{}
	for _, entity := range entities {
		transaction, _ := entity.(*repository.Transaction)
		transactions = append(transactions, *transaction)
	}
	return transactions, projectionclient.MapError(err, repository.EntityType)
}
