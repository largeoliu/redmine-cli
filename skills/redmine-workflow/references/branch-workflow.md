# 开发工作流

## 工作流 1：需求分支

输入：需求 issue 号

```bash
# Step 1 — 查询 issue 信息
redmine issue get <ISSUE_ID> --format json
# 提取: subject, assigned_to.name, tracker.name
# 验证: issue 存在
```

```bash
# Step 2 — 格式化作者名
# assigned_to.name 处理规则:
#   英文: 小写 + 空格替换为下划线 ("John Smith" → "john_smith")
#   中文: 保持原样
```

```bash
# Step 3 — 创建分支
git checkout -b feature/{作者名}/{issue号}
# 示例: git checkout -b feature/john_smith/12345
```

```bash
# Step 4 — 查找"进行中"状态 ID
redmine status list --format json
# 从返回的 issue_statuses 数组中找到 name 为"进行中"的项，提取其 id
```

```bash
# Step 5 — 更新需求状态为"进行中"
redmine issue update <ISSUE_ID> --status-id <进行中ID>
```

```
Step 6 — 向用户确认
展示: 分支名、issue 标题、指派人、状态变更结果
```

### 示例

```bash
# 用户说: "开始需求 12345"

# 1. 查询 issue
redmine issue get 12345 --format json
# 返回: {"id":12345, "subject":"用户登录模块", "assigned_to":{"id":5,"name":"Zhang San"}, ...}

# 2. 格式化作者名: "Zhang San" → "zhang_san"

# 3. 创建分支
git checkout -b feature/zhang_san/12345

# 4. 查找"进行中"状态 ID
redmine status list --format json
# 返回: {"issue_statuses":[{"id":1,"name":"新建"},{"id":2,"name":"进行中"},{"id":3,"name":"已完成"}]}
# "进行中" → id = 2

# 5. 更新状态
redmine issue update 12345 --status-id 2

# 6. 确认
# ✓ 分支 feature/zhang_san/12345 已创建
# Issue #12345: 用户登录模块
# 指派人: Zhang San
# 状态已更新: 新建 → 进行中
```

---

## 工作流 2：任务开发

输入：任务 issue 号

```bash
# Step 1 — 查询任务 issue 信息
redmine issue get <TASK_ID> --format json
# 提取: subject, assigned_to.name, parent.id, tracker.name, status.name
```

```bash
# Step 2 — 验证 parent 字段
# parent 存在 → 提取 parent.id 作为需求号（用于展示关联信息）
# parent 不存在 → 提示用户确认是否为独立任务
```

```bash
# Step 3 — 查找"开发中"状态 ID
redmine status list --format json
# 从返回的 issue_statuses 数组中找到 name 为"开发中"的项，提取其 id
```

```bash
# Step 4 — 更新任务状态为"开发中"
redmine issue update <TASK_ID> --status-id <开发中ID>
```

```
Step 5 — 向用户确认
展示: issue 标题、父需求号、状态变更结果
```

### 示例

```bash
# 用户说: "开始任务 12350"

# 1. 查询任务 issue
redmine issue get 12350 --format json
# 返回: {"id":12350, "subject":"实现登录接口", "assigned_to":{"id":5,"name":"Zhang San"},
#         "parent":{"id":12345}, "status":{"id":1,"name":"新建"}, ...}

# 2. parent.id = 12345（需求号）

# 3. 查找"开发中"状态 ID
redmine status list --format json
# 返回: {"issue_statuses":[{"id":1,"name":"新建"},{"id":2,"name":"开发中"},{"id":3,"name":"已完成"}]}
# "开发中" → id = 2

# 4. 更新状态
redmine issue update 12350 --status-id 2

# 5. 确认
# ✓ Issue #12350: 实现登录接口（父需求 #12345）
# 状态已更新: 新建 → 开发中
```

---

## 自动类型判断

当用户只给了 issue 号、未说明是需求还是任务时：

```bash
# Step 1 — 查询 issue
redmine issue get <ID> --format json

# Step 2 — 判断类型
# 有 parent 字段 → 视为任务，走工作流 2（更新状态）
# 无 parent 字段 → 视为需求，走工作流 1（创建分支 + 更新状态为"进行中"）
```
