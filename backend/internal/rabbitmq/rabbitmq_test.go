package rabbitmq_test

import (
	"sync"
	"testing"

	"github.com/nopryskua/event-catalogue/backend/internal/rabbitmq"
	"github.com/stretchr/testify/require"
)

var run func(string) error

type Task struct {
	Name string
}

func (t Task) Run() error {
	return run(t.Name)
}

func TestRabbitMQ(t *testing.T) {
	url := "amqp://guest:guest@rabbitmq:5672/"

	p, err := rabbitmq.NewProducer[Task](url)
	require.NoError(t, err)
	defer p.Close()

	world := "World"

	var wg sync.WaitGroup
	wg.Add(1)

	run = func(name string) error {
		defer wg.Done()

		require.Equal(t, world, name)

		return nil
	}

	err = p.Produce(Task{
		Name: world,
	})
	require.NoError(t, err)

	c, err := rabbitmq.NewConsumer[Task](url, 1)
	require.NoError(t, err)

	go func() {
		c.Consume()
	}()

	wg.Wait()

	c.Close()
}
