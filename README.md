# Redmine CLI

AI Agent 友好的 Redmine 命令行工具。

## 功能特性

- **AI Agent 友好**：结构化输出，jq 过滤，易于脚本化和自动化
- **多实例支持**：管理多个 Redmine 实例，一键切换
- **完整 CRUD**：支持 issues、projects、users、versions 等资源的完整操作
- **输出格式灵活**：支持 json、table、raw 三种输出格式
- **分页和过滤**：支持 `--limit`、`--offset` 分页，以及多种过滤条件
- **Dry Run 模式**：预览操作结果，避免误操作
- **自动升级**：一行命令自动升级到最新版本或指定版本

## 安装

### 方式一：一键安装（推荐）

**macOS / Linux：**

```bash
curl -fsSL https://raw.githubusercontent.com/largeoliu/redmine-cli/master/scripts/install.sh | sh
```

**Windows (PowerShell)：**

```powershell
irm https://raw.githubusercontent.com/largeoliu/redmine-cli/master/scripts/install.ps1 | iex
```

### 方式二：npm 安装

```bash
npm install -g redminectl
```

> **注意：** npm 包名为 `redminectl`，但安装后的命令为 `redmine`。

### 方式三：手动下载

1. 访问 [GitHub Releases](https://github.com/largeoliu/redmine-cli/releases) 页面
2. 根据您的平台下载对应的压缩包：
   - **macOS**: `redmine-cli_<version>_darwin_amd64.tar.gz` (Intel) 或 `redmine-cli_<version>_darwin_arm64.tar.gz` (Apple Silicon)
   - **Linux**: `redmine-cli_<version>_linux_amd64.tar.gz` (x64) 或 `redmine-cli_<version>_linux_arm64.tar.gz` (ARM64)
   - **Windows**: `redmine-cli_<version>_windows_amd64.zip`
3. 解压文件：
   - macOS/Linux: `tar -xzf redmine-cli_<version>_<os>_<arch>.tar.gz`
   - Windows: 使用文件资源管理器或 PowerShell 解压 `.zip` 文件
4. 将解压后的 `redmine`（或 Windows 下的 `redmine.exe`）移动到 PATH 中的目录，例如：
   - macOS/Linux: `sudo mv redmine /usr/local/bin/`
   - Windows: 移动到 `C:\Windows\System32\` 或添加到 PATH 的目录
5. macOS/Linux 用户需要添加可执行权限：`chmod +x /usr/local/bin/redmine`

### 方式四：从源码构建

**前提条件：** 需要安装 Go 1.23 或更高版本。

```bash
go install github.com/largeoliu/redmine-cli/cmd@latest
```

安装后的二进制文件位于 `$(go env GOPATH)/bin/redmine`，请确保该目录在您的 PATH 中。

## 升级

安装后，可通过 `upgrade` 命令升级到最新版本：

```bash
# 升级到最新版本
redmine upgrade

# 检查是否有新版本
redmine upgrade --check

# 升级到指定版本
redmine upgrade --version v1.2.3
```

## 快速开始

### 1. 登录

```bash
redmine login
```

交互式输入 Redmine URL 和 API key。

### 2. 配置多实例（可选）

```bash
# 设置默认实例
redmine config set my-redmine

# 查看当前配置
redmine config get

# 列出所有配置的实例
redmine config list
```

### 3. 查询数据

```bash
# 列出所有 issues
redmine issue list

# 列出 projects
redmine project list

# 获取单个 issue 详情
redmine issue get 123

# 列出 users
redmine user list
```

### 4. 创建数据

```bash
# 创建 issue
redmine issue create --project-id 1 --subject "Bug report"

# 创建 project
redmine project create --name "My Project" --identifier "my-project"

# 创建 version
redmine version create --project-id 1 --name "v1.0"
```

### 5. 更新和删除

```bash
# 更新 issue
redmine issue update 123 --status-id 2 --priority-id 3

# 删除 issue（会提示确认）
redmine issue delete 123

# 跳过确认直接删除
redmine issue delete 123 --yes
```

## 配置说明

配置文件位于 `~/.redmine-cli/config.yaml`。

### 多实例配置示例

```yaml
default: "work"
instances:
  work:
    url: "https://redmine.company.com"
    api_key: "your-work-api-key"
  personal:
    url: "https://redmine.example.com"
    api_key: "your-personal-api-key"
```

### 环境变量

| 变量 | 说明 |
|------|------|
| `REDMINE_CONFIG_DIR` | 自定义配置目录路径 |

获取 API key：登录 Redmine → "我的账户" → "API access key" → 点击"显示"

## 输出格式

### 三种格式

```bash
# JSON（默认）
redmine issue list --format json

# 表格
redmine issue list --format table

# 原始输出
redmine issue get 123 --format raw
```

### jq 过滤

```bash
# 提取所有 issue 的 id 和 subject
redmine issue list --jq '.issues[] | {id, subject}'

# 统计数量
redmine issue list --jq '.total_count'

# 筛选高优先级 issue
redmine issue list --jq '.issues[] | select(.priority.id >= 3)'
```

### 字段选择

```bash
# 只显示指定字段
redmine issue list --fields id,subject,status

# 支持嵌套字段
redmine issue list --fields id,custom_fields.name
```

### 输出到文件

```bash
redmine issue list --output issues.json
```

## 分页

```bash
# 获取前 50 条
redmine issue list --limit 50

# 获取第 51-100 条
redmine issue list --limit 50 --offset 50
```

## Dry Run

预览操作结果，不实际执行：

```bash
# 预览创建 issue
redmine issue create --project-id 1 --subject "Test" --dry-run

# 预览删除
redmine issue delete 123 --dry-run
```

## 全局 Flags

| Flag | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--url` | `-u` | | Redmine 实例 URL |
| `--key` | `-k` | | API key |
| `--format` | `-f` | `json` | 输出格式：json、table、raw |
| `--jq` | | | jq 过滤表达式 |
| `--fields` | | | 输出的字段，逗号分隔 |
| `--dry-run` | | `false` | 预览模式 |
| `--yes` | `-y` | `false` | 跳过确认提示 |
| `--output` | `-o` | | 输出文件路径 |
| `--limit` | `-l` | `0` | 限制结果数量 |
| `--offset` | | `0` | 分页偏移 |
| `--timeout` | | `30` | 请求超时秒数 |
| `--retries` | | `3` | 失败重试次数 |
| `--verbose` | `-v` | `false` | 详细输出 |
| `--debug` | | `false` | 调试模式 |
| `--instance` | | | 使用指定实例 |

## 常用命令示例

### Issues

```bash
# 列出所有 issues
redmine issue list

# 按项目筛选
redmine issue list --project-id 1

# 按状态筛选
redmine issue list --status-id 1

# 创建 issue
redmine issue create --project-id 1 --subject "Bug" --description "描述" --priority-id 2

# 更新 issue
redmine issue update 123 --status-id 2

# 删除 issue
redmine issue delete 123 --yes
```

### Projects

```bash
# 列出所有项目
redmine project list

# 获取项目详情
redmine project get 1

# 创建项目
redmine project create --name "新项目" --identifier "new-project"

# 更新项目
redmine project update 1 --name "新名称"

# 删除项目
redmine project delete 1 --yes
```

### Users

```bash
# 列出用户
redmine user list

# 获取当前用户
redmine user get-self

# 创建用户
redmine user create --login "jdoe" --password "pass" --first-name "John" --last-name "Doe" --email "jdoe@example.com"

# 更新用户
redmine user update 1 --email "newemail@example.com"

# 删除用户
redmine user delete 1 --yes
```

### Versions

```bash
# 列出项目版本
redmine version list --project-id 1

# 创建版本
redmine version create --project-id 1 --name "v1.0" --due-date 2024-12-31

# 更新版本
redmine version update 1 --status closed

# 删除版本
redmine version delete 1 --yes
```

### Time Entries

```bash
# 列出时间条目
redmine time-entry list --project-id 1

# 按日期范围筛选
redmine time-entry list --from 2024-01-01 --to 2024-01-31

# 记录时间
redmine time-entry create --issue-id 123 --hours 2.5 --activity-id 9 --comments "Working on X"

# 更新时间条目
redmine time-entry update 1 --hours 3.5

# 删除时间条目
redmine time-entry delete 1 --yes
```

### Categories

```bash
# 列出项目分类
redmine category list --project-id 1

# 创建分类
redmine category create --project-id 1 --name "Bug" --assigned-to-id 5

# 更新分类
redmine category update 1 --name "New Name"

# 删除分类
redmine category delete 1 --yes
```

### Trackers、Statuses、Priorities

```bash
# 列出追踪器
redmine tracker list

# 列出状态
redmine status list

# 列出优先级
redmine priority list
```

### Upgrade

```bash
# 升级到最新版本
redmine upgrade

# 检查是否有新版本
redmine upgrade --check

# 升级到指定版本
redmine upgrade --version v1.2.3
```

## 完整命令参考

完整命令文档请参阅 [docs/](docs/) 目录：

- [docs/README.md](docs/README.md) - 文档入口
- [docs/commands.md](docs/commands.md) - 所有命令详细参考
- [docs/config.md](docs/config.md) - 配置文件格式、多实例、环境变量
- [docs/output.md](docs/output.md) - 输出格式、jq 过滤、字段选择

## License

MIT
Thu Apr 16 12:16:35 AM CST 2026
