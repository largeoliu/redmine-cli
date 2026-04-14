---
name: redmine-workflow
description: "Redmine 开发工作流自动化。输入需求 issue 号自动创建 git 分支并更新状态为进行中；输入任务 issue 号自动更新状态为开发中。当用户提到「开始开发」「创建分支」「开需求」「开任务」或提供 Redmine issue 号时使用。"
---

# Redmine 开发工作流 Skill

通过 `redmine` CLI + `git` 命令自动化开发工作流：需求 issue 创建分支+更新状态为"进行中"，任务 issue 更新状态为"开发中"。

## 严格禁止 (NEVER DO)

- 不要编造 issue ID，必须从 `redmine issue get` 返回中提取
- 不要跳过 issue 查询直接创建分支
- 不要在未验证 issue 存在的情况下更新状态
- 不要使用 `redmine` CLI 以外的方式操作 Redmine（禁止 curl、HTTP API）

## 严格要求 (MUST DO)

- 所有 `redmine` 命令必须加 `--format json` 以获取可解析输出
- 创建分支前必须先查询 issue 信息
- 更新状态前必须先通过 `redmine status list` 确认目标状态 ID（需求→"进行中"，任务→"开发中"）

## 前提条件

- 已通过 `redmine login` 配置好 Redmine 实例（url + api_key）
- 当前目录是 git 仓库

## 意图判断

| 用户说... | 工作流 | 说明 |
|-----------|--------|------|
| "开始需求/开需求/需求分支" + issue号 | 需求分支 | 创建 `feature/{作者名}/{issue号}`，更新状态为"进行中" |
| "开始任务/开任务/开发任务" + issue号 | 任务开发 | 查询 issue 信息，将状态更新为"开发中" |
| 只给 issue号，未说明类型 | 自动判断 | 查询 issue，有 parent → 任务开发流程；无 parent → 需求分支流程 |

## 核心流程

1. **查询 issue**：`redmine issue get <id> --format json`，提取 subject、assigned_to、parent、tracker、status
2. **判断类型**：有 parent 字段 → 任务开发流程；无 parent → 需求分支流程
3. **需求分支流程**：格式化作者名 → `git checkout -b feature/{作者名}/{issue号}` → 查找"进行中"状态 ID → `redmine issue update <id> --status-id <ID>`
4. **任务开发流程**：查找"开发中"状态 ID → `redmine issue update <id> --status-id <ID>`

> 详细步骤见 [branch-workflow.md](./references/branch-workflow.md)

## 错误处理

| 场景 | 处理方式 |
|------|---------|
| Issue 不存在 | 报告错误信息，不创建分支 |
| 任务无 parent 字段 | 提示用户确认：按需求分支处理，还是手动提供需求号 |
| 目标状态（"进行中"/"开发中"）未找到 | 展示 `redmine status list` 全部结果，让用户指定 |
| 分支已存在 | 提示 checkout 到已有分支，或确认创建新分支 |
| 未配置 Redmine | 提示运行 `redmine login` 完成配置 |
| 不在 git 仓库 | 报告错误，提示在 git 仓库目录中执行 |

## 参考文档

- [references/branch-workflow.md](./references/branch-workflow.md) — 需求分支与任务开发工作流详细步骤
- [references/redmine-commands.md](./references/redmine-commands.md) — 本 skill 用到的 redmine CLI 命令速查
