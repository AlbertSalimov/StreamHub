package events

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/AlbertSalimov/streamhub/internal/content"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const streamName = "CONTENT_EVENTS"

type Publisher struct {
	nc *nats.Conn
	js jetstream.JetStream
}

func NewPublisher(ctx context.Context, url string) (*Publisher, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("connect to NATS: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("create JetStream client: %w", err)
	}

	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     streamName,
		Subjects: []string{"content.>"},
	})
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("create JetStream stream: %w", err)
	}

	return &Publisher{
		nc: nc,
		js: js,
	}, nil
}

func (p *Publisher) Publish(ctx context.Context, event content.Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	_, err = p.js.Publish(ctx, event.Type, data)
	if err != nil {
		return fmt.Errorf("publish event %q: %w", event.Type, err)
	}

	return nil
}

func (p *Publisher) Close() {
	if p.nc != nil {
		p.nc.Close()
	}
}
