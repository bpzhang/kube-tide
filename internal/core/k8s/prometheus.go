package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const defaultPrometheusTimeout = 30 * time.Second

// PrometheusService Prometheus 查询代理服务
type PrometheusService struct {
	clientManager *ClientManager
	httpClient    *http.Client
}

// NewPrometheusService 创建 Prometheus 服务
func NewPrometheusService(clientManager *ClientManager) *PrometheusService {
	return &PrometheusService{
		clientManager: clientManager,
		httpClient:    &http.Client{Timeout: defaultPrometheusTimeout},
	}
}

// QueryRangeParams Prometheus query_range 参数
type QueryRangeParams struct {
	Query string `json:"query" form:"query" binding:"required"`
	Start string `json:"start" form:"start" binding:"required"`
	End   string `json:"end" form:"end" binding:"required"`
	Step  string `json:"step" form:"step" binding:"required"`
}

// QueryRange 代理 Prometheus query_range API
func (s *PrometheusService) QueryRange(ctx context.Context, clusterName string, params QueryRangeParams, timeout time.Duration) (json.RawMessage, error) {
	promURL, err := s.ResolvePrometheusURL(ctx, clusterName)
	if err != nil {
		return nil, err
	}
	if promURL == "" {
		return nil, fmt.Errorf("集群 %s 未找到 Prometheus 端点", clusterName)
	}
	if err := ValidatePrometheusURL(promURL); err != nil {
		return nil, err
	}
	if len(params.Query) > maxPrometheusQueryLen {
		return nil, fmt.Errorf("PromQL 查询长度超过限制 (%d)", maxPrometheusQueryLen)
	}

	endpoint, err := url.Parse(promURL)
	if err != nil {
		return nil, fmt.Errorf("无效的 Prometheus URL: %w", err)
	}
	endpoint.Path = joinURLPath(endpoint.Path, "/api/v1/query_range")
	q := endpoint.Query()
	q.Set("query", params.Query)
	q.Set("start", params.Start)
	q.Set("end", params.End)
	q.Set("step", params.Step)
	endpoint.RawQuery = q.Encode()

	if timeout <= 0 {
		timeout = defaultPrometheusTimeout
	}
	client := &http.Client{Timeout: timeout}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("创建 Prometheus 请求失败: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Prometheus 请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取 Prometheus 响应失败: %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("Prometheus 返回错误 %d: %s", resp.StatusCode, string(body))
	}
	return json.RawMessage(body), nil
}

// QueryInstant 代理 Prometheus query API（即时查询）；未配置 URL 时自动发现集群内 Prometheus
func (s *PrometheusService) QueryInstant(ctx context.Context, clusterName, query string, timeout time.Duration) (json.RawMessage, error) {
	promURL, err := s.ResolvePrometheusURL(ctx, clusterName)
	if err != nil {
		return nil, err
	}
	if promURL == "" {
		return nil, fmt.Errorf("集群 %s 未找到 Prometheus 端点", clusterName)
	}
	if err := ValidatePrometheusURL(promURL); err != nil {
		return nil, err
	}
	if len(query) > maxPrometheusQueryLen {
		return nil, fmt.Errorf("PromQL 查询长度超过限制 (%d)", maxPrometheusQueryLen)
	}

	endpoint, err := url.Parse(promURL)
	if err != nil {
		return nil, fmt.Errorf("无效的 Prometheus URL: %w", err)
	}
	endpoint.Path = joinURLPath(endpoint.Path, "/api/v1/query")
	q := endpoint.Query()
	q.Set("query", query)
	endpoint.RawQuery = q.Encode()

	if timeout <= 0 {
		timeout = defaultPrometheusTimeout
	}
	client := &http.Client{Timeout: timeout}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("创建 Prometheus 请求失败: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Prometheus 请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取 Prometheus 响应失败: %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("Prometheus 返回错误 %d: %s", resp.StatusCode, string(body))
	}
	return json.RawMessage(body), nil
}

// ResolvePrometheusURL 返回手动配置或集群内自动发现的 Prometheus 地址
func (s *PrometheusService) ResolvePrometheusURL(ctx context.Context, clusterName string) (string, error) {
	if u := s.clientManager.GetPrometheusURL(clusterName); u != "" {
		return u, nil
	}
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return "", err
	}
	if u := DiscoverPrometheusURL(ctx, client); u != "" {
		if err := ValidatePrometheusURL(u); err != nil {
			return "", err
		}
		return u, nil
	}
	return "", nil
}

func joinURLPath(base, path string) string {
	if base == "" || base == "/" {
		return path
	}
	if len(base) > 0 && base[len(base)-1] == '/' {
		return base + path[1:]
	}
	return base + path
}
