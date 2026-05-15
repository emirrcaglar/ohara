package handler

import (
	"context"
	"ohara/src/internal/db"
)

type contextKey string

const UserKey contextKey = "user"

func GetUser(ctx context.Context) *db.User {
	user, ok := ctx.Value(UserKey).(*db.User)
	if !ok {
		return nil
	}
	return user
}
