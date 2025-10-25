总体任务
1. github action 编包（编译后可在ubuntu执行的二进制文件，加项目 backend/etc/hackathon-api.yaml文件） + 上传kodo
2. ai智能解析发布后的 node exporter 采集到的机器指标上报到 Prometheus，生成发布失败原因分析报告
   1. 机器指标采集使用 Prometheus 系的 node exporter，触发生成报告的函数代表触发了 Prometheus 的告警阈值，代表发布有异常
   2. 拿到入参后，调用ai接口生成分析报告，需要编写提示词
      - 分为多种异常场景：物理机资源问题，go runtime问题等等待补充（根据 指标 可判断的异常）
      - 是否有mcp能够让ai主动调用 Prometheus 获取实时指标，然后进行分析
        - 若没有mcp，当ai需要新的数据时，需要定义好提示词限制输出格式
      - ai生成的报告需要包含：异常指标列表，异常原因分析，解决建议
3. mock 不同异常场景的 node exporter 指标数据，验证ai生成报告的准确性