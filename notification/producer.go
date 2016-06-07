package notification

import (
	"time"

	"github.com/Shopify/sarama"
	"github.com/Sirupsen/logrus"
)

// Producer interface
type Producer interface {
	SendEvent(topic, key, value string) error
	Close() error
}

// SASL authentication object
type SASL struct {
	Enable   bool
	User     string
	Password string
}

// ProducerConfig for new Producer objects
type ProducerConfig struct {
	ClientID string
	Brokers  []string
	SASL     SASL
}

type producer struct {
	producer sarama.SyncProducer
}

// NewProducer instantiates a producer
func NewProducer(conf ProducerConfig) (Producer, error) {
	config := sarama.NewConfig()
	config.Net.DialTimeout = time.Second * 60
	if conf.SASL.Enable {
		config.Net.TLS.Enable = true
		config.ClientID = conf.ClientID
		config.Net.SASL.User = conf.SASL.User
		config.Net.SASL.Password = conf.SASL.Password
		config.Net.SASL.Enable = conf.SASL.Enable
	}
	config.Producer.Partitioner = sarama.NewManualPartitioner

	saramaProducer, err := sarama.NewSyncProducer(conf.Brokers, config)

	p := &producer{
		producer: saramaProducer,
	}
	if err != nil {
		return nil, err
	}

	return p, nil
}

// SendEvent produces an event on the topic.
func (p *producer) SendEvent(topic, key, value string) error {
	msg := &sarama.ProducerMessage{
		Topic:     topic,
		Key:       sarama.StringEncoder(key),
		Value:     sarama.StringEncoder(value),
		Partition: 0,
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"partition": partition,
		"offset":    offset,
		"topic":     topic,
		"key":       key,
		"value":     value,
	}).Debug("Message sent")

	return nil
}

// Close the producer and free up resources.
func (p *producer) Close() error {
	if p.producer != nil {
		return p.producer.Close()
	}

	return nil
}
