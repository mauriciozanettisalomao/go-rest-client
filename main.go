package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/mauriciozanettisalomao/go-rest-client/client"
)

func main() {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	restClient := client.NewRestClient().
		WithMethod("GET").
		WithURL("http://localhost:8080").
		WithHeader(map[string]string{"Content-Type": "application/json"}).
		WithTimeout(time.Second * 10).
		WithIntervalSeconds(1).
		WithBackoffRate(2).
		WithMaxAttempts(3)

	r := client.NewRequester(restClient)

	response := make(map[string]interface{})

	status, err := r.Do(ctx, nil, &response)
	if err != nil {
		slog.ErrorContext(ctx, "error making request",
			"error", err,
			"status", status,
		)
		return
	}

	slog.InfoContext(ctx, "response",
		"status", status,
		"response", response,
	)

}
