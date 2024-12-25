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

	world := "World"

	var taskExecuted sync.WaitGroup
	taskExecuted.Add(1)

	run = func(name string) error {
		defer taskExecuted.Done()

		require.Equal(t, world, name)

		return nil
	}

	err = p.Produce(Task{
		Name: world,
	})
	require.NoError(t, err)

	c, err := rabbitmq.NewConsumer[Task](url, 1)
	require.NoError(t, err)

	var gracefullyClosed sync.WaitGroup
	gracefullyClosed.Add(1)

	go func() {
		defer gracefullyClosed.Done()
		c.Consume()
	}()

	taskExecuted.Wait()

	c.Close()
	p.Close()

	gracefullyClosed.Wait()

	m := p.Metrics()
	require.Equal(t, 1, m.SuccessClientInitCount)
	require.Equal(t, 0, m.ErrorClientInitCount)
	require.Equal(t, 1, m.CloseClientCount)
	require.Equal(t, 1, m.SuccessProduceCount)
}
