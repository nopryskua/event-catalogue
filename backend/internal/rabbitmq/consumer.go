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
	url    string
	m      sync.Mutex
	cls    []rabbitMQWithLock
	closed bool
}

type rabbitMQWithLock struct {
	*rabbitMQ
	sync.Mutex
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

	msgs, err := c.cls[i].Consume()
	if err != nil {
		return err
	}

	for msg := range msgs {
		ack := c.run(msg)
		if ack {
			if err := msg.Ack(false); err != nil {
				log.Println(err.Error())
			}

			continue
		}

		if err := msg.Nack(false, true); err != nil {
			log.Println(err.Error())
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

func (p *rabbitMQConsumer[T]) initClient(i int) error {
	p.cls[i].Lock()
	defer p.cls[i].Unlock()

	if p.cls[i].rabbitMQ != nil {
		return nil
	}

	result, err := newRabbitMQ(p.url, util.TypeName[T](), true)
	if err != nil {
		return err
	}

	p.cls[i].rabbitMQ = result

	return nil
}

func (c *rabbitMQConsumer[T]) closeClient(i int) {
	c.cls[i].Lock()
	defer c.cls[i].Unlock()

	cl := c.cls[i].rabbitMQ
	if cl == nil {
		return
	}

	defer cl.Close()

	c.cls[i].rabbitMQ = nil
}
