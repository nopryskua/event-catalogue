package producer

import (
	"github.com/nopryskua/event-catalogue/backend/internal/task"
)

type T[M task.T] interface {
	Produce(M) error
	Close()
}
