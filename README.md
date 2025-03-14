# S3 获取预签名接口 Server

## 构建和运行方式，请按版本查看具体说明

# S3预签名URL API文档

该API用于生成S3对象的预签名URL，支持上传(PUT)和下载(GET)操作。

## 请求方式

- **URL**: `/presigned-url`
- **方法**: `GET`
- **示例**: `http://127.0.0.1:1998/presigned-url?key=tmp/test&content-length=231703`

## 请求参数

| 参数名 | 类型 | 必填 | 默认值 | 说明 | 作用范围
|-------|------|-----|-------|------|-----|
| key | String | 是 | - | 对象的键名/路径，用于在S3存储桶中标识对象 | `PUT` & `GET` |
| content-length | Int | 否 | 不限制 | 上传内容的长度（字节），范围：大于0且不超过1GB(1024MB) | `PUT` |
| no-wait | Int | 否 | 0 | 开启"即传即收"功能的等待超时时间（秒），最大值为10秒 | `GET` |
| max-requests | Int | 否 | 0 | 最大下载次数限制，指定URL可被访问的最大次数 | `PUT` & `GET` |
| expire | Int | 否 | 3600 | URL的有效期（秒），默认为1小时 | `PUT` & `GET` |
| force-download | Bool | 否 | false | 是否强制下载（设置为true时会添加attachment响应头） | `GET` |
| limit-rate | Int | 否 | 0 | 单线程限速值（字节/秒） | `GET` |

## 参数详细说明

### key
- **描述**: 对象在S3存储桶中的唯一标识符
- **格式**: 可包含路径，如 `img/u1000.jpg` (首字符不以 / 开头)
- **限制**: 不能为空

### content-length
- **描述**: 指定上传内容的长度
- **用途**: 用于PUT请求的`ContentLength`字段
- **限制**: 如果指定，必须大于0且不超过1GB（可以根据业务需求修改）
- **特殊情况**: 不提供此参数时表示不限制内容长度

### no-wait
- **描述**: 开启"即传即收"功能的等待超时时间
- **功能**: 启用simul-transfer功能，允许在上传过程中同时下载内容
- **限制**: 最大值为10秒，超过会被截断为10秒
- **用法**: 值大于0时生效

### max-requests
- **描述**: 限制URL可被访问的最大次数
- **功能**: 通过`x-bitiful-max-requests`头控制下载次数
- **用法**: 值大于0时生效

### expire
- **描述**: 预签名URL的有效期
- **默认值**: 3600秒（1小时）
- **范围**: 任意正整数

### force-download
- **描述**: 是否强制浏览器下载而非在线预览 (注意：缤纷云免费账户会强制下载)
- **功能**: 设置为true时添加`response-content-disposition=attachment`头
- **值类型**: 布尔值（true/false）

### limit-rate
- **描述**: 单线程传输速率限制
- **功能**: 通过`x-bitiful-limit-rate`头控制传输速度
- **单位**: 字节/秒 (取值 1024 ~ 104857600 之间)
- **用法**: 值 >= 1024 时生效

## 响应

### 成功响应

- **状态码**: 200 OK
- **内容类型**: application/json
- **CORS头**:
  - Access-Control-Allow-Origin: *
  - Access-Control-Allow-Methods: *
  - Access-Control-Allow-Headers: *

**响应示例**:
```json
{
  "get-url": "https://s3.bitiful.net/fanfan/example.jpg?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=...",
  "put-url": "https://s3.bitiful.net/fanfan/example.jpg?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=..."
}
```

### 错误响应

- **状态码**: 400 Bad Request
  - 原因: key为空或content-length无效（如果提供）
- **状态码**: 500 Internal Server Error
  - 原因: 生成预签名URL时发生错误

## 使用示例

### 基本用法
```
GET /presigned-url?key=images/photo.jpg
```

### 指定内容长度和有效期
```
GET /presigned-url?key=documents/report.pdf&content-length=1048576&expire=7200
```

### 启用即传即收和强制下载
```
GET /presigned-url?key=videos/movie.mp4&no-wait=5&force-download=true
```

### 限制下载次数和速率
```
GET /presigned-url?key=archives/backup.zip&max-requests=3&limit-rate=1048576
```
