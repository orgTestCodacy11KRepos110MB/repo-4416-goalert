package slack

import (
	"context"

	"github.com/target/goalert/notification"
)

type DMSender struct {
	*ChannelSender
}

var _ notification.FriendlyValuer = (*DMSender)(nil)

func (s *ChannelSender) DMSender() *DMSender {
	return &DMSender{s}
}

// FriendlyValue implements notification.FriendlyValuer.
func (s *DMSender) FriendlyValue(ctx context.Context, id string) (string, error) {
	usr, err := s.User(ctx, id)
	if err != nil {
		return "", err
	}

	return "@" + usr.Name, nil
}
