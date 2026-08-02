package turnstile

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"
)

type Result struct {
	Token       string
	Hostname    string
	Action      string
	ChallengeTS time.Time
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func clientIP(r *http.Request) string {
	if ip := strings.TrimSpace(r.Header.Get("CF-Connecting-IP")); ip != "" {
		return ip
	}
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return ""
}

func isPublicPath(p string) bool {
	// Health checks and monitoring
	if p == "/api/healthz" || p == "/health" || p == "/metrics" {
		return true
	}

	// Documentation
	if strings.HasPrefix(p, "/swagger/") {
		return true
	}

	// Private API endpoints that should bypass Turnstile
	if p == "/api/private/peers" {
		return true
	}

	return false
}

type Middleware struct {
	Config   Config
	Verifier Verifier
	Cache    Cache
}

func (m *Middleware) extractToken(r *http.Request) string {
	token := r.Header.Get("X-Turnstile-Token")
	if token != "" {
		return token
	}
	if strings.Contains(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		_ = r.ParseForm()
		return r.FormValue("cf-turnstile-response")
	}
	return ""
}

func (m *Middleware) verifyAndCache(ctx context.Context, token, ip, ua string) (int, map[string]any) {
	vr, err := m.Verifier.Verify(ctx, token, ip)
	if err != nil {
		return http.StatusBadGateway, map[string]any{"error": "turnstile_unavailable"}
	}
	if !vr.Success {
		return http.StatusUnauthorized, map[string]any{"error": "turnstile_failed", "codes": vr.ErrorCodes}
	}
	if len(m.Config.AllowedHostnames) > 0 {
		if _, ok := m.Config.AllowedHostnames[vr.Hostname]; !ok {
			return http.StatusForbidden, map[string]any{"error": "hostname_mismatch", "got": vr.Hostname}
		}
	}
	if m.Config.ExpectedAction != "" && vr.Action != "" && vr.Action != m.Config.ExpectedAction {
		return http.StatusForbidden, map[string]any{"error": "action_mismatch", "got": vr.Action}
	}
	ts, _ := time.Parse(time.RFC3339Nano, vr.ChallengeTS)
	exp := time.Now().Add(m.Config.CacheTTL)
	m.Cache.Set(token, CachedResult{
		Valid:       true,
		Hostname:    vr.Hostname,
		Action:      vr.Action,
		ChallengeTS: ts,
		IP:          ip,
		UA:          ua,
		ExpiresAt:   exp,
	})
	return 0, nil
}

// Echo-compatible middleware signature adapter
func (m *Middleware) Echo() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions || isPublicPath(r.URL.Path) || m.Config.Mode == ModeDisabled {
				next.ServeHTTP(w, r)
				return
			}

			token := m.extractToken(r)
			if token == "" && m.Config.Mode == ModeEnforce {
				writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "missing_token"})
				return
			}

			ip := clientIP(r)
			ua := r.Header.Get("User-Agent")

			if cr, ok := m.Cache.Get(token); ok && (!m.Config.BindCacheToIPAndUA || (cr.IP == ip && cr.UA == ua)) {
				next.ServeHTTP(w, r)
				return
			}

			switch m.Config.Mode {
			case ModeEnforce:
				if code, body := m.verifyAndCache(r.Context(), token, ip, ua); code != 0 {
					writeJSON(w, code, body)
					return
				}
			case ModeLogOnly:
				_, _ = m.verifyAndCache(r.Context(), token, ip, ua)
			case ModeDisabled:
				// already handled above
			}

			next.ServeHTTP(w, r)
		})
	}
}
