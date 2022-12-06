package slack

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
	"github.com/target/goalert/permission"
)

// User will lookup a single Slack user
func (s *ChannelSender) User(ctx context.Context, id string) (*User, error) {
	err := permission.LimitCheckAny(ctx, permission.User, permission.System)
	if err != nil {
		return nil, err
	}

	// call slack api with team:name id and get user info to return
	var usr *slack.User
	err = s.withClient(ctx, func(c *slack.Client) error {
		usr, err = c.GetUserInfoContext(ctx, id)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	return &User{
		ID:     usr.ID,
		Name:   usr.Name,
		TeamID: usr.TeamID,
	}, nil
}

func (s *ChannelSender) Users(ctx context.Context, search string) ([]*User, error) {
	err := permission.LimitCheckAny(ctx, permission.User, permission.System)
	if err != nil {
		return nil, err
	}

	// call slack api with team:name id and get user info to return
	var usr *[]slack.User
	err = s.withClient(ctx, func(c *slack.Client) error {
		usr, err = c.GetUsersInfoContext(ctx, "")
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	return nil, nil
}
