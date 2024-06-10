package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDo(t *testing.T) {

	tests := []struct {
		name            string
		method          string
		headers         map[string]string
		maxAttempts     int64
		intervalSeconds float64
		backoffRate     float64
		mockResponse    string
		statusCode      int
		expected        interface{}
		expectedStatus  int64
		expectedError   error
	}{
		{
			name:            "success",
			method:          "GET",
			headers:         map[string]string{"Content-Type": "application/json"},
			maxAttempts:     3,
			intervalSeconds: 1,
			backoffRate:     2,
			mockResponse:    `{"message": "success"}`,
			statusCode:      200,
			expected:        map[string]interface{}{"message": "success"},
			expectedStatus:  200,
			expectedError:   nil,
		},
	}

	assertion := assert.New(t)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, "%v", tc.mockResponse)
			}))
			defer svr.Close()

			m := &RestClient{}
			m.WithURL(svr.URL)
			m.WithMethod(tc.method)
			m.WithHeader(tc.headers)
			m.WithMaxAttempts(tc.maxAttempts)
			m.WithIntervalSeconds(tc.intervalSeconds)
			m.WithBackoffRate(tc.backoffRate)

			var result map[string]interface{}
			status, err := m.Do(context.Background(), nil, &result)
			if err != nil {
				assertion.Contains(err.Error(), tc.expectedError.Error())
				return
			}
			assertion.Equal(tc.expected, result)
			assertion.Equal(tc.expectedStatus, status)
		})
	}
}

func TestDoFailed(t *testing.T) {

	tests := []struct {
		name            string
		method          string
		headers         map[string]string
		maxAttempts     int64
		intervalSeconds float64
		backoffRate     float64
		timeout         time.Duration
		endpointSleep   time.Duration
		mockResponse    string
		statusCode      int
		expected        interface{}
		expectedStatus  int64
		expectedError   error
	}{
		{
			name:            "success",
			method:          "GET",
			headers:         map[string]string{"Content-Type": "application/json"},
			maxAttempts:     3,
			intervalSeconds: 1,
			backoffRate:     2,
			endpointSleep:   0,
			timeout:         time.Second * 2,
			mockResponse:    `{"message": "success"}`,
			statusCode:      200,
			expected:        map[string]interface{}{"message": "success"},
			expectedStatus:  200,
			expectedError:   nil,
		},
		{
			name:            "timeout",
			method:          "GET",
			headers:         map[string]string{"Content-Type": "application/json"},
			maxAttempts:     1,
			intervalSeconds: 0,
			backoffRate:     0,
			endpointSleep:   time.Millisecond * 100,
			timeout:         time.Millisecond * 1,
			mockResponse:    `{"message": "success"}`,
			statusCode:      999,
			expected:        nil,
			expectedStatus:  0,
			expectedError:   fmt.Errorf("context deadline exceeded"),
		},
	}

	assertion := assert.New(t)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(tc.endpointSleep * time.Second)
				fmt.Fprintf(w, "%v", tc.mockResponse)
			}))
			defer svr.Close()

			m := &RestClient{}
			m.WithURL(svr.URL)
			m.WithMethod(tc.method)
			m.WithHeader(tc.headers)
			m.WithMaxAttempts(tc.maxAttempts)
			m.WithIntervalSeconds(tc.intervalSeconds)
			m.WithBackoffRate(tc.backoffRate)
			m.WithTimeout(tc.timeout)

			var result map[string]interface{}
			status, err := m.Do(context.Background(), nil, &result)
			if err != nil {
				assertion.Contains(err.Error(), tc.expectedError.Error())
				return
			}
			assertion.Equal(tc.expected, result)
			assertion.Equal(tc.expectedStatus, status)
		})
	}
}
