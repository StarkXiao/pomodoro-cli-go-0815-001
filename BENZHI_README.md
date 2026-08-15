# BENZHI README

## 项目说明

这是一个使用 Go 标准库开发的本地番茄钟统计工具，支持：

- 自定义专注时长、短休息、长休息和长休触发规则
- 记录任务名称、开始时间、结束时间、计划专注时长和完成状态
- 使用本地 JSON 文件持久化数据
- 输出每日汇总、每周报告和终端进度提示
- 使用锁文件避免多个写入命令同时覆盖数据

## 标准命令

```bash
go build ./...
go run ./cmd/pomodoro help
go test ./...
```

## 示例命令

```bash
go run ./cmd/pomodoro config set --focus 25m --short-break 5m --long-break 15m --long-break-every 4
go run ./cmd/pomodoro start --task "撰写周报"
go run ./cmd/pomodoro summary daily --date 2026-08-15
go run ./cmd/pomodoro report weekly --date 2026-08-15
go run ./cmd/pomodoro unlock
```
