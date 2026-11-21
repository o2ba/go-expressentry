package expressentry

import (
	"context"
	"fmt"
	"io"
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
// does. A response other than 200 OK for the English file is returned as an
// [*HTTPError]. If only the French file fails, the rounds come back with
// English text and a [FrenchUnavailable] issue.
func (c *Client) Rounds(ctx context.Context) (*Rounds, error) {
	r, _, err := c.RoundsSince(ctx, Version{})
	return r, err
}

// RoundsSince is Rounds, but asks the server to send the files only if
// they changed after v, which came from an earlier [Rounds.Version]. When
// neither file changed it returns nil, false and a nil error. When one
// changed, both are read again and changed is true. A zero v behaves like
// Rounds.
func (c *Client) RoundsSince(ctx context.Context, v Version) (r *Rounds, changed bool, err error) {
	en, err := c.get(ctx, c.englishURL, v.EN)
	if err != nil {
		return nil, false, err
	}
	var fr fetched
	var frErr error
	if c.frenchURL != "" {
		fr, frErr = c.get(ctx, c.frenchURL, v.FR)
	}
	frUnchanged := c.frenchURL == "" || frErr != nil || fr.notModified
	if en.notModified && frUnchanged {
		return nil, false, nil
	}
	if en.notModified {
		if en, err = c.get(ctx, c.englishURL, time.Time{}); err != nil {
			return nil, false, err
		}
	}
	if fr.notModified {
		fr, frErr = c.get(ctx, c.frenchURL, time.Time{})
	}

	if frErr != nil {
		fr = fetched{}
	}
	r, err = Parse(en.body, fr.body)
	if err != nil {
		return nil, false, err
	}
	if frErr != nil {
		r.Issues = append([]Issue{{Kind: FrenchUnavailable, Err: frErr}}, r.Issues...)
	}
	r.Version = Version{EN: en.lastModified, FR: fr.lastModified}
	return r, true, nil
}

const maxFileSize = 16 << 20

type fetched struct {
	body         []byte
	lastModified time.Time
	notModified  bool
}

func (c *Client) get(ctx context.Context, url string, since time.Time) (fetched, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fetched{}, fmt.Errorf("expressentry: %w", err)
	}
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
	if !since.IsZero() {
		req.Header.Set("If-Modified-Since", since.UTC().Format(http.TimeFormat))
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fetched{}, fmt.Errorf("expressentry: %w", err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusNotModified && !since.IsZero():
		return fetched{lastModified: since, notModified: true}, nil
	case resp.StatusCode != http.StatusOK:
		return fetched{}, &HTTPError{URL: url, StatusCode: resp.StatusCode}
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxFileSize+1))
	if err != nil {
		return fetched{}, fmt.Errorf("expressentry: GET %s: %w", url, err)
	}
	if len(body) > maxFileSize {
		return fetched{}, fmt.Errorf("%w: GET %s: larger than %d bytes", ErrUnexpectedFormat, url, maxFileSize)
	}
	f := fetched{body: body}
	if t, err := http.ParseTime(resp.Header.Get("Last-Modified")); err == nil {
		f.lastModified = t
	}
	return f, nil
}
