package httpclient

import (
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
)

func New(cfg Config) *http.Client {
	cfg = applyConfig(cfg)

	retryClient := retryablehttp.NewClient()

	retryClient.RetryMax = cfg.RetryMax
	retryClient.RetryWaitMin = cfg.RetryWaitMin
	retryClient.RetryWaitMax = cfg.RetryWaitMax

	retryClient.Logger = nil
	retryClient.CheckRetry = retryablehttp.DefaultRetryPolicy
	retryClient.Backoff = retryablehttp.DefaultBackoff

	httpClient := retryClient.StandardClient()
	httpClient.Timeout = cfg.Timeout

	return httpClient
}
