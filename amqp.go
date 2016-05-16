package main

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

/*
 * AMQP bookkeeping
 */

func NewConsumer(amqpURI, queueName, ctag string, routingKeys []string) (*Consumer, <-chan amqp.Delivery, error) {
	c := &Consumer{
		conn:    nil,
		channel: nil,
		tag:     ctag,
	}

	//exchange := "api.events"
	exchange := "api.rawsensordata"

	//exchangeType := "topic"
	exchangeType := "fanout"

	var err error

	log.Println("Connecting to message bus...")
	log.Printf("Connecting to exchange: %s", exchange)
	c.conn, err = amqp.Dial(amqpURI)

	if err != nil {
		return nil, nil, fmt.Errorf("Dial: %s", err)
	}

	c.channel, err = c.conn.Channel()
	if err != nil {
		return nil, nil, fmt.Errorf("Channel: %s", err)
	}

	err = c.channel.ExchangeDeclare(
		exchange,
		exchangeType,
		true,  // durable
		false, // delete when complete
		false, // internal
		false, // nowait
		nil,   // arguments
	)
	if err != nil {
		return nil, nil, fmt.Errorf("Exchange declare: %s", err)
	}

	queue, err := c.channel.QueueDeclare(
		queueName,
		false, // durable,
		true,  // delete when unused
		false, // exclusive
		false, // nowait
		nil,   // arguments
	)
	if err != nil {
		return nil, nil, fmt.Errorf("Queue declare: %s", err)
	}

	err = c.channel.QueueBind(queueName, "#", exchange, false, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("QueueBind: %s", err)
	}

	/*
	for _, key := range routingKeys {
		err = c.channel.QueueBind(queueName, "#", exchange, false, nil)
		//err = c.channel.QueueBind(queueName, key, exchange, false, nil)
		if err != nil {
			return nil, nil, fmt.Errorf("QueueBind: %s", err)
		}
		log.Printf("Subscribed to %s", key)
	}
	*/

	deliveries, err := c.channel.Consume(
		queue.Name,
		c.tag,
		true,  // automatic ack
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // arguments
	)

	if err != nil {
		return nil, nil, fmt.Errorf("Queue consume: %s", err)
	}

	return c, deliveries, nil
}

func (c *Consumer) Shutdown() error {
	if err := c.channel.Cancel(c.tag, true); err != nil {
		return fmt.Errorf("Consumer cancel failed: %s", err)
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("AMQP connection close error: %s", err)
	}

	defer log.Printf("AMQP shutdown OK")

	return nil
}
