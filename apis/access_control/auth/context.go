package auth

import "context"

const (
	ContextKeyInitMainUser = "init_main_user"
)

// IsInitMainUser .
func IsInitMainUser(ctx context.Context) bool {
	val := ctx.Value(ContextKeyInitMainUser)
	ret, _ := val.(bool)
	return ret
}
