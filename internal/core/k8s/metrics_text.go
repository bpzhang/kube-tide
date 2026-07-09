package k8s

import (
	"bufio"
	"bytes"
	"sort"
	"strconv"
	"strings"
)

type promSample struct {
	Name   string
	Labels map[string]string
	Value  float64
}

func parsePrometheusText(body []byte) []promSample {
	samples := make([]promSample, 0)
	scanner := bufio.NewScanner(bytes.NewReader(body))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if sample, ok := parsePrometheusLine(line); ok {
			samples = append(samples, sample)
		}
	}
	return samples
}

func parsePrometheusLine(line string) (promSample, bool) {
	space := strings.LastIndex(line, " ")
	if space <= 0 {
		return promSample{}, false
	}
	val, err := strconv.ParseFloat(strings.TrimSpace(line[space+1:]), 64)
	if err != nil {
		return promSample{}, false
	}
	metricPart := strings.TrimSpace(line[:space])
	name := metricPart
	labels := map[string]string{}
	if i := strings.Index(metricPart, "{"); i >= 0 && strings.HasSuffix(metricPart, "}") {
		name = metricPart[:i]
		raw := metricPart[i+1 : len(metricPart)-1]
		for _, pair := range splitPromLabels(raw) {
			k, v, ok := strings.Cut(pair, "=")
			if !ok {
				continue
			}
			labels[k] = strings.Trim(v, `"`)
		}
	}
	return promSample{Name: name, Labels: labels, Value: val}, true
}

func splitPromLabels(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := make([]string, 0)
	var b strings.Builder
	inQuotes := false
	for i := 0; i < len(raw); i++ {
		ch := raw[i]
		if ch == '"' {
			inQuotes = !inQuotes
			b.WriteByte(ch)
			continue
		}
		if ch == ',' && !inQuotes {
			parts = append(parts, b.String())
			b.Reset()
			continue
		}
		b.WriteByte(ch)
	}
	if b.Len() > 0 {
		parts = append(parts, b.String())
	}
	return parts
}

func topHubbleDrops(samples []promSample, limit int) []HubbleDropStat {
	stats := make([]HubbleDropStat, 0)
	for _, s := range samples {
		if s.Name != "hubble_drop_total" || s.Value <= 0 {
			continue
		}
		reason := s.Labels["reason"]
		if reason == "" {
			reason = "unknown"
		}
		stats = append(stats, HubbleDropStat{Reason: reason, Count: s.Value})
	}
	return topNByCountDrop(stats, limit)
}

func topHubblePorts(samples []promSample, limit int) []HubblePortStat {
	stats := make([]HubblePortStat, 0)
	for _, s := range samples {
		if s.Name != "hubble_port_distribution_total" || s.Value <= 0 {
			continue
		}
		stats = append(stats, HubblePortStat{
			Protocol: s.Labels["protocol"],
			Port:     s.Labels["port"],
			Count:    s.Value,
		})
	}
	return topNByCountPort(stats, limit)
}

func topNByCountDrop(items []HubbleDropStat, n int) []HubbleDropStat {
	return sortTakeDrop(items, n)
}

func sortTakeDrop(items []HubbleDropStat, n int) []HubbleDropStat {
	sort.Slice(items, func(i, j int) bool { return items[i].Count > items[j].Count })
	if len(items) > n {
		return items[:n]
	}
	return items
}

func sortTakePort(items []HubblePortStat, n int) []HubblePortStat {
	sort.Slice(items, func(i, j int) bool { return items[i].Count > items[j].Count })
	if len(items) > n {
		return items[:n]
	}
	return items
}

func topNByCountPort(items []HubblePortStat, n int) []HubblePortStat {
	return sortTakePort(items, n)
}
