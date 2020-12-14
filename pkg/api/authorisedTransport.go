package api

import (
	"fmt"
	"net/http"
)

type authorizedTransport struct {
	underlyingTransport http.RoundTripper
	bearerToken         string
}

func (t *authorizedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	headerValue := fmt.Sprintf("Bearer %s", t.bearerToken)
	req.Header.Add("Authorization", headerValue)
	return t.underlyingTransport.RoundTrip(req)
}