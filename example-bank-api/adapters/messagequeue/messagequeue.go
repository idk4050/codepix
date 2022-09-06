package messagequeue

import (
	"codepix/example-bank-api/config"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-redis/redis/v9"
	"github.com/phayes/freeport"
)

type MessageQueue struct {
	config  config.Config
	client  *redis.Client
	onClose func() error
	logger  logr.Logger
}

func Open(config config.Config, logger logr.Logger) (*MessageQueue, error) {
	cfg := config.MessageQueue
	logger = logger.WithName("messagequeue")

	var clientOpts *redis.Options
	var onClose func() error

	if cfg.InMemory {
		freePort, err := freeport.GetFreePort()
		if err != nil {
			return nil, fmt.Errorf("open message queue: %w", err)
		}
		port := fmt.Sprint(freePort)

		container, err := exec.Command(
			"podman", "run", "--detach", "--rm", "-p", port+":6379", "docker.io/redis:7.0.4",
		).Output()
		if err != nil {
			return nil, fmt.Errorf("open message queue: %w", err)
		}
		containerID := strings.TrimSpace(string(container))

		clientOpts = &redis.Options{
			Addr: fmt.Sprintf("%s:%s", "localhost", port),
		}
		onClose = func() error {
			return exec.Command("podman", "stop", containerID).Run()
		}
	} else {
		clientOpts = &redis.Options{
			Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
			Username: cfg.User,
			Password: cfg.Password,
		}
		onClose = func() error {
			return nil
		}
	}
	client := redis.NewClient(clientOpts)
	pong, err := client.Ping(context.Background()).Result()
	if err != nil || pong != "PONG" {
		return nil, fmt.Errorf("open message queue: %w", err)
	}
	logger.Info("message queue opened")

	messageQueue := &MessageQueue{
		config:  config,
		client:  client,
		onClose: onClose,
		logger:  logger,
	}
	return messageQueue, nil
}

func (q MessageQueue) Close() error {
	err := q.client.Close()
	if err != nil {
		q.logger.Error(err, "message queue failed to close")
		return err
	}
	err = q.onClose()
	if err != nil {
		q.logger.Error(err, "message queue failed to close")
		return err
	}
	q.logger.Info("message queue closed")
	return nil
}

const messageKey = "message"

func (q MessageQueue) Write(ctx context.Context, message any, streams []string) error {
	messageJson, err := json.Marshal(message)
	if err != nil {
		return err
	}
	pipe := q.client.Pipeline()
	cmds := []*redis.StringCmd{}
	for _, stream := range streams {
		args := &redis.XAddArgs{
			Stream: stream,
			Values: []string{messageKey, string(messageJson)},
		}
		cmd := pipe.XAdd(ctx, args)
		cmds = append(cmds, cmd)
	}
	_, err = pipe.Exec(ctx)

	results := []string{}
	for _, cmd := range cmds {
		results = append(results, cmd.Val())
	}
	if err != nil {
		q.logger.Error(err, "message failed", "streams", streams, "results", results)
		return err
	}
	q.logger.Info("message written", "streams", streams, "results", results)
	return nil
}

func (q MessageQueue) CreateReadGroup(ctx context.Context, stream, group string) error {
	response, err := q.client.XGroupCreateMkStream(ctx, stream, group, "0").Result()
	if err != nil {
		if !strings.HasPrefix(err.Error(), "BUSYGROUP") {
			return fmt.Errorf("create read group: %w", err)
		}
	} else if response != "OK" {
		err := fmt.Errorf("got %s response", response)
		return fmt.Errorf("create read group: %w", err)
	}
	return nil
}

func (q MessageQueue) Ack(ctx context.Context, stream, group string, messageIDs []string) error {
	kvs := []any{
		"stream", stream,
		"group", group,
		"messages", messageIDs,
	}
	if len(messageIDs) == 0 {
		q.logger.Info("no acks received", kvs...)
		return nil
	}
	_, err := q.client.XAck(ctx, stream, group, messageIDs...).Result()
	if err != nil {
		q.logger.Error(err, "ack failed", kvs...)
		return fmt.Errorf("ack messages: %w", err)
	}
	q.logger.Info("messages acked", kvs...)
	return nil
}

type ReadOptions struct {
	Stream        string
	Group         string
	Consumer      string
	MaxPendingAge time.Duration
	BlockDuration time.Duration
}

func Read[T any](q *MessageQueue, ctx context.Context, o ReadOptions) (
	messages []T, messageIDs []string, readErr error,
) {
	defer func() {
		kvs := []any{
			"stream", o.Stream,
			"group", o.Group,
			"consumer", o.Consumer,
			"messages", messageIDs,
		}
		if readErr != nil {
			q.logger.Error(readErr, "read failed", kvs...)
		} else {
			q.logger.Info("messages read", kvs...)
		}
	}()

	pending, _, err := q.client.XAutoClaim(ctx, &redis.XAutoClaimArgs{
		Stream:   o.Stream,
		Group:    o.Group,
		Consumer: o.Consumer,
		Start:    "0",
		MinIdle:  o.MaxPendingAge,
	}).Result()
	if err != nil {
		readErr = fmt.Errorf("get pending messages: %w", err)
		return
	}

	var redisMsgs []redis.XMessage
	if len(pending) > 0 {
		redisMsgs = pending
	} else {
		streamSlices, err := q.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Streams:  []string{o.Stream, ">"},
			Group:    o.Group,
			Consumer: o.Consumer,
			Block:    o.BlockDuration,
		}).Result()
		if err == redis.Nil {
			return
		}
		if err != nil {
			readErr = fmt.Errorf("get messages: %w", err)
			return
		}
		redisMsgs = streamSlices[0].Messages
	}

	for _, redisMsg := range redisMsgs {
		msgJson := redisMsg.Values[messageKey].(string)
		msg := *new(T)
		err := json.Unmarshal([]byte(msgJson), &msg)
		if err != nil {
			readErr = fmt.Errorf("unmarshal message: %w", err)
			return
		}
		messages = append(messages, msg)
		messageIDs = append(messageIDs, redisMsg.ID)
	}
	return
}
