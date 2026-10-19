package core_auth

import "context"

type userIDContextKey struct{}

var key = userIDContextKey{}

func ToContext(ctx context.Context, userID int) context.Context {
	return context.WithValue(
		ctx,
		key,
		userID,
	)
}

func UserIDFromContext(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(key).(int)
	if !ok {
		return 0, false
	}

	return userID, true
}
