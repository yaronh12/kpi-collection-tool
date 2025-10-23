package prometheus

import "net/http"

type tokenRoundTripper struct {
	Token string
	RT    http.RoundTripper
}

// tokenRoundTripper adds Bearer token authentication to HTTP requests
func (t *tokenRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.Token)
	return t.RT.RoundTrip(req)
}
