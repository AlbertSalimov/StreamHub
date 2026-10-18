package events

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/AlbertSalimov/streamhub/internal/content"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func RunConsumer(ctx context.Context, url string) error {
	nc, err := nats.Connect(url)
	if err != nil {
		return err
	}
	defer nc.Close()

	js, err := jetstream.New(nc)
	if err != nil {
		return err
	}

	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     streamName,
		Subjects: []string{"content.>"},
	})
	if err != nil {
		return err
	}

	consumer, err := js.CreateOrUpdateConsumer(ctx, streamName, jetstream.ConsumerConfig{
		Durable:       "content-events-logger",
		AckPolicy:     jetstream.AckExplicitPolicy,
		FilterSubject: "content.>",
	})
	if err != nil {
		return err
	}

	messages, err := consumer.Messages()
	if err != nil {
		return err
	}
	defer messages.Stop()

	slog.Info("NATS consumer started", "durable", "content-events-logger")

	for {
		if ctx.Err() != nil {
			return nil
		}

		msg, err := messages.Next(jetstream.NextMaxWait(1 * time.Second))
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			continue
		}

		var event content.Event
		if err := json.Unmarshal(msg.Data(), &event); err != nil {
			slog.Error("invalid content event", "error", err)
			_ = msg.Term()
			continue
		}

		slog.Info(
			"content event received",
			"event_type", event.Type,
			"content_id", event.ContentID,
			"event_id", event.ID,
		)

		if err := msg.Ack(); err != nil {
			slog.Error("failed to acknowledge event", "error", err)
		}
	}
}
