package monitoring

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryMetricsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryMetricsLogic(ctx context.Context, svcCtx *svc.ServiceContext) QueryMetricsLogic {
	return QueryMetricsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryMetricsLogic) QueryMetrics(req *types.QueryMetricsReq) (resp *types.QueryMetricsResp, err error) {
	// 验证参数
	if req.Query == "" {
		return nil, errors.New("查询语句不能为空")
	}

	if req.Start == "" || req.End == "" {
		return nil, errors.New("开始时间和结束时间不能为空")
	}

	startTime, err := strconv.ParseInt(req.Start, 10, 64)
	if err != nil {
		return nil, errors.New("开始时间格式错误")
	}

	endTime, err := strconv.ParseInt(req.End, 10, 64)
	if err != nil {
		return nil, errors.New("结束时间格式错误")
	}

	l.Infof("[QueryMetrics] Query=%s, Start=%d, End=%d, Step=%s", req.Query, startTime, endTime, req.Step)

	// 构建VictoriaMetrics API URL
	vmURL := l.svcCtx.Config.VM.VMUIURL
	if vmURL == "" {
		return nil, errors.New("VictoriaMetrics URL未配置")
	}

	// 构建完整的 API URL
	// vmURL 是基础URL，如 http://150.158.152.112:9300
	// 需要转换为 http://150.158.152.112:9300/api/v1/query_range
	vmAPIURL := vmURL + "/api/v1/query_range"

	// 调用VictoriaMetrics API
	series, err := l.queryVMAPI(vmAPIURL, req.Query, startTime, endTime, req.Step)
	if err != nil {
		l.Errorf("[QueryMetrics] queryVMAPI error:%v", err)
		return nil, fmt.Errorf("查询监控数据失败: %v", err)
	}

	// 转换数据格式
	var monitorSeries []types.MonitorSeries
	l.Infof("[QueryMetrics] Processing %d series", len(series))
	for i, s := range series {
		l.Infof("[QueryMetrics] Series %d: metric=%v, values count=%d", i, s.Metric, len(s.Values))
		// 打印 metric 中的所有标签键，便于调试
		if i == 0 && len(s.Metric) > 0 {
			keys := make([]string, 0, len(s.Metric))
			for k := range s.Metric {
				keys = append(keys, k)
			}
			l.Infof("[QueryMetrics] Available metric keys: %v", keys)
		}
		var dataPoints []types.DataPoint
		for j, value := range s.Values {
			// 记录前3个数据点用于调试
			if j < 3 {
				l.Infof("[QueryMetrics] Value[%d]: type=%T, value[0]=%v, value[1]=%v", j, value, value[0], value[1])
			}

			// VictoriaMetrics返回的时间戳是秒级，不需要转换
			timestamp := int64(value[0].(float64))

			val := value[1].(string)
			floatVal, err := strconv.ParseFloat(val, 64)
			if err != nil {
				l.Errorf("[QueryMetrics] ParseFloat error: %v", err)
				continue
			}

			if j < 3 {
				l.Infof("[QueryMetrics] Parsed: timestamp=%d, value=%f", timestamp, floatVal)
			}

			dataPoints = append(dataPoints, types.DataPoint{
				Timestamp: timestamp,
				Value:     floatVal,
			})
		}

		instance := extractInstanceFromMetric(s.Metric)
		l.Infof("[QueryMetrics] Series %d: extracted instance/hostname: %s", i, instance)

		monitorSeries = append(monitorSeries, types.MonitorSeries{
			Instance: instance,
			Metric:   "custom",
			Unit:     "",
			Data:     dataPoints,
			Labels:   s.Metric, // 保存所有标签用于前端显示
		})
	}

	l.Infof("[QueryMetrics] Successfully queried metrics with query: %s", req.Query)

	return &types.QueryMetricsResp{
		Series: monitorSeries,
	}, nil
}

// VictoriaMetrics API响应结构
type VMAPIResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string     `json:"resultType"`
		Result     []VMResult `json:"result"`
	} `json:"data"`
}

type VMResult struct {
	Metric map[string]string `json:"metric"`
	Values [][]interface{}   `json:"values"`
}

// queryVMAPI 调用VictoriaMetrics API查询数据
func (l *QueryMetricsLogic) queryVMAPI(baseURL, promQL string, start, end int64, step string) ([]VMResult, error) {
	// 构建请求URL
	reqURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	q := reqURL.Query()
	q.Set("query", promQL)
	q.Set("start", fmt.Sprintf("%d", start))
	q.Set("end", fmt.Sprintf("%d", end))
	q.Set("step", step)
	reqURL.RawQuery = q.Encode()

	l.Debugf("[queryVMAPI] Request URL: %s", reqURL.String())

	resp, err := http.Get(reqURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("VictoriaMetrics API返回错误: %s", string(body))
	}

	// 解析响应
	var apiResp VMAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	if apiResp.Status != "success" {
		return nil, fmt.Errorf("API返回失败状态: %s", apiResp.Status)
	}

	return apiResp.Data.Result, nil
}

// extractInstanceFromMetric 从metric map中提取instance标识
func extractInstanceFromMetric(metric map[string]string) string {
	// 使用 hostname 标签（实际监控数据使用 hostname）
	if hostname, ok := metric["hostname"]; ok {
		return hostname
	}

	// 如果没有 hostname，尝试其他可能的标签
	for key, value := range metric {
		if key == "instance" || key == "hostname" || key == "host" {
			return value
		}
	}

	return "unknown"
}
