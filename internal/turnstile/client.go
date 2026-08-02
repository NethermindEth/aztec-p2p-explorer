package turnstile

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"
)

const siteverifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

type verifyResponse struct {
	Success     bool     `json:"success"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes"`
	Action      string   `json:"action"`
	CData       string   `json:"cdata"`
}

type Verifier interface {
	Verify(ctx context.Context, token, remoteIP string) (verifyResponse, error)
}

type httpVerifier struct {
	secret string
	client *http.Client
}

func NewHTTPVerifier(secret string, timeout time.Duration) Verifier {
	return &httpVerifier{
		secret: secret,
		client: &http.Client{Timeout: timeout},
	}
}

func (v *httpVerifier) Verify(ctx context.Context, token, remoteIP string) (verifyResponse, error) {
	if token == "" || len(token) > 2048 {
		return verifyResponse{}, errors.New("invalid token")
	}
	form := url.Values{}
	form.Set("secret", v.secret)
	form.Set("response", token)
	if remoteIP != "" {
		form.Set("remoteip", remoteIP)
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, siteverifyURL, bytes.NewBufferString(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := v.client.Do(req)
	if err != nil {
		return verifyResponse{}, err
	}
	defer resp.Body.Close()

	var vr verifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&vr); err != nil {
		return verifyResponse{}, err
	}
	return vr, nil
}
