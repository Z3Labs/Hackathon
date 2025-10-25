package prom

type Sample struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

type RangeQueryResult struct {
	Metric map[string]string `json:"metric"`
	Values []Sample          `json:"values"`
}

type InstantQueryResult struct {
	Metric map[string]string `json:"metric"`
	Value  Sample            `json:"value"`
}

type AlertRule struct {
	Alert       string            `yaml:"alert"`
	Expr        string            `yaml:"expr"`
	For         string            `yaml:"for,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

type vmResponse struct {
	Status string      `json:"status"`
	Data   vmData      `json:"data"`
	Error  string      `json:"error,omitempty"`
}

type vmData struct {
	ResultType string        `json:"resultType"`
	Result     []vmResult    `json:"result"`
}

type vmResult struct {
	Metric map[string]string `json:"metric"`
	Value  []interface{}     `json:"value,omitempty"`
	Values [][]interface{}   `json:"values,omitempty"`
}
