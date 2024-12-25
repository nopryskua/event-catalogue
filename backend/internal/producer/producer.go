package producer

import (
	"github.com/nopryskua/event-catalogue/backend/internal/task"
)

type T[M task.T] interface {
	Produce(M) error
	Close()
	Metrics() Metrics
}

type Metrics struct {
	SuccessClientInitCount int
	ErrorClientInitCount   int
	CloseClientCount       int
	SuccessProduceCount    int
}
