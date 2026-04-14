# 登录流程优化设计

## 1. 交互精简

- 移除 `▶ Step X:` 前缀，改为直接提示
- 移除步骤间的空行，流程更紧凑
- 简化后流程示例：

```
Redmine URL: https://redmine.example.com
正在检查连通性... ✓
API Key: ••••••••
正在验证连接... ✓
实例名称 [default]: mycompany
```

## 2. URL 连通性校验

- 输入 URL 后立即校验 HTTP HEAD 请求
- 校验通过再进入下一步
- 校验失败立即提示，不等待到 API Key 步骤

## 3. 错误提示（快速排查清单）

| 错误类型 | 显示内容 |
|---------|---------|
| URL 格式无效 | `✗ URL 格式无效` |
| URL 无法连通 | `✗ URL 无法访问。请确认：1) URL 正确 2) 网络畅通 3) 服务正常运行` |
| API Key 无效 | `✗ API Key 无效。请确认：1) API Key 正确 2) 前往 Settings → API access 查看` |
| 连接失败（综合） | `✗ 连接失败。请确认：1) URL 可访问 2) API Key 有效 3) 网络畅通` |

## 4. 数据流

```
runLogin()
  └── promptInput("Redmine URL")
      └── validateURL(url) → 连通性校验
          └── promptSecret("API Key")
              └── TestAuth() → 验证 API Key
                  └── promptInput("实例名称")
                      └── saveConfig()
```

## 5. 关键实现点

- `validateURL(url)` 使用 HTTP HEAD 请求，3秒超时
- 复用现有的 `client.Client` 结构，需新增 `Ping(ctx context.Context)` 方法
- 错误分类复用 `internal/errors` 的 validation/auth/api 类型
- 新增 `errors.ErrNetwork` 错误类型用于网络连通性问题
