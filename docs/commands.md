# 命令参考

## 目录

- [全局参数](#全局参数)
- [认证](#认证)
- [Upgrade](#upgrade)
- [配置](#配置)
- [Agile](#agile)
- [Sprint](#sprint)
- [Issues](#issues)
- [Projects](#projects)
- [Users](#users)
- [Versions](#versions)
- [Time Entries](#time-entries)
- [Categories](#categories)
- [Trackers](#trackers)
- [Statuses](#statuses)
- [Priorities](#priorities)

---

## 全局参数

适用于所有命令：

| 参数 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--url` | `-u` | | Redmine 实例 URL |
| `--key` | `-k` | | API key |
| `--format` | `-f` | `json` | 输出格式：`json`、`table`、`raw` |
| `--jq` | | | jq 过滤表达式，用于 JSON 转换 |
| `--fields` | | | 输出的字段，逗号分隔 |
| `--dry-run` | | `false` | 预览模式，不实际执行 |
| `--yes` | `-y` | `false` | 跳过确认提示 |
| `--output` | `-o` | | 输出文件路径 |
| `--limit` | `-l` | `0` | 限制结果数量 |
| `--offset` | | `0` | 分页偏移 |
| `--timeout` | | `30` | 请求超时秒数 |
| `--retries` | | `3` | 失败重试次数 |
| `--verbose` | `-v` | `false` | 详细输出 |
| `--debug` | | `false` | 调试模式 |
| `--instance` | | | 使用配置中指定的实例名称 |

---

## 认证

### login

交互式登录 Redmine 实例。

```bash
redmine login
```

交互式提示输入 URL 和 API key，并保存到配置文件。

---

## Upgrade

### upgrade

升级 redmine CLI 到最新版本。

```bash
redmine upgrade
```

**示例：**

```bash
# 升级到最新版本
redmine upgrade

# 仅检查更新，不执行升级
redmine upgrade --check

# 升级到指定版本
redmine upgrade --version v1.2.3
```

**参数：**

| 参数 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--check` | | `false` | 仅检查更新，不执行升级 |
| `--version` | | | 升级到指定版本 |

---

## 配置

### config get

显示当前配置。

```bash
redmine config get
```

### config set

设置默认实例。

```bash
redmine config set <instance-name>
```

### config list

列出所有配置的实例。

```bash
redmine config list
```

---

## Agile

### agile board

显示项目当前 Sprint 或指定 Sprint 中的 issue 内容，支持按 tracker 过滤。

```bash
redmine agile board city --format raw
redmine agile board 42 --sprint 8 --tracker 需求 --format table
```

**参数：**

| 参数 | 说明 |
|------|------|
| `<project>` | 项目 ID 或 identifier |
| `--sprint` | `current` 或 sprint ID |
| `--tracker` | tracker 名称，`全部` 表示不过滤 |

**输出：**

- `raw`：文本分组输出，按 Sprint 展示卡片内容
- `table`：扁平表格输出，包含 Sprint 列
- `json`：结构化报告，包含 `project`、`current_sprint`、`groups` 和 `cards`，其中 `current_sprint` 表示当前展示的 sprint

---

## Sprint

### sprint list

列出项目的 sprint 列表。

```bash
redmine sprint list city --format table
redmine sprint list 42 --format json
redmine sprint list city --details --format table
```

**参数：**

| 参数 | 说明 |
|------|------|
| `<project>` | 项目 ID 或 identifier |
| `--details` | 展开每个 sprint 为完整详情后输出 |

**输出：**

- `json`：sprint 数组，`--details` 时包含完整字段
- `table`：展示 sprint 的全部字段
- `raw`：单行 JSON

---

## Issues

### issue list

列出 issues。

```bash
redmine issue list
redmine issues list
```

**参数：**

| 参数 | 说明 |
|------|------|
| `--project-id` | 按项目 ID 筛选 |
| `--status-id` | 按状态 ID 筛选 |
| `--tracker` | 按追踪器名称筛选，`全部` 表示不过滤 |
| `--tracker-id` | 按追踪器 ID 筛选 |
| `--assigned-to-id` | 按 assignee ID 筛选 |
| `--include` | 包含关联数据（children、attachments、relations、changesets、journals、watchers） |

**示例：**

```bash
# 列出所有 issues
redmine issue list

# 按项目筛选
redmine issue list --project-id 1

# 按状态筛选
redmine issue list --status-id 1

# 按追踪器名称筛选
redmine issue list --tracker 需求

# 不筛选追踪器
redmine issue list --tracker 全部

# 包含关联数据
redmine issue list --include relations,journals

# 分页
redmine issue list --limit 50 --offset 100
```

### issue get

获取 issue 详情。

```bash
redmine issue get <id>
redmine issues get <id>
```

**参数：**

| 参数 | 说明 |
|------|------|
| `--include` | 包含关联数据 |

**示例：**

```bash
redmine issue get 123
redmine issue get 123 --include relations,attachments
```

### issue create

创建新 issue。

```bash
redmine issue create [flags]
redmine issues create [flags]
```

**参数：**

| 参数 | 说明 |
|------|------|
| `--project-id`（必需）| 项目 ID |
| `--subject`（必需）| Issue 标题 |
| `--description` | Issue 描述 |
| `--tracker-id` | 追踪器 ID |
| `--status-id` | 状态 ID |
| `--priority-id` | 优先级 ID |
| `--assigned-to-id` | 指派用户 ID |
| `--category-id` | 分类 ID |
| `--version-id` | 版本 ID |
| `--parent-issue-id` | 父 issue ID |
| `--watchers` | 关注者用户 ID 列表 |

**示例：**

```bash
# 最小参数创建
redmine issue create --project-id 1 --subject "Bug report"

# 完整参数创建
redmine issue create \
  --project-id 1 \
  --subject "Bug report" \
  --description "Found a bug" \
  --tracker-id 1 \
  --priority-id 2 \
  --assigned-to-id 5

# 预览模式
redmine issue create --project-id 1 --subject "Test" --dry-run
```

### issue update

更新现有 issue。

```bash
redmine issue update <id> [flags]
redmine issues update <id> [flags]
```

**参数：**

| 参数 | 说明 |
|------|------|
| `--subject` | Issue 标题 |
| `--description` | Issue 描述 |
| `--tracker-id` | 追踪器 ID |
| `--status-id` | 状态 ID |
| `--priority-id` | 优先级 ID |
| `--assigned-to-id` | 指派用户 ID |
| `--category-id` | 分类 ID |
| `--version-id` | 版本 ID |
| `--parent-issue-id` | 父 issue ID |

**示例：**

```bash
redmine issue update 123 --status-id 2
redmine issue update 123 --priority-id 3 --assigned-to-id 5
```

### issue delete

删除 issue。

```bash
redmine issue delete <id>
redmine issues delete <id>
```

**示例：**

```bash
# 删除（会提示确认）
redmine issue delete 123

# 跳过确认
redmine issue delete 123 --yes

# 预览模式
redmine issue delete 123 --dry-run
```

---

## Projects

### project list

列出项目。

```bash
redmine project list
redmine projects list
```

**参数：**

| 参数 | 说明 |
|------|------|
| `--include` | 包含关联数据（trackers、issue_categories、enabled_modules） |

**示例：**

```bash
redmine project list
redmine project list --include trackers,issue_categories
```

### project get

获取项目详情。

```bash
redmine project get <id>
redmine projects get <id>
```

**示例：**

```bash
redmine project get 1
```

### project create

创建新项目。

```bash
redmine project create [flags]
redmine projects create [flags]
```

**参数：**

| 参数 | 说明 |
|------|------|
| `--name`（必需）| 项目名称 |
| `--identifier`（必需）| 项目标识符 |
| `--description` | 项目描述 |
| `--tracker-ids` | 追踪器 ID 列表 |
| `--enabled-module-names` | 启用的模块名称列表 |
| `--issue-category-names` | Issue 分类名称列表 |

**示例：**

```bash
redmine project create --name "My Project" --identifier "my-project"
```

### project update

更新项目。

```bash
redmine project update <id> [flags]
redmine projects update <id> [flags]
```

**参数：**

| 参数 | 说明 |
|------|------|
| `--name` | 项目名称 |
| `--description` | 项目描述 |
| `--enabled-module-names` | 启用的模块名称列表 |
| `--issue-category-names` | Issue 分类名称列表 |

**示例：**

```bash
redmine project update 1 --name "New Name"
```

### project delete

删除项目。

```bash
redmine project delete <id>
redmine projects delete <id>
```

**示例：**

```bash
redmine project delete 1 --yes
```

---

## Users

### user list

列出用户。

```bash
redmine user list
redmine users list
```

**参数：**

| 参数 | 说明 |
|------|------|
| `--status` | 按状态筛选 |
| `--name` | 按名称筛选 |
| `--group-id` | 按用户组 ID 筛选 |
| `--include` | 包含关联数据（memberships、groups） |

**示例：**

```bash
redmine user list
redmine user list --status 1
redmine user list --include groups
```

### user get

获取用户详情。

```bash
redmine user get <id>
redmine users get <id>
```

**示例：**

```bash
redmine user get 1
```

### user get-self

获取当前认证用户信息。

```bash
redmine user get-self
```

**示例：**

```bash
redmine user get-self
```

### user create

创建新用户。

```bash
redmine user create [flags]
redmine users create [flags]
```

**参数：**

| 参数 | 说明 |
|------|------|
| `--login`（必需）| 用户登录名 |
| `--password`（必需）| 用户密码 |
| `--first-name`（必需）| 名 |
| `--last-name`（必需）| 姓 |
| `--email`（必需）| 邮箱地址 |
| `--admin` | 设为管理员 |
| `--auth-source-id` | 认证源 ID |

**示例：**

```bash
redmine user create \
  --login "jdoe" \
  --password "secret" \
  --first-name "John" \
  --last-name "Doe" \
  --email "jdoe@example.com"
```

### user update

更新用户。

```bash
redmine user update <id> [flags]
redmine users update <id> [flags]
```

**参数：**

| 参数 | 说明 |
|------|------|
| `--first-name` | 名 |
| `--last-name` | 姓 |
| `--email` | 邮箱地址 |
| `--admin` | 设为管理员 |
| `--password` | 密码 |

**示例：**

```bash
redmine user update 1 --email "newemail@example.com"
```

### user delete

删除用户。

```bash
redmine user delete <id>
redmine users delete <id>
```

**示例：**

```bash
redmine user delete 1 --yes
```

---

## Versions

### version list

列出项目版本。

```bash
redmine version list
redmine versions list
```

**参数：**

| 参数 | 说明 |
|------|------|
| `--project-id`（必需）| 项目 ID |

**示例：**

```bash
redmine version list --project-id 1
```

### version get

获取版本详情。

```bash
redmine version get <id>
redmine versions get <id>
```

**示例：**

```bash
redmine version get 1
```

### version create

创建新版本。

```bash
redmine version create [flags]
redmine versions create [flags]
```

**参数：**

| 参数 | 说明 |
|------|------|
| `--project-id`（必需）| 项目 ID |
| `--name`（必需）| 版本名称 |
| `--description` | 版本描述 |
| `--due-date` | 截止日期（YYYY-MM-DD）|
| `--status` | 版本状态（open、locked、closed）|
| `--sharing` | 版本共享（none、descendants、hierarchy、tree、system）|

**示例：**

```bash
redmine version create --project-id 1 --name "v1.0"
redmine version create --project-id 1 --name "v1.0" --due-date 2024-12-31 --status open
```

### version update

更新版本。

```bash
redmine version update <id> [flags]
redmine versions update <id> [flags]
```

**参数：**

| 参数 | 说明 |
|------|------|
| `--name` | 版本名称 |
| `--description` | 版本描述 |
| `--due-date` | 截止日期 |
| `--status` | 版本状态 |
| `--sharing` | 版本共享 |

**示例：**

```bash
redmine version update 1 --status closed
```

### version delete

删除版本。

```bash
redmine version delete <id>
redmine versions delete <id>
```

**示例：**

```bash
redmine version delete 1 --yes
```

---

## Time Entries

### time-entry list

列出时间条目。

```bash
redmine time-entry list
redmine time-entries list
```

**参数：**

| 参数 | 说明 |
|------|------|
| `--project-id` | 按项目 ID 筛选 |
| `--issue-id` | 按 issue ID 筛选 |
| `--user-id` | 按用户 ID 筛选 |
| `--from` | 开始日期（YYYY-MM-DD）|
| `--to` | 结束日期（YYYY-MM-DD）|
| `--spent-on` | 具体花费日期 |

**示例：**

```bash
redmine time-entry list
redmine time-entry list --project-id 1
redmine time-entry list --from 2024-01-01 --to 2024-01-31
```

### time-entry get

获取时间条目详情。

```bash
redmine time-entry get <id>
redmine time-entries get <id>
```

**示例：**

```bash
redmine time-entry get 1
```

### time-entry create

创建时间条目。

```bash
redmine time-entry create [flags]
redmine time-entries create [flags]
```

**参数：**

| 参数 | 说明 |
|------|------|
| `--issue-id` | Issue ID |
| `--project-id` | 项目 ID |
| `--hours`（必需）| 花费小时数 |
| `--activity-id` | 活动 ID |
| `--spent-on` | 花费日期（YYYY-MM-DD）|
| `--comments` | 备注 |

**示例：**

```bash
redmine time-entry create --issue-id 123 --hours 2.5 --activity-id 9
redmine time-entry create --project-id 1 --hours 1 --comments "Working on X"
```

### time-entry update

更新时间条目。

```bash
redmine time-entry update <id> [flags]
redmine time-entries update <id> [flags]
```

**参数：**

| 参数 | 说明 |
|------|------|
| `--hours` | 花费小时数 |
| `--activity-id` | 活动 ID |
| `--spent-on` | 花费日期 |
| `--comments` | 备注 |

**示例：**

```bash
redmine time-entry update 1 --hours 3.5
```

### time-entry delete

删除时间条目。

```bash
redmine time-entry delete <id>
redmine time-entries delete <id>
```

**示例：**

```bash
redmine time-entry delete 1 --yes
```

---

## Categories

### category list

列出 issue 分类。

```bash
redmine category list
redmine categories list
```

**参数：**

| 参数 | 说明 |
|------|------|
| `--project-id`（必需）| 项目 ID |

**示例：**

```bash
redmine category list --project-id 1
```

### category get

获取分类详情。

```bash
redmine category get <id>
redmine categories get <id>
```

**示例：**

```bash
redmine category get 1
```

### category create

创建 issue 分类。

```bash
redmine category create [flags]
redmine categories create [flags]
```

**参数：**

| 参数 | 说明 |
|------|------|
| `--project-id`（必需）| 项目 ID |
| `--name`（必需）| 分类名称 |
| `--assigned-to-id` | 默认指派用户 ID |

**示例：**

```bash
redmine category create --project-id 1 --name "Bug"
redmine category create --project-id 1 --name "Feature" --assigned-to-id 5
```

### category update

更新分类。

```bash
redmine category update <id> [flags]
redmine categories update <id> [flags]
```

**参数：**

| 参数 | 说明 |
|------|------|
| `--name` | 分类名称 |
| `--assigned-to-id` | 默认指派用户 ID |

**示例：**

```bash
redmine category update 1 --name "New Name"
```

### category delete

删除分类。

```bash
redmine category delete <id>
redmine categories delete <id>
```

**示例：**

```bash
redmine category delete 1 --yes
```

---

## Trackers

### tracker list

列出追踪器。

```bash
redmine tracker list
redmine trackers list
```

**示例：**

```bash
redmine tracker list
```

---

## Statuses

### status list

列出 issue 状态。

```bash
redmine status list
redmine statuses list
```

**示例：**

```bash
redmine status list
```

---

## Priorities

### priority list

列出 issue 优先级。

```bash
redmine priority list
redmine priorities list
```

**示例：**

```bash
redmine priority list
```

---

## 帮助

### help

获取任何命令的帮助。

```bash
redmine help [command]
```

**示例：**

```bash
redmine help issue
redmine help issue create
redmine help project list
```
