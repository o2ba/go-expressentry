package expressentry

import (
	"context"
	"net/http"
	"time"
)

// URLs of the rounds files IRCC publishes, used unless [WithURLs] is given.
const (
	DefaultEnglishURL = "https://www.canada.ca/content/dam/ircc/documents/json/ee_rounds_123_en.json"
	DefaultFrenchURL  = "https://www.canada.ca/content/dam/ircc/documents/json/ee_rounds_123_fr.json"
)

const defaultTimeout = 30 * time.Second

// Client downloads and reads IRCC's rounds files. The zero value is not
// usable; create one with [NewClient]. A Client is safe for concurrent use.
type Client struct {
	httpClient *http.Client
	englishURL string
	frenchURL  string
	userAgent  string
}

// Option configures a [Client].
type Option func(*Client)

// NewClient returns a Client that reads IRCC's published files with a
// 30-second timeout, unless options say otherwise.
func NewClient(opts ...Option) *Client {
	c := &Client{
		httpClient: &http.Client{Timeout: defaultTimeout},
		englishURL: DefaultEnglishURL,
		frenchURL:  DefaultFrenchURL,
		userAgent:  "go-expressentry",
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// WithHTTPClient makes the Client send requests with hc. A nil hc is
// ignored.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		if hc != nil {
			c.httpClient = hc
		}
	}
}

// WithURLs reads the rounds files from en and fr instead of IRCC's, for
// tests, mirrors or proxies. An empty fr skips the French file, and every
// [Name] then has only its English text.
func WithURLs(en, fr string) Option {
	return func(c *Client) {
		c.englishURL = en
		c.frenchURL = fr
	}
}

// WithUserAgent sets the User-Agent header sent with every request.
func WithUserAgent(ua string) Option {
	return func(c *Client) {
		c.userAgent = ua
	}
}

// Rounds downloads the English and French files and reads them as [Parse]
// does. A response other than 200 OK is returned as an [*HTTPError].
func (c *Client) Rounds(ctx context.Context) (*Rounds, error) {
	return nil, ErrNotImplemented
}

// RoundsSince is Rounds, but asks the server to send the files only if
// they changed after v, which came from an earlier [Rounds.Version]. When
// neither file changed it returns nil, false and a nil error. When one
// changed, both are read again and changed is true. A zero v behaves like
// Rounds.
func (c *Client) RoundsSince(ctx context.Context, v Version) (r *Rounds, changed bool, err error) {
	return nil, false, ErrNotImplemented
}
