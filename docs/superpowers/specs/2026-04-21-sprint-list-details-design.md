# Sprint List Details 设计文档

## 概述

为 `redmine sprint list <project>` 添加可选的 `--details` 标志，使命令能够返回完整的 sprint 详情，而不是轻量级的列表数据。

## 问题

当前 sprint 列表返回的数据过于简单，实用性不足。它只暴露了每个 sprint 的少量字段，导致 `table` 输出很窄，且用户在查看 sprint 历史或状态时期望的字段都缺失了。

## 设计

### 命令行为

- `redmine sprint list <project>` 保持为入口点。
- 添加 `--details` 作为布尔标志，默认为 `false`。
- 不使用 `--details` 时，命令返回列表端点中的 sprint 切片。
- 使用 `--details` 时，命令：
  - 解析项目
  - 调用 `GET /projects/{projectID}/agile_sprints.json`
  - 按列表顺序提取 sprint ID
  - 使用 `GET /projects/{projectID}/agile_sprints/{id}.json` 获取每个 sprint 的详情
  - 用详细的 sprint 数据替换每个轻量级的 sprint
  - 在最终输出中保持原始列表顺序

### 字段覆盖

启用 `--details` 时，每个 `Sprint` 应包含 API 已暴露的完整详情集：

- `id`
- `name`
- `description`
- `status`
- `start_date`
- `end_date`
- `goal`
- `is_default`
- `is_closed`
- `is_archived`

该命令不应引入第二种摘要模式或单独的详情子命令。

### 输出格式

- `json` 返回 `[]Sprint`
- `table` 使用所有导出的字段渲染相同的 `[]Sprint`

不需要自定义格式化器。共享的输出层已经将切片转换为表格行。

### 数据流

1. 从 `<project>` 解析项目。
2. 从项目 sprint 列表端点获取 sprint ID。
3. 如果 `--details` 被禁用，立即返回轻量级切片。
4. 如果 `--details` 被启用，使用现有的批量辅助模式并行获取 sprint 详情。
5. 将详细的 payload 按原始切片顺序合并回去。
6. 通过共享输出路径写入最终的 sprint 切片。

### 错误处理

- 如果项目解析失败，返回现有的项目查询错误。
- 如果 sprint 列表请求失败，原样返回该 API 错误。
- 如果任何 sprint 详情请求失败，中止命令并返回该 sprint 的错误。
- 启用详情扩展时不输出部分结果。

## 实现注意事项

- 更新 `Sprint` 以包含 `Description` 字段。
- 保持 sprint 列表响应解码与 `agile_sprints` 和 `sprints` 顶级键兼容。
- 重用现有的 agile 客户端方法，而不是添加新的 HTTP 层。
- 将命令保留在顶级 `sprint` 命令下，`sprints` 作为别名。

## 测试计划

- 根命令包含新的 `sprint` 顶级命令。
- `sprint list <project>` 解析项目并返回 sprint 切片。
- `--details` 触发每个 sprint 的详情获取。
- 最终结果保持原始 sprint 顺序。
- Sprint 详情解码包含 `description` 和其他完整字段。
- Sprint 列表解码接受真实的 `sprints` payload 格式。
- `table` 输出在 `--details` 开启时接收扩展的 sprint 字段。

## 假设

- 默认行为保持轻量级以保证速度。
- `--details` 是唯一的扩展机制。
- 列表端点的顺序是权威顺序，不应在扩展后重新排序。
