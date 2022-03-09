package connector

import (
	"time"
)

type (
	Connector interface {
		Connect() error
		Disconnect() error
		Subscribe(topicName string,
			decode func(message interface{}, subject string),
			opts *SubscriptionOptions) error
		GetTopics() ([]string, error)
		Publish(TopicName string, message interface{})
		Rpc(PublishTopicName string,
			PublishMessage interface{},
			SubscribeTopicName string,
			Decode func(message interface{}, Subject string),
			Opts *SubscriptionOptions)
		Configure(ctx interface{})
	}

	Options struct {
		ClientName   string
		ServerUrls   []string
		IsTlsEnabled bool
	}

	SubscriptionOptions struct {
		GetRaw      bool
		StartAtTime time.Time
	}
)
