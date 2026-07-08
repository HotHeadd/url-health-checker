// Package checker defines Checker type and it's implementation.
package checker

import (
	"context"
	"net/http"
	"time"
	"url-checker/internal/storage"

	"golang.org/x/sync/errgroup"
)

const requestTimeout time.Duration = 5 * time.Second

type Checker struct {
	client     *http.Client
	numWorkers int
}

func NewChecker() *Checker {
	return &Checker{
		client:     &http.Client{Timeout: requestTimeout},
		numWorkers: 20,
	}
}

func (c *Checker) CheckOne(ctx context.Context, url string) storage.Result {
	result := storage.Result{URL: url}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		result.Err = err.Error()
		return result
	}
	start := time.Now()
	resp, err := c.client.Do(req)
	result.Duration = time.Since(start).Milliseconds()

	if err != nil {
		result.Err = err.Error()
		return result
	}
	defer resp.Body.Close() //nolint:errcheck // close error intentionally ignored
	result.StatusCode = resp.StatusCode
	return result
}

func (c *Checker) CheckAll(ctx context.Context, urls []string) []storage.Result {
	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(c.numWorkers)

	result := make([]storage.Result, len(urls))
	for i, url := range urls {
		eg.Go(func() error {
			result[i] = c.CheckOne(ctx, url)
			return nil
		})
	}
	_ = eg.Wait()

	return result
}
