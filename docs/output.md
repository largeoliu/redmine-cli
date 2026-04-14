# 输出格式

## 输出格式

Redmine CLI 支持三种输出格式：`json`、`table` 和 `raw`。

### JSON 格式（默认）

```bash
redmine issue get 123 --format json
```

输出带 2 空格缩进的格式化 JSON。

### Table 格式

```bash
redmine issue list --format table
```

输出带表头和分隔符的可读表格格式。

### Raw 格式

```bash
redmine issue get 123 --format raw
```

输出原始字符串或单行 JSON。

## 字段过滤

使用 `--fields` 选择要输出的特定字段：

```bash
redmine issue list --fields id,subject,status
```

对于数组中的嵌套字段：

```bash
redmine issue list --fields id,custom_fields.name
```

## jq 过滤

使用 `--jq` 通过 jq 语法过滤和转换 JSON 输出：

```bash
# 列出所有 issue 的 id 和 subject
redmine issue list --jq '.issues[] | {id, subject}'

# 提取所有 issue 的 id
redmine issue list --jq '.issues[].id'

# 筛选特定状态的 issue
redmine issue list --jq '.issues[] | select(.status.id == 1)'

# 统计总数量
redmine issue list --jq '.total_count'
```

## 输出到文件

使用 `--output` 将输出写入文件而不是 stdout：

```bash
redmine issue list --output issues.json
redmine issue list --format table --output issues.txt
```

## 分页

使用 `--limit` 和 `--offset` 对结果进行分页：

```bash
# 获取前 50 条
redmine issue list --limit 50

# 获取下一批 50 条
redmine issue list --limit 50 --offset 50
```

## 组合使用

```bash
# JSON 输出 + jq 过滤 + 写入文件
redmine issue list --format json --jq '.issues[] | select(.priority.id >= 3)' --output high_priority.json

# 表格格式 + 指定字段
redmine issue list --format table --fields id,subject,assigned_to --limit 20
```

## 全局参数参考

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `--format`, `-f` | `json` | 输出格式 |
| `--jq` | | jq 过滤表达式 |
| `--fields` | | 要包含的字段 |
| `--output`, `-o` | | 输出文件路径 |
| `--limit`, `-l` | `0` | 限制结果数量 |
| `--offset` | `0` | 分页偏移 |