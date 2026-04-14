# Redmine CLI 命令速查

本 skill 用到的 `redmine` CLI 命令子集。

## 命令总览

| 命令 | 用途 | 关键参数 |
|------|------|---------|
| `redmine issue get <id>` | 获取 issue 详情 | `--format json` |
| `redmine issue update <id>` | 更新 issue（如状态） | `--status-id <ID>` |
| `redmine status list` | 列出所有 issue 状态 | `--format json` |

## 获取 issue 详情

```bash
redmine issue get <id> --format json
```

返回中用到的字段路径：

| 字段路径 | 说明 | 用途 |
|---------|------|------|
| `issue.id` | Issue ID | 分支名中的 issue 号 |
| `issue.subject` | Issue 标题 | 展示给用户确认 |
| `issue.assigned_to.name` | 指派人名称 | 分支名中的作者名 |
| `issue.parent.id` | 父 issue ID | 任务所属需求号（任务分支流程） |
| `issue.tracker.name` | Tracker 名称 | 辅助判断 issue 类型 |
| `issue.status.name` | 当前状态名称 | 展示状态变更前后对比 |

可选参数：`--include children,relations` 获取关联数据。

## 更新 issue 状态

```bash
redmine issue update <id> --status-id <STATUS_ID>
```

仅传需要修改的字段，未传入的保持原值。

## 列出所有状态

```bash
redmine status list --format json
```

返回结构：

```json
{
  "issue_statuses": [
    {"id": 1, "name": "新建", "is_closed": false, "is_default": true},
    {"id": 2, "name": "开发中", "is_closed": false, "is_default": false},
    {"id": 3, "name": "已完成", "is_closed": true, "is_default": false}
  ]
}
```

查找状态 ID：
- 需求流程：找到 `name == "进行中"` 的项，取其 `id`
- 任务流程：找到 `name == "开发中"` 的项，取其 `id`

## 全局标志

| 标志 | 说明 |
|------|------|
| `--format json` | 所有命令必须添加，确保输出可解析 |
| `--instance <name>` | 指定 Redmine 实例（多实例场景） |
| `--verbose` | 详细输出，排错时使用 |
| `--dry-run` | 预览操作不执行 |
