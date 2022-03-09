package connector

import (
	"time"
)

type (
	Connector interface {
		Connect() error
		Disconnect() error
		Subscribe(topicName string,
			handle func(message interface{}, subject string),
			opts *SubscriptionOptions),decode Decode) (msg interface{},error)
		GetTopics() ([]string, error)
		Publish(TopicName string, message interface{})
		Rpc(PublishTopicName string,
			PublishMessage interface{},
			SubscribeTopicName string,
			handle  func(message interface{}, Subject string),
			decode Decode,
			Opts *SubscriptionOptions)
		Configure(ctx interface{})
	}

	Options struct {
		ClientName      string
		ServerUrls      []string
		IsTlsEnabled    bool
		Ca              []byte
		ClientCert      []byte
		IsConsumerGroup bool
	}

	SubscriptionOptions struct {
		GetRaw      bool
		StartAtTime time.Time
	}
)


type Decode func(handleMsg func(message interface{},topicName string),msg []byte,topicName string)error