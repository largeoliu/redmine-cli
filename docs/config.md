# 配置说明

## 配置文件

Redmine CLI 将配置存储在 `~/.redmine-cli/config.yaml`。

可以通过设置环境变量 `REDMINE_CONFIG_DIR` 自定义配置目录。

## 配置结构

```yaml
default: "instance-name"           # 默认实例名称
instances:
  instance-name:
    url: "https://redmine.example.com"
    api_key: "your-api-key"
settings:
  timeout: 30s                     # 请求超时时间
  retries: 3                       # 重试次数
  output_format: "json"            # 默认输出格式
  page_size: 100                   # 分页大小
git:
  auto_link: true                  # Git 集成自动关联
  commit_pattern: `#(\d+)`         # 关联提交的匹配模式
  default_project: 0
report:
  templates_dir: ""                # 报告模板目录
  default_format: "table"          # 默认报告格式
```

## 多实例管理

Redmine CLI 支持管理多个 Redmine 实例。

### 交互式登录

```bash
redmine login
```

交互式提示输入 URL 和 API key，并保存到配置文件。

### 设置默认实例

```bash
redmine config set <instance-name>
```

### 列出所有实例

```bash
redmine config list
```

### 显示当前配置

```bash
redmine config get
```

### 使用非默认实例

使用 `--instance` 参数指定要使用的实例：

```bash
redmine --instance <instance-name> issue list
redmine -u <url> -k <api-key> issue list
```

## 环境变量

| 变量 | 说明 |
|------|------|
| `REDMINE_CONFIG_DIR` | 自定义配置目录路径 |

## 实例配置

每个实例在 `instances` 中需要：

- `url`：Redmine 实例的基础 URL
- `api_key`：你的 Redmine API key

获取 API key 方法：
1. 登录 Redmine
2. 进入"我的账户" → "API 访问密钥"
3. 点击"显示"查看你的 API key