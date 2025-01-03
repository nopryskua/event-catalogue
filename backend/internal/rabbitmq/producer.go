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
	cl      *rabbitMQ
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

	if err := p.cl.Publish(b); err != nil {
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
	p.m.Lock()
	defer p.m.Unlock()

	if p.cl != nil {
		return nil
	}

	result, err := newRabbitMQ(p.url, util.TypeName[T](), false)
	if err != nil {
		p.metrics.ErrorClientInitCount++
		return err
	}

	p.cl = result
	p.metrics.SuccessClientInitCount++

	return nil
}

func (p *rabbitMQProducer[T]) closeClient() {
	p.m.Lock()
	defer p.m.Unlock()

	p.metrics.CloseClientCount++

	if p.cl == nil {
		return
	}

	c := p.cl
	defer c.Close()

	p.cl = nil
}
