package prom

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type VMClient interface {
	QueryInstant(query string) ([]InstantQueryResult, error)
	QueryRange(query string, start, end time.Time, step time.Duration) ([]RangeQueryResult, error)
	AddAlert(rule AlertRule) error
	DeleteAlert(alertName string) error
	UpdateAlert(rule AlertRule) error
}

type vmClient struct {
	config *VMClientConfig
}

func NewVMClient(config *VMClientConfig) VMClient {
	return &vmClient{
		config: config,
	}
}

func (c *vmClient) QueryInstant(query string) ([]InstantQueryResult, error) {
	apiURL := fmt.Sprintf("%s/api/v1/query?query=%s", c.config.BaseURL, url.QueryEscape(query))
	
	resp, err := c.config.HTTPClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to query instant: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("query failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result vmResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("query failed: %s", result.Error)
	}

	return c.parseInstantResults(result.Data.Result), nil
}

func (c *vmClient) QueryRange(query string, start, end time.Time, step time.Duration) ([]RangeQueryResult, error) {
	params := url.Values{}
	params.Add("query", query)
	params.Add("start", strconv.FormatInt(start.Unix(), 10))
	params.Add("end", strconv.FormatInt(end.Unix(), 10))
	params.Add("step", fmt.Sprintf("%ds", int(step.Seconds())))

	apiURL := fmt.Sprintf("%s/api/v1/query_range?%s", c.config.BaseURL, params.Encode())

	resp, err := c.config.HTTPClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to query range: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("query failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result vmResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("query failed: %s", result.Error)
	}

	return c.parseRangeResults(result.Data.Result), nil
}

func (c *vmClient) parseInstantResults(results []vmResult) []InstantQueryResult {
	instantResults := make([]InstantQueryResult, 0, len(results))
	
	for _, r := range results {
		if len(r.Value) >= 2 {
			timestamp, _ := r.Value[0].(float64)
			value, _ := strconv.ParseFloat(r.Value[1].(string), 64)
			
			instantResults = append(instantResults, InstantQueryResult{
				Metric: r.Metric,
				Value: Sample{
					Timestamp: int64(timestamp),
					Value:     value,
				},
			})
		}
	}
	
	return instantResults
}

func (c *vmClient) parseRangeResults(results []vmResult) []RangeQueryResult {
	rangeResults := make([]RangeQueryResult, 0, len(results))
	
	for _, r := range results {
		samples := make([]Sample, 0, len(r.Values))
		
		for _, v := range r.Values {
			if len(v) >= 2 {
				timestamp, _ := v[0].(float64)
				value, _ := strconv.ParseFloat(v[1].(string), 64)
				
				samples = append(samples, Sample{
					Timestamp: int64(timestamp),
					Value:     value,
				})
			}
		}
		
		rangeResults = append(rangeResults, RangeQueryResult{
			Metric: r.Metric,
			Values: samples,
		})
	}
	
	return rangeResults
}
