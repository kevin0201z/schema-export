# API 规范

> AI 快速摘要
>
> - 适用场景：需要新增或更新 API 文档时读取。
> - 必须产出：请求、响应、错误返回、权限要求和验证要求。
> - 硬性阻断：接口语义未确认前，不要凭空补字段或错误码。
> - 相关规则：`project-kit/ai-guide/05-contract-sync.md`
> - 可跳过场景：当前任务不涉及 API 或外部接口。

## 1. 通用规则

- 基础路径：
- API 版本：
- 鉴权方式：
- 请求格式：JSON
- 响应格式：JSON
- 时间格式：
- 幂等性规则：
- 兼容性策略：

## 2. 通用响应结构

```json
{
  "data": {},
  "message": "string",
  "requestId": "string"
}
```

## 3. 通用错误结构

```json
{
  "message": "string",
  "code": "string",
  "requestId": "string",
  "details": {}
}
```

## 4. 分页规则

- 分页参数：
- 排序参数：
- 默认页大小：
- 最大页大小：

## 5. 错误码表

| 错误码 | HTTP 状态码 | 场景 | 处理建议 |
|---|---|---|---|
|  |  |  |  |

## 6. 接口列表

### 接口名称

`METHOD /path`

#### 说明

#### 权限

#### 幂等性

#### Request Headers

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
|  |  | 是 / 否 |  |

#### Request Params

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
|  |  | 是 / 否 |  |

#### Request Query

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
|  |  | 是 / 否 |  |

#### Request Body

```json
{}
```

#### Response 200

```json
{}
```

#### Response Fields

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
|  |  | 是 / 否 |  |

#### Error Responses

| 状态码 | 错误码 | 场景 | 返回 |
|---|---|---|---|
| 400 | INVALID_REQUEST | 参数错误 | `{ "message": "Invalid request", "code": "INVALID_REQUEST" }` |

#### 验证要求

- 成功响应：
- 失败响应：
- 权限场景：
