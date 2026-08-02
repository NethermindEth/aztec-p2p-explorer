package turnstile

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Mode string

const (
	ModeEnforce  Mode = "enforce"
	ModeLogOnly  Mode = "log-only"
	ModeDisabled Mode = "disabled"
)

type Config struct {
	Mode               Mode
	SecretKey          string
	AllowedHostnames   map[string]struct{}
	ExpectedAction     string
	Timeout            time.Duration
	CacheTTL           time.Duration
	BindCacheToIPAndUA bool
}

func FromEnv() Config {
	mode := Mode(os.Getenv("TURNSTILE_MODE"))
	if mode == "" {
		mode = ModeDisabled // default to disabled unless configured
	}

	// Default to 1 hour (3600s) to match frontend's 55-minute token lifetime in localStorage
	ttl := 3600 * time.Second
	if v := os.Getenv("TURNSTILE_CACHE_TTL_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			ttl = time.Duration(n) * time.Second
		}
	}

	timeout := 2 * time.Second
	if v := os.Getenv("TURNSTILE_TIMEOUT_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			timeout = time.Duration(n) * time.Millisecond
		}
	}

	allowed := map[string]struct{}{}
	if hs := os.Getenv("TURNSTILE_ALLOWED_HOSTNAMES"); hs != "" {
		for _, h := range strings.Split(hs, ",") {
			h = strings.TrimSpace(h)
			if h != "" {
				allowed[h] = struct{}{}
			}
		}
	}

	return Config{
		Mode:               mode,
		SecretKey:          os.Getenv("TURNSTILE_SECRET"),
		AllowedHostnames:   allowed,
		ExpectedAction:     os.Getenv("TURNSTILE_EXPECTED_ACTION"),
		Timeout:            timeout,
		CacheTTL:           ttl,
		BindCacheToIPAndUA: strings.EqualFold(os.Getenv("TURNSTILE_BIND_CACHE_TO_IP_UA"), "true"),
	}
}
