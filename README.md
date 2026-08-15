# Pomodoro CLI

一个使用 Go 编写的本地番茄钟统计工具，支持自定义专注时长、短休息和长休息规则，按 JSON 文件持久化保存数据，并提供日报、周报和终端进度提示。

## 功能

- 自定义默认专注时长、短休息、长休息和长休触发规则
- 启动专注计时并实时显示终端进度
- 记录每次专注的任务名称、开始时间、结束时间、计划时长和完成状态
- 将数据保存到本地 JSON 文件，默认路径为 `./data/pomodoro.json`
- 提供每日汇总和每周报告
- 对缺失数据文件、非法参数、损坏 JSON 等常见边界情况提供明确错误信息

## 依赖说明

- Go 1.26 或更高版本
- 仅使用 Go 标准库，无第三方依赖

## 启动步骤

```bash
go build -o pomodoro ./cmd/pomodoro
./pomodoro config set --focus 25m --short-break 5m --long-break 15m --long-break-every 4
./pomodoro start --task "撰写需求文档"
```

为了方便快速体验，也可以设置秒级时长：

```bash
./pomodoro start --task "演示用任务" --focus 10s --short-break 3s --long-break 6s --long-break-every 2
```

## 常用命令

```bash
./pomodoro config show
./pomodoro summary daily --date 2026-08-15
./pomodoro report weekly --date 2026-08-15
./pomodoro start --task "修复边界情况" --focus 30m
./pomodoro unlock
```

## 锁文件恢复

为了避免多个写入实例同时覆盖本地 JSON，写操作会短暂使用 `*.lock` 锁文件。如果程序异常退出导致锁文件残留，可以执行：

```bash
./pomodoro unlock
```

## 数据格式

程序会在本地 JSON 中保存：

- 当前默认配置
- 长休计数运行状态
- 每条专注记录的任务名、开始时间、结束时间、计划时长、完成状态和下一次休息建议

## 测试

```bash
go test ./...
```
