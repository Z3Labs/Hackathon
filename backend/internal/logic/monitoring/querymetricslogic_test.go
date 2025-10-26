package monitoring

import (
	"testing"
)

func TestTimestampConversion(t *testing.T) {
	// 模拟VictoriaMetrics返回的数据
	tests := []struct {
		name      string
		timestamp interface{}
		expected  int64
	}{
		{
			name:      "regular timestamp",
			timestamp: 1761458481.0, // VictoriaMetrics返回秒级时间戳
			expected:  1761458481,
		},
		{
			name:      "another timestamp",
			timestamp: 1761458541.0,
			expected:  1761458541,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timestamp := int64(tt.timestamp.(float64))
			if timestamp != tt.expected {
				t.Errorf("timestamp conversion failed: got %d, expected %d", timestamp, tt.expected)
			}
			t.Logf("Success: timestamp=%d", timestamp)
		})
	}
}

func TestParseMetricsData(t *testing.T) {
	// 模拟VictoriaMetrics的完整数据结构
	values := [][]interface{}{
		{1761458481.0, "2.083333333333337"},
		{1761458541.0, "2.0783333333335263"},
		{1761458601.0, "2.083333333333326"},
	}

	for i, value := range values {
		timestamp := int64(value[0].(float64))
		val := value[1].(string)

		t.Logf("Data point %d: timestamp=%d, value=%s", i, timestamp, val)

		if timestamp != 1761458481+int64(i*60) {
			t.Errorf("unexpected timestamp at index %d: got %d", i, timestamp)
		}
	}
}

func TestSimulateRealData(t *testing.T) {
	// 模拟真实API返回的数据结构
	type VMResult struct {
		Metric map[string]string `json:"metric"`
		Values [][]interface{}   `json:"values"`
	}

	vmResult := VMResult{
		Metric: map[string]string{
			"instance": "localhost:9301",
		},
		Values: [][]interface{}{
			{1761458481.0, "2.083333333333337"},
			{1761458541.0, "2.0783333333335263"},
			{1761458601.0, "2.083333333333326"},
			{1761458661.0, "2.09500000000602"},
			{1761458721.0, "2.128333333331578"},
		},
	}

	// 模拟实际转换过程
	t.Logf("Simulating real VictoriaMetrics data:")
	for i, value := range vmResult.Values {
		timestamp := int64(value[0].(float64))
		val := value[1].(string)

		t.Logf("  DataPoint[%d]: timestamp=%d, value=%s", i, timestamp, val)

		// 验证时间戳是正确的Unix时间戳
		if timestamp < 1000000000 { // 小于2001年的时间戳不合理
			t.Errorf("timestamp %d looks suspicious", timestamp)
		}
	}

	t.Log("All data points processed successfully")
}
