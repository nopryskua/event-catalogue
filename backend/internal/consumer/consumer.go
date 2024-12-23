package consumer

import "github.com/nopryskua/event-catalogue/backend/internal/task"

type T[M task.T] interface {
	// Consume runs task processing
	// and blocks until Close is called.
	// The function should be called only once.
	Consume()
	Close()
}
