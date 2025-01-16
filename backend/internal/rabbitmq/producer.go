package rabbitmq

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/nopryskua/event-catalogue/backend/internal/producer"
	"github.com/nopryskua/event-catalogue/backend/internal/task"
	"github.com/nopryskua/event-catalogue/backend/internal/util"
)

type rabbitMQProducer[T task.T] struct {
	url     string
	m       sync.RWMutex
	cl      rabbitMQWithLock
	metrics producer.Metrics
}

// NewProducer creates a RabbitMQ message producer.
//
// It executes basic input validation yet does
// not retry an error in case the connection is
// not established because the connection may
// be established later.
func NewProducer[T task.T](url string) (producer.T[T], error) {
	if url == "" {
		return nil, errors.New("URL must be set")
	}

	return &rabbitMQProducer[T]{
		url: url,
	}, nil
}

func (p *rabbitMQProducer[T]) Produce(t T) error {
	if err := p.initClient(); err != nil {
		return err
	}

	b, err := json.Marshal(t)
	if err != nil {
		return err
	}

	if err := p.publish(b); err != nil {
		p.closeClient()

		return err
	}

	p.m.Lock()
	p.metrics.SuccessProduceCount++
	p.m.Unlock()

	return nil
}

func (p *rabbitMQProducer[T]) Close() {
	p.closeClient()
}

func (p *rabbitMQProducer[T]) Metrics() producer.Metrics {
	p.m.RLock()
	defer p.m.RUnlock()

	return p.metrics
}

func (p *rabbitMQProducer[T]) initClient() error {
	p.cl.Lock()
	defer p.cl.Unlock()

	if p.cl.rabbitMQ != nil {
		return nil
	}

	result, err := newRabbitMQ(p.url, util.TypeName[T](), false)
	if err != nil {
		p.metrics.ErrorClientInitCount++
		return err
	}

	p.cl.rabbitMQ = result

	p.m.Lock()
	p.metrics.SuccessClientInitCount++
	p.m.Unlock()

	return nil
}

func (p *rabbitMQProducer[T]) closeClient() {
	p.cl.Lock()
	defer p.cl.Unlock()

	p.m.Lock()
	p.metrics.CloseClientCount++
	p.m.Unlock()

	if p.cl.rabbitMQ == nil {
		return
	}

	c := p.cl.rabbitMQ
	defer c.Close()

	p.cl.rabbitMQ = nil
}

func (p *rabbitMQProducer[T]) publish(b []byte) error {
	p.cl.RLock()
	defer p.cl.RUnlock()

	if p.cl.rabbitMQ == nil {
		return errors.New("client closed")
	}

	return p.cl.Publish(b)
}
