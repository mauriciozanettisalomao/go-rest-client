package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"math"
	"net/http"
	"time"
)

const (
	internalStatusRequestError = 999
)

// RestClient is a client that can make HTTP requests.
type RestClient struct {
	method          string
	url             string
	header          map[string]string
	maxAttempts     int64
	intervalSeconds float64
	backoffRate     float64
	timeout         time.Duration
}

// WithMethod sets the HTTP method for the request.
func (r *RestClient) WithMethod(method string) *RestClient {
	r.method = method
	return r
}

// WithURL sets the URL for the request.
func (c *RestClient) WithURL(url string) *RestClient {
	c.url = url
	return c
}

// WithHeader sets the headers for the request.
func (r *RestClient) WithHeader(header map[string]string) *RestClient {
	r.header = header
	return r
}

// WithIntervalSeconds sets the interval between retries.
func (r *RestClient) WithIntervalSeconds(intervalSeconds float64) *RestClient {
	r.intervalSeconds = intervalSeconds
	return r
}

// WithBackoffRate sets the backoff rate for retries.
func (r *RestClient) WithBackoffRate(backoffRate float64) *RestClient {
	r.backoffRate = backoffRate
	return r
}

// WithMaxAttempts sets the maximum number of attempts.
func (r *RestClient) WithMaxAttempts(maxAttempts int64) *RestClient {
	r.maxAttempts = maxAttempts
	return r
}

// WithTimeout sets the timeout for the request.
func (r *RestClient) WithTimeout(timeout time.Duration) *RestClient {
	r.timeout = timeout
	return r
}

// Do makes an HTTP request
func (r *RestClient) Do(ctx context.Context, request interface{}, response interface{}) (int64, error) {

	var (
		retries int64
		status  int64
		err     error
		resp    []byte
	)

	client := &http.Client{}
	if r.timeout > 0 {
		client.Timeout = r.timeout
	}

	sleep := float64(0)
	for i := int64(0); i < r.maxAttempts; i++ {

		time.Sleep(time.Second * time.Duration(sleep))

		status, resp, err = r.call(ctx, *client, request)

		// if it is handled error, there is no need to retry
		if status < http.StatusInternalServerError {
			break
		}
		retries++

		slog.WarnContext(ctx, "retrying request",
			"error", err,
			"url", r.url,
			"status", status,
			"backoff", sleep,
			"interval", r.intervalSeconds,
			"attempt", retries,
			"time", time.Now().Format(time.RFC3339),
		)

		sleep = r.intervalSeconds * (math.Pow(r.backoffRate, float64(i+1)))

	}

	if err != nil {
		slog.ErrorContext(ctx, "error calling api",
			"err", err,
			"url", r.url,
		)
		return internalStatusRequestError, err
	}

	if err = json.Unmarshal(resp, &response); err != nil {
		slog.ErrorContext(ctx, "failed to Unmarshal data",
			"err", err,
			"url", r.url,
		)
		return internalStatusRequestError, err
	}

	slog.DebugContext(ctx, "request done",
		"url", r.url,
		"retries", retries,
	)

	return status, err
}

func (r *RestClient) call(ctx context.Context, client http.Client, request interface{}) (int64, []byte, error) {

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(request)
	if err != nil {
		slog.ErrorContext(ctx, "error encoding request",
			"err", err,
		)
		return internalStatusRequestError, nil, err
	}

	req, err := http.NewRequest(r.method, r.url, &buf)
	if err != nil {
		slog.ErrorContext(ctx, "error creating request",
			"err", err,
		)
		return internalStatusRequestError, nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "error making request",
			"err", err,
		)
		return internalStatusRequestError, nil, err
	}

	defer resp.Body.Close()
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.ErrorContext(ctx, "error reading response",
			"err", err,
		)
		return internalStatusRequestError, nil, err
	}

	return int64(resp.StatusCode), bytes, nil
}

// NewRestClient creates a new Rest Client
func NewRestClient() *RestClient {
	return &RestClient{}
}
