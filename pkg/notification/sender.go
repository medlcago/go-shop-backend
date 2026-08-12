package notification

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var (
	ErrSenderNotFound           = errors.New("sender not found")
	ErrChannelAlreadyRegistered = errors.New("channel already registered")
)

type Sender interface {
	Channel() Channel
	Send(ctx context.Context, n Notification) error
}

type Registry interface {
	Get(ch Channel) (Sender, error)
	Register(ch Channel, s Sender)
}

type registry struct {
	senders map[Channel]Sender
	mu      sync.RWMutex
}

func NewRegistry() *registry {
	return &registry{
		senders: make(map[Channel]Sender),
	}
}

func (r *registry) Get(ch Channel) (Sender, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sender, ok := r.senders[ch]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrSenderNotFound, ch)
	}

	return sender, nil
}

func (r *registry) Register(ch Channel, s Sender) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.senders[ch]; ok {
		panic(fmt.Errorf("%w: %s", ErrChannelAlreadyRegistered, ch))
	}

	r.senders[ch] = s
}
