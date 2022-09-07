package queue

import (
	"codepix/example-bank-api/adapters/messagequeue"
	"codepix/example-bank-api/customer/signup"
	"codepix/example-bank-api/customer/signup/repository"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func SetupReaders(ctx context.Context, messageQueue *messagequeue.MessageQueue,
	repository repository.Repository,
) error {
	err := setup(ctx, messageQueue, StartedStream, "repository", func(message Started) error {
		signUp := signup.SignUp{
			Name:  message.Name,
			Email: message.Email,
			Token: message.Token,
		}
		return repository.Add(signUp)
	})
	if err != nil {
		return err
	}
	err = setup(ctx, messageQueue, StartedStream, "mailer", func(message Started) error {
		fmt.Println(message.Email, message.Token)
		return nil
	})
	if err != nil {
		return err
	}
	err = setup(ctx, messageQueue, FinishedStream, "repository", func(message Finished) error {
		return repository.Remove(message.Token)
	})
	if err != nil {
		return err
	}
	return nil
}

func setup[T any](ctx context.Context, messageQueue *messagequeue.MessageQueue,
	stream, group string, handler func(message T) error,
) error {
	err := messageQueue.CreateReadGroup(ctx, stream, group)
	if err != nil {
		return err
	}
	randomID := uuid.NewString()
	consumer := randomID[:8]
	go func() {
		options := messagequeue.ReadOptions{
			Stream:        stream,
			Group:         group,
			Consumer:      consumer,
			MaxPendingAge: time.Second,
			BlockDuration: 0,
		}
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			messages, messageIDs, err := messagequeue.Read[T](messageQueue, ctx, options)
			if err != nil {
				continue
			}
			acks := []string{}
			for i, message := range messages {
				err := handler(message)
				if err != nil {
					continue
				}
				acks = append(acks, messageIDs[i])
			}
			if len(acks) == 0 {
				continue
			}
			err = messageQueue.Ack(ctx, stream, group, acks)
			if err != nil {
				continue
			}
		}
	}()
	return nil
}
