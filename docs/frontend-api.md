# 前端接口文档

## 文档范围

本文档描述当前仓库已经实际暴露的全部前端对接接口，覆盖以下能力：

- 用户注册
- 用户登录
- 图片文件上传
- 图片 URL 上传
- 图片业务详情
- 图片管理原始详情
- 图片业务分页列表
- 图片管理原始分页列表
- 首页轮播图

当前后端返回的是图片元数据和图片地址，不会通过接口直接返回图片二进制。前端应当直接使用 `url` 或 `thumbnailUrl` 去加载图片资源。

## 基础信息

- Base URL：`http://localhost:8888/api`
- Content-Type：
  - JSON 接口：`application/json`
  - 文件上传接口：`multipart/form-data`
- 统一响应包裹：

```json
{
  "code": 200,
  "message": "成功",
  "data": {}
}
```

响应规则：

- `code`：当前实现里业务码与 HTTP 状态族语义保持一致
- `message`：用户可读信息，`4xx` 场景一般可直接展示
- `data`：成功时返回，失败时通常省略
- 前端成功判断条件：`body.code === 200`

## 通用错误码

当前仓库里常见错误码如下：

- `200`：成功
- `400`：请求参数错误、文件非法、URL 非法、分页参数非法
- `401`：未登录、JWT 无效、密码错误
- `403`：已登录但无权限，例如查看未过审图片、访问管理员接口
- `404`：资源不存在
- `409`：资源冲突，当前主要用于注册邮箱重复
- `500`：服务端异常，例如数据库失败、生成 Token 失败、COS 上传失败、文件处理失败

## 鉴权

受保护接口需要携带：

```http
Authorization: Bearer <jwt-token>
```

Token 来源：

- 前端通过 `POST /user/login` 获取
- 当前 token 中包含的主要 claim：
  - `userId`
  - `iat`
  - `exp`

## 接口总览

| 方法 | 路径 | 用途 | 鉴权 |
| --- | --- | --- | --- |
| POST | `/user/register` | 用户注册 | 无 |
| POST | `/user/login` | 用户登录 | 无 |
| POST | `/picture/upload` | 图片文件上传 | 需要 |
| POST | `/picture/upload/url` | 图片 URL 上传 | 需要 |
| GET | `/picture/get/vo?id=...` | 图片业务详情 | 审核通过可公开访问 |
| GET | `/picture/:id` | 图片业务详情别名 | 审核通过可公开访问 |
| GET | `/picture/get?id=...` | 图片管理原始详情 | 仅管理员 |
| POST | `/picture/list/page/vo` | 图片业务分页列表 | 无 |
| POST | `/picture/list/page` | 图片管理原始分页列表 | 仅管理员 |
| GET | `/picture/home/carousel` | 首页轮播图 | 无 |

## 用户接口

### POST /user/register

- 完整路径：`POST /api/user/register`
- Content-Type：`application/json`
- 鉴权：无

请求体：

```json
{
  "userEmail": "new@example.com",
  "userPassword": "secret123",
  "userCheckPassword": "secret123"
}
```

请求字段说明：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `userEmail` | string | 是 | 用户唯一邮箱 |
| `userPassword` | string | 是 | 前端传明文密码 |
| `userCheckPassword` | string | 是 | 必须与 `userPassword` 一致 |

成功响应：

```json
{
  "code": 200,
  "message": "成功",
  "data": {
    "id": 11
  }
}
```

常见失败场景：

- `400`：JSON 格式错误、缺少必填字段、两次密码不一致
- `409`：邮箱已存在
- `500`：数据库查询失败、密码加密失败、插入失败

### POST /user/login

- 完整路径：`POST /api/user/login`
- Content-Type：`application/json`
- 鉴权：无

请求体：

```json
{
  "userEmail": "user@example.com",
  "userPassword": "secret123"
}
```

请求字段说明：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `userEmail` | string | 是 | 登录邮箱 |
| `userPassword` | string | 是 | 前端传明文密码 |

成功响应：

```json
{
  "code": 200,
  "message": "成功",
  "data": {
    "token": "jwt-token",
    "id": 1,
    "userAccount": "user@example.com",
    "userName": "用户",
    "userAvatar": "",
    "userProfile": "",
    "userRole": "user",
    "createTime": "2026-04-18 20:00:00",
    "updateTime": "2026-04-18 20:00:00"
  }
}
```

响应字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `token` | string | 后续受保护接口的 Bearer Token |
| `id` | int64 | 当前用户 ID |
| `userAccount` | string | 当前实现里等同于 `userEmail` |
| `userName` | string | 显示名称 |
| `userAvatar` | string | 头像地址 |
| `userProfile` | string | 个人简介 |
| `userRole` | string | 当前角色，可能为 `user` 或 `admin` |
| `createTime` | string | 格式：`yyyy-MM-dd HH:mm:ss` |
| `updateTime` | string | 格式：`yyyy-MM-dd HH:mm:ss` |

常见失败场景：

- `400`：JSON 格式错误、缺少必填字段
- `401`：密码错误
- `404`：账号不存在
- `500`：数据库查询失败、Token 生成失败

## 图片数据结构

### 业务图片返回结构

业务接口会返回创建人信息 `user`，用于前端直接展示。

```json
{
  "id": 101,
  "url": "https://picture-storage-1325426290.cos.ap-guangzhou.myqcloud.com/public/1/demo.jpg?imageMogr2/format/webp",
  "name": "cover",
  "introduction": "summer trip",
  "category": "travel",
  "tags": ["travel", "sea"],
  "picSize": 123456,
  "picWidth": 1920,
  "picHeight": 1080,
  "picScale": 1.7777777778,
  "picFormat": "jpg",
  "userId": 1,
  "user": {
    "id": 1,
    "userName": "用户",
    "userAvatar": "",
    "userProfile": "",
    "userRole": "user"
  },
  "createTime": "2026-04-19 20:30:00",
  "editTime": "2026-04-19 20:30:00",
  "updateTime": "2026-04-19 20:30:00",
  "reviewStatus": 1,
  "reviewMessage": "审核通过",
  "reviewerId": 9,
  "reviewTime": "2026-04-19 20:35:00",
  "thumbnailUrl": "https://picture-storage-1325426290.cos.ap-guangzhou.myqcloud.com/public/1/demo.jpg?imageMogr2/thumbnail/128x128%3E/format/webp",
  "picColor": "#AABBCC",
  "viewCount": 12,
  "likeCount": 3
}
```

字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 图片 ID |
| `url` | string | 主图访问地址 |
| `name` | string | 图片名称 |
| `introduction` | string | 简介，可为空 |
| `category` | string | 分类，可为空 |
| `tags` | string[] | 标签数组，可为空 |
| `picSize` | int64 | 文件大小，单位字节 |
| `picWidth` | int64 | 宽度 |
| `picHeight` | int64 | 高度 |
| `picScale` | float64 | 宽高比 |
| `picFormat` | string | 当前可能为 `jpg`、`jpeg`、`png`、`webp` |
| `userId` | int64 | 创建人 ID |
| `user` | object | 创建人摘要，仅业务接口返回 |
| `createTime` | string | 创建时间，格式 `yyyy-MM-dd HH:mm:ss` |
| `editTime` | string | 编辑时间，格式 `yyyy-MM-dd HH:mm:ss` |
| `updateTime` | string | 更新时间，格式 `yyyy-MM-dd HH:mm:ss` |
| `reviewStatus` | int64 | 当前实现里 `0 = 待审核`，`1 = 已通过` |
| `reviewMessage` | string | 审核说明 |
| `reviewerId` | int64 | 审核人 ID |
| `reviewTime` | string | 审核时间 |
| `thumbnailUrl` | string | 缩略图/压缩图地址 |
| `picColor` | string | 主色调，例如 `#AABBCC` |
| `viewCount` | int64 | 浏览次数 |
| `likeCount` | int64 | 点赞次数 |

### 管理原始图片返回结构

管理员原始接口当前仍复用同一套图片字段，但不会补 `user` 创建人摘要。

## 图片上传接口

### POST /picture/upload

- 完整路径：`POST /api/picture/upload`
- Content-Type：`multipart/form-data`
- 鉴权：需要

表单字段：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `file` | file | 是 | 图片文件 |
| `id` | string/int | 否 | 已存在图片 ID，用于重传/更新 |
| `picName` | string | 否 | 自定义图片名称 |
| `introduction` | string | 否 | 简介 |
| `category` | string | 否 | 分类 |
| `tags` | string | 否 | 支持 JSON 数组字符串，如 `["a","b"]`，或逗号分隔字符串，如 `a,b` |

后端行为：

- 必须先登录
- 文件不能为空
- 允许格式：`jpg`、`jpeg`、`png`、`webp`
- 文件大小上限：`30MB`
- 后端会先写入临时文件，再提取元数据，再上传到 COS，最后入库
- 对象存储路径格式：`public/{userId}/{yyyy-MM-dd}_{16位随机hex}.{ext}`
- `url` 指向处理后的 webp 展示地址
- `thumbnailUrl` 指向缩略图/压缩图地址
- 如果 `picName` 为空，则取原始文件名去掉扩展名
- 如果传了 `id`：
  - 会先查询原记录
  - 只有图片所有者或管理员能更新
  - 原有 `createTime`、`userId`、`viewCount`、`likeCount` 会保留
- 审核逻辑：
  - 管理员上传：自动通过
  - 普通用户上传：`reviewStatus = 0`

成功响应：

- 返回业务图片结构，包含 `user`

常见失败场景：

- `400`：缺少文件、`id` 非法、`tags` 非法、格式不支持、文件过大
- `401`：未登录或 JWT 无效
- `403`：尝试更新别人的图片
- `404`：目标图片不存在
- `500`：临时文件失败、元数据提取失败、数据库失败、COS 失败

### POST /picture/upload/url

- 完整路径：`POST /api/picture/upload/url`
- Content-Type：`application/json`
- 鉴权：需要

请求体：

```json
{
  "id": 101,
  "fileUrl": "https://example.com/demo.webp",
  "picName": "remote-demo",
  "introduction": "remote upload",
  "category": "demo",
  "tags": ["remote", "sample"]
}
```

请求字段说明：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 否 | 已存在图片 ID，用于重传/更新 |
| `fileUrl` | string | 是 | 远程图片 URL |
| `picName` | string | 否 | 自定义图片名称 |
| `introduction` | string | 否 | 简介 |
| `category` | string | 否 | 分类 |
| `tags` | string[] | 否 | 标签数组 |

后端行为：

- 必须先登录
- `fileUrl` 必须是合法的 `http` 或 `https`
- 后端会先尝试发一次 `HEAD`
- 如果 `HEAD` 返回 `200`，会校验：
  - `Content-Type` 以 `image/` 开头
  - `Content-Length <= 10MB`
- 如果目标站点不支持 `HEAD`，后端仍会继续走 `GET`
- 后端会把远程图片下载到临时文件，再走与文件上传相同的后续流程
- 远程图片大小上限：`10MB`
- 最终允许的图片格式仍然只有 `jpg`、`jpeg`、`png`、`webp`

成功响应：

- 返回业务图片结构，包含 `user`

常见失败场景：

- `400`：`fileUrl` 为空、URL 非法、协议不支持、远程文件不是图片、远程文件过大、下载失败
- `401`：未登录或 JWT 无效
- `403`：尝试更新别人的图片
- `404`：目标图片不存在
- `500`：临时文件失败、元数据提取失败、数据库失败、COS 失败

## 图片详情接口

### GET /picture/get/vo

- 完整路径：`GET /api/picture/get/vo?id=<pictureId>`
- 鉴权：
  - 审核通过图片可匿名访问
  - 未通过图片需要本人或管理员

查询参数：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 图片 ID |

后端行为：

- 查询有效图片记录
- 如果 `reviewStatus = 1`，该图片公开可见
- 如果 `reviewStatus != 1`，只有本人或管理员可见
- 每次成功访问会把 `viewCount + 1`
- 返回里会补创建人摘要 `data.user`

成功响应：

- 返回业务图片结构

常见失败场景：

- `400`：`id` 缺失或非法
- `403`：图片未公开，且调用者不是本人/管理员
- `404`：图片不存在或已删除
- `500`：数据库失败、浏览数更新失败

### GET /picture/:id

- 完整路径：`GET /api/picture/:id`
- 行为：上面业务详情接口的路径别名
- 可见性规则和 `viewCount` 副作用完全一致

### GET /picture/get

- 完整路径：`GET /api/picture/get?id=<pictureId>`
- 鉴权：仅管理员

后端行为：

- 查询有效图片记录
- 不走业务可见性判断
- 不增加 `viewCount`
- 不补创建人摘要 `user`

成功响应：

- 返回管理员原始图片结构

常见失败场景：

- `400`：`id` 缺失或非法
- `401`：未登录或 JWT 无效
- `403`：非管理员
- `404`：图片不存在或已删除
- `500`：数据库失败

## 图片分页列表接口

### POST /picture/list/page/vo

- 完整路径：`POST /api/picture/list/page/vo`
- Content-Type：`application/json`
- 鉴权：无

请求体示例：

```json
{
  "pageNum": 1,
  "pageSize": 12,
  "category": "travel",
  "tags": ["sea"],
  "searchText": "sunset"
}
```

业务规则：

- 默认 `pageNum = 1`
- 默认 `pageSize = 10`
- `pageSize` 最大只能是 `20`
- 业务列表会强制追加：
  - `isDelete = 0`
  - `reviewStatus = 1`
- 返回结果会给每条图片补创建人摘要 `user`

支持的请求字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `pageNum` | int64 | 可选 |
| `pageSize` | int64 | 可选，最大 `20` |
| `id` | int64 | 精确匹配 |
| `name` | string | 模糊匹配 |
| `introduction` | string | 模糊匹配 |
| `category` | string | 精确匹配 |
| `tags` | string[] | 每个标签都会生成一个 `tags like` 过滤 |
| `picSize` | int64 | 精确匹配 |
| `picWidth` | int64 | 精确匹配 |
| `picHeight` | int64 | 精确匹配 |
| `picScale` | float64 | 精确匹配 |
| `picFormat` | string | 精确匹配 |
| `userId` | int64 | 精确匹配 |
| `searchText` | string | 同时对 `name` 和 `introduction` 做模糊匹配 |
| `editTimeStart` | string | 支持 `yyyy-MM-dd` 或 `yyyy-MM-dd HH:mm:ss` |
| `editTimeEnd` | string | 支持 `yyyy-MM-dd` 或 `yyyy-MM-dd HH:mm:ss` |

即使传了也会被业务列表忽略的字段：

- `reviewStatus`
- `reviewMessage`
- `reviewerId`

成功响应：

```json
{
  "code": 200,
  "message": "成功",
  "data": {
    "pageNum": 1,
    "pageSize": 12,
    "total": 2,
    "list": [
      {
        "id": 101,
        "name": "cover"
      }
    ]
  }
}
```

### POST /picture/list/page

- 完整路径：`POST /api/picture/list/page`
- Content-Type：`application/json`
- 鉴权：仅管理员

规则：

- 默认 `pageNum = 1`
- 默认 `pageSize = 10`
- `pageSize` 最大只能是 `20`
- 只会固定过滤 `isDelete = 0`
- 不会强制 `reviewStatus = 1`
- 不会补 `user` 创建人摘要

支持字段：

- 业务分页列表的全部字段
- 以及额外支持：
  - `reviewStatus`
  - `reviewMessage`
  - `reviewerId`

两个分页接口的常见失败场景：

- `400`：时间格式非法、`pageSize > 20`
- `401`：管理员分页接口未登录或 JWT 无效
- `403`：管理员分页接口调用者不是管理员
- `500`：总数查询失败、分页查询失败

## 首页轮播图接口

### GET /picture/home/carousel

- 完整路径：`GET /api/picture/home/carousel`
- 鉴权：无

后端行为：

- 只查询公开图片
- 等价过滤条件：
  - `isDelete = 0`
  - `reviewStatus = 1`
- 排序规则：`viewCount desc, id desc`
- 固定只取 `6` 条
- 每条记录都会补 `user` 创建人摘要

成功响应：

```json
{
  "code": 200,
  "message": "成功",
  "data": {
    "list": [
      {
        "id": 101,
        "name": "cover"
      }
    ]
  }
}
```

## 前端对接注意事项

- 业务图片接口用于正常产品页面，管理原始接口用于后台审核或管理控制台。
- `GET /picture/get/vo` 和 `GET /picture/:id` 都会增加浏览数，前端不要对这两个接口做高频轮询或无意义预取。
- 当前仓库还没有 `spaceId`，所以这里实现的都是公共图库逻辑，不包含空间图库。
- 文件上传时，`tags` 必须作为一个表单字段字符串传递，不能用多个同名 multipart key 替代。
- 业务分页页大小不要请求超过 `20`。
- 上传接口真正联调前，需要先把 [usercenter.yaml](D:/Development%20Tools/Project/Go-Projects/photo-album/etc/usercenter.yaml) 里的 COS 占位密钥替换成你本地有效配置。
