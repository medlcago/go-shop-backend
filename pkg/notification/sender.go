package notification

import "context"

type Sender interface {
	Channel() Channel
	Send(ctx context.Context, n Notification) error
}

type SenderRegistry interface {
	For(ch Channel) (Sender, bool)
}

type MapRegistry map[Channel]Sender

func (r MapRegistry) For(ch Channel) (Sender, bool) {
	s, ok := r[ch]
	return s, ok
}
