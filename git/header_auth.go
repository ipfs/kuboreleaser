package git

import (
	"encoding/base64"
	"fmt"
	"net/http"
)

// https://github.com/go-git/go-git/issues/474
type HeaderAuth struct {
	Key   string
	Value string
}

func NewHeaderAuth(token string) (*HeaderAuth, error) {
	auth := base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("pat:%s", token)))

	return &HeaderAuth{
		Key:   "Authorization",
		Value: fmt.Sprintf("Basic %s", auth),
	}, nil
}

func (h HeaderAuth) String() string {
	return fmt.Sprintf("%s: %s", h.Key, h.Value)
}

func (h HeaderAuth) Name() string {
	return "extraheader"
}

func (h HeaderAuth) SetAuth(r *http.Request) {
	r.Header.Set(h.Key, h.Value)
}
