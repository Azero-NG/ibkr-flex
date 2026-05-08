package flex

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	defaultBaseURL  = "https://ndcdyn.interactivebrokers.com/AccountManagement/FlexWebService"
	apiVersion      = "3"
	pollInterval    = 3 * time.Second
	pollMaxDuration = 60 * time.Second
	userAgent       = "ibkr-flex/0.1 (+go)"
)

type Client struct {
	HTTP    *http.Client
	BaseURL string
	OnRetry func(elapsed time.Duration)
}

func NewClient() *Client {
	return &Client{
		HTTP:    &http.Client{Timeout: 30 * time.Second},
		BaseURL: defaultBaseURL,
	}
}

func (c *Client) Fetch(ctx context.Context, token, queryID string) ([]byte, error) {
	ref, err := c.sendRequest(ctx, token, queryID)
	if err != nil {
		return nil, err
	}
	return c.pollStatement(ctx, token, ref)
}

func (c *Client) sendRequest(ctx context.Context, token, queryID string) (string, error) {
	u := fmt.Sprintf("%s/SendRequest?t=%s&q=%s&v=%s",
		c.BaseURL,
		url.QueryEscape(token),
		url.QueryEscape(queryID),
		apiVersion,
	)
	body, err := c.get(ctx, u)
	if err != nil {
		return "", err
	}
	resp, err := parseStatementResponse(body)
	if err != nil {
		return "", err
	}
	if resp.Status != "Success" {
		return "", errorFromCode(resp.ErrorCode, resp.ErrorMessage)
	}
	if resp.ReferenceCode == "" {
		return "", fmt.Errorf("flex: empty ReferenceCode despite Status=Success")
	}
	return resp.ReferenceCode, nil
}

func (c *Client) pollStatement(ctx context.Context, token, refCode string) ([]byte, error) {
	u := fmt.Sprintf("%s/GetStatement?t=%s&q=%s&v=%s",
		c.BaseURL,
		url.QueryEscape(token),
		url.QueryEscape(refCode),
		apiVersion,
	)
	start := time.Now()
	for {
		body, err := c.get(ctx, u)
		if err != nil {
			return nil, err
		}
		if !isStatementResponse(body) {
			return body, nil
		}
		resp, err := parseStatementResponse(body)
		if err != nil {
			return nil, err
		}
		mapped := errorFromCode(resp.ErrorCode, resp.ErrorMessage)
		if !errors.Is(mapped, ErrStatementInProgress) {
			return nil, mapped
		}
		if time.Since(start) >= pollMaxDuration {
			return nil, ErrStatementTimeout
		}
		if c.OnRetry != nil {
			c.OnRetry(time.Since(start))
		}
		if err := sleepCtx(ctx, pollInterval); err != nil {
			return nil, err
		}
	}
}

func (c *Client) get(ctx context.Context, u string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("flex: http request: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("flex: read response body: %w", err)
	}
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("flex: HTTP %d: %s", resp.StatusCode, bytes.TrimSpace(body))
	}
	return body, nil
}

func isStatementResponse(body []byte) bool {
	dec := xml.NewDecoder(bytes.NewReader(body))
	for {
		tok, err := dec.Token()
		if err != nil {
			return false
		}
		if start, ok := tok.(xml.StartElement); ok {
			return start.Name.Local == "FlexStatementResponse"
		}
	}
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
