package secrets

const (
	UserEmailKey    = "email"
	UserPasswordKey = "password"
)

type DexSecret struct {
	UserEmail    string
	UserPassword string
}
