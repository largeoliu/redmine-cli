# 文档

## 目录

- [命令参考](commands.md) - 所有可用命令和参数
- [配置说明](config.md) - 配置文件和环境变量
- [输出格式](output.md) - 输出格式、jq 过滤、字段选择
- [工程文档](engineering/README.md) - 仓库内部开发约定、结构与实现规则

## 快速链接

### 常用命令

```bash
# 登录
redmine login

# 列出 issues
redmine issue list

# 创建 issue
redmine issue create --project-id 1 --subject "Bug report"

# 获取 issue 详情
redmine issue get 123

# 查看 agile 内容
redmine agile board city --format raw
redmine agile board city --sprint 8 --tracker 需求 --format raw

# 查看 sprint 列表
redmine sprint list city --format table
redmine sprint list city --details --format table
```

### 输出控制

```bash
# JSON 输出（默认）
redmine issue list --format json

# 表格输出
redmine issue list --format table

# jq 过滤
redmine issue list --jq '.issues[] | {id, subject}'

# 指定字段
redmine issue list --fields id,subject,status
```

### 配置管理

```bash
# 设置默认实例
redmine config set my-instance

# 列出所有配置的实例
redmine config list

# 显示当前配置
redmine config get
```
