package rabbitmq

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type rabbitMQ struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	q    amqp.Queue
}

func newRabbitMQ(url string, queueName string, consumer bool) (*rabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	if consumer {
		if err := ch.Qos(
			1,
			0,
			false,
		); err != nil {
			return nil, err
		}
	}

	q, err := ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &rabbitMQ{
		conn: conn,
		ch:   ch,
		q:    q,
	}, nil
}

func (r *rabbitMQ) Publish(body []byte) error {
	return r.ch.Publish(
		"",
		r.q.Name,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         []byte(body),
		})
}

func (r *rabbitMQ) Consume() (<-chan amqp.Delivery, error) {
	return r.ch.Consume(
		r.q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
}

func (r *rabbitMQ) Close() {
	defer r.conn.Close()
	defer r.ch.Close()
}
