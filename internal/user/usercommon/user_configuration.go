package usercommon

const passwordProtectSaltLen = 32
const passwordProtectKeyLen = 32
const passwordProtectTime = 3
const passwordProtectMemory = 32 * 1024
const passwordProtectThreads = 4

// Config provides access to the configuration used by the user
// related functionalities.
type Config struct {
	SaltLen         int    `configkey:"app.user.password_salt_len" default:"32" validate:"min=32"`
	KeyLen          uint32 `configkey:"app.user.password_key_len" default:"32" validate:"min=32"`
	PasswordTime    uint32 `configkey:"app.user.password_time" default:"3" validate:"min=3"`
	PasswordMemory  uint32 `configkey:"app.user.password_memory" default:"32768"`
	PasswordThreads uint8  `configkey:"app.user.password_threads" default:"4"`
}

// DefaultUserConfiguration returns a user configuration with default values.
// Mainly intended to be used for testing purpose.
func DefaultUserConfiguration() *Config {
	return &Config{
		SaltLen:         passwordProtectSaltLen,
		KeyLen:          passwordProtectKeyLen,
		PasswordTime:    passwordProtectTime,
		PasswordMemory:  passwordProtectMemory,
		PasswordThreads: passwordProtectThreads,
	}
}
