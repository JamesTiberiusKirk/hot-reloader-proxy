package proxy

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	_ "embed"

	"github.com/JamesTiberiusKirk/hot-reloader-proxy/cmd/hrp/logger"
	"github.com/JamesTiberiusKirk/hot-reloader-proxy/sse"
)

//go:embed script.js
var script string

const scriptTag = `<script src="/_hot_reloader_proxy/script.js"></script>`

type Handler struct {
	log    logger.Logger
	URL    string
	Target *url.URL
	p      *httputil.ReverseProxy
	sse    *sse.Handler
}

func New(log logger.Logger, port int, target *url.URL) *Handler {
	p := httputil.NewSingleHostReverseProxy(target)
	p.Transport = &roundTripper{
		maxRetries:      10,
		initialDelay:    100 * time.Millisecond,
		backoffExponent: 1.5,
	}
	p.ModifyResponse = func(r *http.Response) error {
		contentType := r.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "text/html") {
			return nil
		}

		var reader io.ReadCloser
		encoding := r.Header.Get("Content-Encoding")
		switch encoding {
		case "gzip":
			plainr, err := gzip.NewReader(r.Body)
			if err != nil {
				log.Info("Error decoding body error:", err.Error())
				return err
			}
			defer plainr.Close()
			reader = plainr
		default:
			reader = r.Body
		}

		body, err := io.ReadAll(reader)
		if err != nil {
			log.Info("Error reading body error: ", err.Error())
			return err
		}

		bodyString := string(body)
		updated := strings.Replace(bodyString, "</body>", scriptTag+"</body>", -1)

		switch encoding {
		case "gzip":
			var buf bytes.Buffer
			gzw := gzip.NewWriter(&buf)
			defer gzw.Close()

			_, err = gzw.Write([]byte(updated))
			if err != nil {
				return err
			}
			err = gzw.Close()
			if err != nil {
				return err
			}

			r.Body = io.NopCloser(&buf)
			r.ContentLength = int64(buf.Len())
			r.Header.Set("Content-Length", strconv.Itoa(buf.Len()))
		default:
			r.Body = io.NopCloser(strings.NewReader(updated))
			r.ContentLength = int64(len(updated))
			r.Header.Set("Content-Length", strconv.Itoa(len(updated)))
		}

		return nil
	}

	return &Handler{
		log:    log,
		URL:    fmt.Sprintf("http://127.0.0.1:%d", port),
		Target: target,
		p:      p,
		sse:    sse.New(),
	}
}

func (p *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/_hot_reloader_proxy/script.js" {
		// Provides a script that reloads the page.
		w.Header().Add("Content-Type", "text/javascript")
		_, err := io.WriteString(w, script)
		if err != nil {
			p.log.Info("Failed to write script:", err)
		}
		return
	}
	if r.URL.Path == "/_hot_reloader_proxy/sse" {
		// Provides a list of messages including a reload message.
		p.sse.ServeHTTP(w, r)
		return
	}
	p.p.ServeHTTP(w, r)
}

func (p *Handler) SendSSE(eventType string, data string) {
	p.sse.Send(eventType, data)
}

type roundTripper struct {
	maxRetries      int
	initialDelay    time.Duration
	backoffExponent float64
}

func (rt *roundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	// Read and buffer the body.
	var bodyBytes []byte
	if r.Body != nil && r.Body != http.NoBody {
		var err error
		bodyBytes, err = io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		r.Body.Close()
	}

	// Retry logic.
	var resp *http.Response
	var err error
	for retries := 0; retries < rt.maxRetries; retries++ {
		// Clone the request and set the body.
		req := r.Clone(r.Context())
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		// Execute the request.
		resp, err = http.DefaultTransport.RoundTrip(req)
		if err != nil {
			time.Sleep(rt.initialDelay * time.Duration(math.Pow(rt.backoffExponent, float64(retries))))
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("max retries reached")
}
