package plugin

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/evcc-io/evcc/plugin/pipeline"
	"github.com/evcc-io/evcc/util"
	"github.com/evcc-io/evcc/util/request"
	"github.com/evcc-io/evcc/util/transport"
	"github.com/gregjones/httpcache"
)

// HTTP implements HTTP request provider
type HTTP struct {
	*getter
	*request.Helper
	url, method string
	headers     map[string]string
	body        string
	pipeline    *pipeline.Pipeline
}

func init() {
	registry.AddCtx("http", NewHTTPPluginFromConfig)
}

var mc = httpcache.NewMemoryCache()

// NewHTTPPluginFromConfig creates a HTTP provider
func NewHTTPPluginFromConfig(ctx context.Context, other map[string]interface{}) (Plugin, error) {
	cc := struct {
		URI, Method       string
		Headers           map[string]string
		Body              string
		pipeline.Settings `mapstructure:",squash"`
		Scale             float64
		Insecure          bool
		Auth              Auth
		Timeout           time.Duration
		Cache             time.Duration
	}{
		Headers: make(map[string]string),
		Method:  http.MethodGet,
		Scale:   1,
		Timeout: request.Timeout,
	}

	if err := util.DecodeOther(other, &cc); err != nil {
		return nil, err
	}

	if cc.URI == "" {
		return nil, errors.New("missing uri")
	}

	log := contextLogger(ctx, util.NewLogger("http"))
	p := NewHTTP(
		log,
		strings.ToUpper(cc.Method),
		cc.URI,
		cc.Insecure,
		cc.Cache,
	).
		WithHeaders(cc.Headers).
		WithBody(cc.Body)

	p.Client.Timeout = cc.Timeout

	p.getter = defaultGetters(p, cc.Scale)

	if cc.Auth.Type != "" || cc.Auth.Source != "" {
		transport, err := cc.Auth.Transport(ctx, log, p.Client.Transport)
		if err != nil {
			return nil, err
		}
		p.Client.Transport = transport
	}

	pipe, err := pipeline.New(log, cc.Settings)
	if err != nil {
		return nil, err
	}
	p.pipeline = pipe

	return p, nil
}

// NewHTTP create HTTP provider
func NewHTTP(log *util.Logger, method, uri string, insecure bool, cache time.Duration) *HTTP {
	p := &HTTP{
		Helper: request.NewHelper(log),
		url:    uri,
		method: method,
	}

	// http cache
	p.Client.Transport = &httpcache.Transport{
		Cache:     mc,
		Transport: p.Client.Transport,
	}

	if cache > 0 {
		cacheHeader := fmt.Sprintf("max-age=%d, must-revalidate", int(cache.Seconds()))
		p.Client.Transport = &transport.Decorator{
			Decorator: transport.DecorateHeaders(map[string]string{
				"Cache-Control": cacheHeader,
			}),
			Base: p.Client.Transport,
		}
	}

	// ignore the self signed certificate
	if insecure {
		p.Client.Transport = request.NewTripper(log, transport.Insecure())
	}

	return p
}

// WithBody adds request body
func (p *HTTP) WithBody(body string) *HTTP {
	if body != "" {
		p.body = body
		if p.method == http.MethodGet {
			p.method = http.MethodPost
		}
	}
	return p
}

// WithHeaders adds request headers
func (p *HTTP) WithHeaders(headers map[string]string) *HTTP {
	p.headers = headers
	return p
}

// request executes the configured request or returns the cached value
func (p *HTTP) request(url string, body string) ([]byte, error) {
	var b io.Reader
	if p.method != http.MethodGet {
		b = strings.NewReader(body)
	}

	url = util.DefaultScheme(url, "http")

	// empty method becomes GET
	req, err := request.New(p.method, url, b, p.headers)
	if err != nil {
		return []byte{}, err
	}

	val, err := p.DoBody(req)
	if err != nil {
		if err2 := knownErrors(val); err2 != nil {
			err = err2
		}
	}

	return val, err
}

var _ Getters = (*HTTP)(nil)

// StringGetter sends string request
func (p *HTTP) StringGetter() (func() (string, error), error) {
	return func() (string, error) {
		url, err := setFormattedValue(p.url, "", "")
		if err != nil {
			return "", err
		}

		b, err := p.request(url, p.body)

		if err == nil && p.pipeline != nil {
			b, err = p.pipeline.Process(b)
		}

		return string(b), err
	}, nil
}

func (p *HTTP) set(param string, val interface{}) error {
	url, err := setFormattedValue(p.url, param, val)
	if err != nil {
		return err
	}

	body, err := setFormattedValue(p.body, param, val)
	if err != nil {
		return err
	}

	_, err = p.request(url, body)

	return err
}

var _ IntSetter = (*HTTP)(nil)

// IntSetter sends int request
func (p *HTTP) IntSetter(param string) (func(int64) error, error) {
	return func(val int64) error {
		return p.set(param, val)
	}, nil
}

var _ FloatSetter = (*HTTP)(nil)

// FloatSetter sends int request
func (p *HTTP) FloatSetter(param string) (func(float64) error, error) {
	return func(val float64) error {
		return p.set(param, val)
	}, nil
}

var _ StringSetter = (*HTTP)(nil)

// StringSetter sends string request
func (p *HTTP) StringSetter(param string) (func(string) error, error) {
	return func(val string) error {
		return p.set(param, val)
	}, nil
}

var _ BoolSetter = (*HTTP)(nil)

// BoolSetter sends bool request
func (p *HTTP) BoolSetter(param string) (func(bool) error, error) {
	return func(val bool) error {
		return p.set(param, val)
	}, nil
}
