package kafka

import (
	"github.com/shopify/sarama"
)

type (
	Kafka struct {
		ClientName    string
		ClientUrls    []string
		Consumer      sarama.Consumer
		ConsumerGroup sarama.Consumer
		IsTlsEnabled  bool
		CaCert        []byte
		ClientCert    []byte
		ClientKey     []byte
	}
)

func NewConnector(Options Connector.Options) connector.Connector {
	return &kafka{
		ClientName:      Options.ClientName,
		ClientUrls:      Options.CleintUrls,
		IsTlsEnabled:    Options.IsTlsEnabled,
		CaCert:          Options.ca,
		CleintCert:      options.cert,
		IsConsumerGroup: options.IsConsumerGroup,
	}
}

func (k *Kafka) Connect(Options Connector.Options) (e error) {
	config := sarama.NewConfig()
	config.Version = Options.ClientName

}
func (k *Kafka) Disconnect(Options Connector.Options) (e error) {
	if e := k.stopSubscriber(); e != nil {
		return e
	}

}
