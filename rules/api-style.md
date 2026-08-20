# API 规范

## 路由设计
- RESTful 风格：`GET /poems`, `POST /poems`, `GET /poems/:id`
- Admin 路由前缀：`/api/admin/`
- User 路由前缀：`/api/user/`
- 公开路由：`/api/public/`

## 请求格式
- JSON body，Content-Type: `application/json`
- 查询参数通过 URL query 传递
- 分页参数：`page`（默认 1）, `page_size`（默认 10，最大 100）

## 响应格式
```json
{
  "code": 200,
  "message": "success",
  "data": {}
}
```

## 错误码
- `200`: 成功
- `400`: 请求参数错误
- `401`: 未认证
- `403`: 无权限
- `404`: 资源不存在
- `500`: 服务器内部错误
