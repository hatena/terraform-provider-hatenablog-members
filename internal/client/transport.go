package client

import (
	"fmt"
	"net/http"

	"github.com/motemen/go-wsse"
)

// transport wraps wsse.Transport to set X-WSSE header.
// Additionally, it sets User-Agent header
type transport struct {
	Transport wsse.Transport

	version string
}

func newTransport(username, apikey, version string) *transport {
	return &transport{
		Transport: wsse.Transport{
			Username: username,
			Password: apikey,
		},
		version: version,
	}
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	ua := fmt.Sprintf("terraform-provider-hatenablog-members/%s (+https://github.com/hatena/terraform-provider-hatenablog-members)", t.version)
	req.Header.Set("User-Agent", ua)

	return t.Transport.RoundTrip(req)
}
