// Copyright 2016 IBM Corporation
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package checker

import (
	"errors"
	"time"

	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
)

//Consumer interface
type Consumer interface {
	ReceiveEvent() (string, string, error)
	Close() error
}

type consumer struct {
	consumer     sarama.Consumer
	partConsumer sarama.PartitionConsumer
}

// ConsumerConfig config for new Consumer objects
type ConsumerConfig struct {
	Brokers     []string
	Username    string
	Password    string
	ClientID    string
	Topic       string
	SASLEnabled bool
}

// NewConsumer returns a new Consumer
func NewConsumer(conf ConsumerConfig) (Consumer, error) {
	c := new(consumer)

	config := sarama.NewConfig()
	config.Net.DialTimeout = time.Second * 60
	if conf.SASLEnabled {
		config.Net.TLS.Enable = true
		config.Net.SASL.User = conf.Username
		config.Net.SASL.Password = conf.Password
		config.Net.SASL.Enable = conf.SASLEnabled
		config.ClientID = conf.ClientID
	}

	var err error
	c.consumer, err = sarama.NewConsumer(conf.Brokers, config)
	if err != nil {
		return nil, err
	}

	c.partConsumer, err = c.consumer.ConsumePartition(conf.Topic, 0, sarama.OffsetNewest)
	if err != nil {
		return nil, err
	}

	return c, nil

}

// ReceiveEvent blocks until an event is received
func (c *consumer) ReceiveEvent() (string, string, error) {
	// Wait for event
	select {
	case msg := <-c.partConsumer.Messages():
		// Should never get a nil, but we check to avoid a panic.
		if msg != nil {
			log.WithFields(log.Fields{
				"key":   string(msg.Key),
				"value": string(msg.Value),
			}).Debug("Received message")
			return string(msg.Key), string(msg.Value), nil
		}
		return "", "", errors.New("Kafka provided a nil message")

	case err := <-c.partConsumer.Errors():
		return "", "", err
	}

}

// Close the consumer.  Required to prevent memory leaks
func (c *consumer) Close() error {
	var err error
	if c.partConsumer != nil {
		err = c.partConsumer.Close()
		if err != nil {
			return err
		}
	}

	if c.consumer != nil {
		err = c.consumer.Close()
	}
	return err
}
