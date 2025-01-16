package rabbitmq

import (
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/nopryskua/event-catalogue/backend/internal/consumer"
	"github.com/nopryskua/event-catalogue/backend/internal/task"
	"github.com/nopryskua/event-catalogue/backend/internal/util"
	amqp "github.com/rabbitmq/amqp091-go"
)

type rabbitMQConsumer[T task.T] struct {
	url     string
	m       sync.RWMutex
	cls     []rabbitMQWithLock
	closed  bool
	metrics consumer.Metrics
}

type rabbitMQWithLock struct {
	*rabbitMQ
	sync.RWMutex
}

// NewConsumer creates a RabbitMQ message consumer.
//
// It executes basic input validation yet does
// not retry an error in case the connection is
// not established because the connection may
// be established later.
func NewConsumer[T task.T](url string, parallelism int) (consumer.T[T], error) {
	if url == "" {
		return nil, errors.New("URL should be set")
	}

	if parallelism < 1 {
		return nil, errors.New("parallelism should be at least 1")
	}

	return &rabbitMQConsumer[T]{
		url: url,
		cls: make([]rabbitMQWithLock, parallelism),
	}, nil
}

func (c *rabbitMQConsumer[T]) Consume() {
	var wg sync.WaitGroup

	wg.Add(len(c.cls))

	for i := 0; i < len(c.cls); i++ {
		go c.consumeWithRetry(i, &wg)
	}

	wg.Wait()
}

func (c *rabbitMQConsumer[T]) Close() {
	c.m.Lock()
	if c.closed {
		c.m.Unlock()
		return
	}

	c.closed = true
	c.m.Unlock()

	for i := 0; i < len(c.cls); i++ {
		c.closeClient(i)
	}
}

func (c *rabbitMQConsumer[T]) Metrics() consumer.Metrics {
	c.m.RLock()
	defer c.m.RUnlock()

	return c.metrics
}

func (c *rabbitMQConsumer[T]) consumeWithRetry(i int, wg *sync.WaitGroup) {
	defer wg.Done()

	util.Loop(func() bool {
		if err := c.consume(i); err != nil {
			log.Println(err)
		}

		c.m.Lock()
		defer c.m.Unlock()

		return !c.closed
	}, 5*time.Second)
}

func (c *rabbitMQConsumer[T]) consume(i int) error {
	if err := c.initClient(i); err != nil {
		return err
	}
	defer c.closeClient(i)

	msgs, err := c.startConsume(i)
	if err != nil {
		return err
	}

	for msg := range msgs {
		ack := c.run(msg)
		if ack {
			if err := msg.Ack(false); err != nil {
				log.Println(err.Error())
			} else {
				c.m.Lock()
				c.metrics.AckCount++
				c.m.Unlock()
			}

			continue
		}

		if err := msg.Nack(false, true); err != nil {
			log.Println(err.Error())
		} else {
			c.m.Lock()
			c.metrics.NackCount++
			c.m.Unlock()
		}
	}

	return nil
}

func (c *rabbitMQConsumer[T]) run(msg amqp.Delivery) (ack bool) {
	var t T
	if err := json.Unmarshal([]byte(msg.Body), &t); err != nil {
		log.Println(err.Error())
		return true
	}

	if err := t.Run(); err != nil {
		log.Println(err.Error())
		return false
	}

	return true
}

func (c *rabbitMQConsumer[T]) initClient(i int) error {
	c.cls[i].Lock()
	defer c.cls[i].Unlock()

	if c.cls[i].rabbitMQ != nil {
		return nil
	}

	result, err := newRabbitMQ(c.url, util.TypeName[T](), true)
	if err != nil {
		c.m.Lock()
		c.metrics.ErrorClientInitCount++
		c.m.Unlock()

		return err
	}

	c.m.Lock()
	c.metrics.SuccessClientInitCount++
	c.m.Unlock()

	c.cls[i].rabbitMQ = result

	return nil
}

func (c *rabbitMQConsumer[T]) closeClient(i int) {
	c.cls[i].Lock()
	defer c.cls[i].Unlock()

	cl := c.cls[i].rabbitMQ
	if cl == nil {
		return
	}

	c.m.Lock()
	c.metrics.CloseClientCount++
	c.m.Unlock()

	defer cl.Close()

	c.cls[i].rabbitMQ = nil
}

func (c *rabbitMQConsumer[T]) startConsume(i int) (<-chan amqp.Delivery, error) {
	c.cls[i].RLock()
	defer c.cls[i].RUnlock()

	msgs, err := c.cls[i].Consume()
	if err != nil {
		return nil, err
	}

	return msgs, nil
}
