package channels

import (
	"context"
	"fmt"
	"job-forge/internal/domain"
)

type Handler interface {
	Handle(ctx context.Context, job domain.Job) error
}

type Router struct {
	handlers map[string]Handler
}

func NewRouter(handlers map[string]Handler) *Router {
	return &Router{handlers: handlers}
}

func (r *Router) Handle(ctx context.Context, job domain.Job) error {
	h, ok := r.handlers[job.Channel]
	if !ok {
		return fmt.Errorf("no handler for channel=%q", job.Channel)
	}
	return h.Handle(ctx, job)
}
