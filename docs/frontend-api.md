# 前端接口文档

## 文档范围

本文档描述当前仓库已经实际暴露的前端接口，覆盖以下能力：

- 用户注册
- 用户登录
- 公开用户详情
- 当前登录用户详情
- 当前登录用户资料更新
- 图片文件上传
- 图片 URL 上传
- 图片业务详情
- 图片管理原始详情
- 图片删除
- 图片管理员审核
- 图片业务分页列表
- 当前登录用户作品分页列表
- 图片管理原始分页列表
- 首页轮播图

当前后端返回的是图片元数据和图片地址，不会通过接口直接返回图片二进制。前端应直接使用 `url` 或 `thumbnailUrl` 加载图片资源，其中 `thumbnailUrl` 当前实际语义为压缩图地址。

## 基础信息

- Base URL：`http://localhost:8888/api`
- JSON 接口 Content-Type：`application/json`
- 文件上传接口 Content-Type：`multipart/form-data`
- 鉴权方式：`Authorization: Bearer <jwt-token>`

统一响应包裹：

```json
{
  "code": 200,
  "message": "成功",
  "data": {}
}
```

响应规则：

- `code`：当前实现里业务码与 HTTP 状态语义保持一致
- `message`：用户可读信息，`4xx` 场景通常可直接展示
- `data`：成功时返回，失败时通常省略
- 前端成功判断条件：`body.code === 200`

## 雪花 ID 约定

后端当前所有对外 `id / userId / reviewerId` 都按字符串返回，避免前端 JavaScript 精度丢失。

前端接入规则：

- 所有 ID 在前端状态、路由参数、请求参数里都按 `string` 处理
- JSON 请求体里的 ID，后端同时兼容：
  - `"1921565896585154562"`
  - `1921565896585154562`
- 仍然强烈建议前端统一传字符串，避免浏览器在发送前先把大整数转成不安全的 `number`
- query/path/form-data 里的 ID 一律按字符串传

示例：

```json
{
  "id": "1921565896585154562"
}
```

## 通用错误码

- `200`：成功
- `400`：请求参数错误、文件非法、URL 非法、分页参数非法、时间格式错误
- `401`：未登录、JWT 无效、密码错误
- `403`：已登录但无权限，例如查看未过审图片、访问管理员接口、删除他人图片
- `404`：资源不存在
- `409`：资源冲突，当前主要用于注册邮箱重复
- `500`：服务端异常，例如数据库失败、生成 Token 失败、COS 上传失败、文件处理失败

## 鉴权

受保护接口需要携带：

```http
Authorization: Bearer <jwt-token>
```

Token 来源：

- 前端通过 `POST /api/user/login` 获取
- 当前 token 里主要包含：
  - `userId`：字符串
  - `iat`
  - `exp`

建议：

- 前端如果需要从 token 中读取 `userId`，也按字符串处理
- 不要把 token 里的 `userId` 当作 `number`

## 接口总览

| 方法 | 路径 | 用途 | 鉴权 |
| --- | --- | --- | --- |
| POST | `/user/register` | 用户注册 | 无 |
| POST | `/user/login` | 用户登录 | 无 |
| GET | `/user/get/vo?id=...` | 公开用户详情 | 无 |
| GET | `/user/my` | 当前登录用户详情 | 需要 |
| POST | `/user/my` | 更新当前登录用户资料 | 需要 |
| PATCH | `/user/my` | 更新当前登录用户资料 | 需要 |
| POST | `/picture/upload` | 图片文件上传 | 需要 |
| POST | `/picture/upload/url` | 图片 URL 上传 | 需要 |
| GET | `/picture/get/vo?id=...` | 图片业务详情 | 审核通过可公开访问 |
| GET | `/picture/:id` | 图片业务详情别名 | 审核通过可公开访问 |
| GET | `/picture/get?id=...` | 图片管理原始详情 | 仅管理员 |
| POST | `/picture/delete` | 图片删除 | 需要 |
| POST | `/picture/review` | 图片管理员审核 | 仅管理员 |
| POST | `/picture/list/page/vo` | 图片业务分页列表 | 无 |
| POST | `/picture/my/list/page` | 当前登录用户作品分页列表 | 需要 |
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
| `userPassword` | string | 是 | 明文密码 |
| `userCheckPassword` | string | 是 | 必须与 `userPassword` 一致 |

成功响应：

```json
{
  "code": 200,
  "message": "成功",
  "data": {
    "id": "1921565896585154562"
  }
}
```

常见失败场景：

- `400`：JSON 格式错误、缺少必填字段、两次密码不一致
- `409`：邮箱已存在
- `500`：数据库失败、密码加密失败、插入失败

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

成功响应：

```json
{
  "code": 200,
  "message": "成功",
  "data": {
    "token": "jwt-token",
    "id": "1921565896585154562",
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
| `id` | string | 当前用户 ID，雪花 ID 字符串 |
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

### UserProfileResponse

用户详情相关接口统一返回以下结构：

```json
{
  "id": "1921565896585154562",
  "userName": "用户",
  "userAvatar": "https://example.com/avatar.png",
  "userProfile": "热爱摄影",
  "userRole": "user",
  "createTime": "2026-04-18 20:00:00",
  "updateTime": "2026-04-20 11:00:00",
  "pictureCount": 8,
  "approvedPictureCount": 5,
  "pendingPictureCount": 2,
  "rejectedPictureCount": 1
}
```

字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | string | 用户 ID，雪花 ID 字符串 |
| `userName` | string | 用户昵称 |
| `userAvatar` | string | 头像地址 |
| `userProfile` | string | 用户简介 |
| `userRole` | string | 用户角色，当前可能为 `user` 或 `admin` |
| `createTime` | string | 创建时间，格式 `yyyy-MM-dd HH:mm:ss` |
| `updateTime` | string | 更新时间，格式 `yyyy-MM-dd HH:mm:ss` |
| `pictureCount` | number | 当前用户未删除作品总数 |
| `approvedPictureCount` | number | 当前用户已通过作品数 |
| `pendingPictureCount` | number | 当前用户待审核作品数 |
| `rejectedPictureCount` | number | 当前用户已拒绝作品数 |

### GET /user/get/vo?id=<userId>

- 完整路径：`GET /api/user/get/vo?id=<userId>`
- 鉴权：无

查询参数：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | string | 是 | 用户 ID |

后端行为：

- 只查询未删除用户
- 返回 `UserProfileResponse`
- 作品统计基于 `pictures.isDelete = 0`

常见失败场景：

- `400`：`id` 非法
- `404`：用户不存在
- `500`：数据库失败

### GET /user/my

- 完整路径：`GET /api/user/my`
- 鉴权：需要

后端行为：

- 根据 Bearer Token 识别当前登录用户
- 返回 `UserProfileResponse`
- 作品统计基于当前登录用户全部未删除作品

常见失败场景：

- `401`：未登录、Token 无效、登录用户不存在
- `500`：数据库失败

### POST /user/my

- 完整路径：`POST /api/user/my`
- Content-Type：`application/json`
- 鉴权：需要

### PATCH /user/my

- 完整路径：`PATCH /api/user/my`
- Content-Type：`application/json`
- 鉴权：需要

说明：

- `POST` 和 `PATCH` 当前行为完全一致，前端任选其一
- 只允许更新：
  - `userName`
  - `userAvatar`
  - `userProfile`
- 至少传一个字段
- `userName.trim()` 不能为空
- 长度限制：
  - `userName` 最长 `256`
  - `userAvatar` 最长 `1024`
  - `userProfile` 最长 `512`

请求体示例：

```json
{
  "userName": "旅行摄影师",
  "userAvatar": "https://example.com/new-avatar.png",
  "userProfile": "记录每一次出发"
}
```

成功响应：

- 返回更新后的 `UserProfileResponse`

常见失败场景：

- `400`：请求体为空、没有可更新字段、`userName` 为空、字段长度超限
- `401`：未登录、Token 无效、登录用户不存在
- `500`：数据库失败

## 图片数据结构

### UserSummary

```json
{
  "id": "1921565896585154562",
  "userName": "用户",
  "userAvatar": "",
  "userProfile": "",
  "userRole": "user"
}
```

### PictureResponse

说明：字段名仍然叫 `thumbnailUrl`，但当前后端为了兼容旧前端，没有改字段名；它实际表示“压缩图地址”，不是 `128x128` 缩略图。

```json
{
  "id": "1921565896585154562",
  "url": "https://picture-storage-1325426290.cos.ap-guangzhou.myqcloud.com/public/1/demo.jpg",
  "name": "cover",
  "introduction": "summer trip",
  "category": "travel",
  "tags": ["travel", "sea"],
  "picSize": 12345678,
  "picWidth": 1920,
  "picHeight": 1080,
  "picScale": 1.7777777778,
  "picFormat": "jpg",
  "userId": "1921565896585154562",
  "user": {
    "id": "1921565896585154562",
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
  "reviewerId": "1921565896585154563",
  "reviewTime": "2026-04-19 20:35:00",
  "thumbnailUrl": "https://picture-storage-1325426290.cos.ap-guangzhou.myqcloud.com/public/1/demo.jpg?imageMogr2/format/webp",
  "picColor": "#AABBCC",
  "viewCount": 12,
  "likeCount": 3
}
```

字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | string | 图片 ID，雪花 ID 字符串 |
| `url` | string | 主图访问地址 |
| `name` | string | 图片名称 |
| `introduction` | string | 简介，可为空 |
| `category` | string | 分类，可为空 |
| `tags` | string[] | 标签数组，可为空 |
| `picSize` | number | 文件大小，单位字节 |
| `picWidth` | number | 宽度 |
| `picHeight` | number | 高度 |
| `picScale` | number | 宽高比 |
| `picFormat` | string | 当前支持 `jpg`、`jpeg`、`png`、`webp` |
| `userId` | string | 创建人 ID |
| `user` | object | 创建人摘要，业务接口通常会返回，管理员原始接口可能为空 |
| `createTime` | string | 创建时间，格式 `yyyy-MM-dd HH:mm:ss` |
| `editTime` | string | 编辑时间，格式 `yyyy-MM-dd HH:mm:ss` |
| `updateTime` | string | 更新时间，格式 `yyyy-MM-dd HH:mm:ss` |
| `reviewStatus` | number | `0 = 待审核`，`1 = 已通过`，`2 = 已拒绝` |
| `reviewMessage` | string | 审核说明 |
| `reviewerId` | string | 审核人 ID，可能为空字符串 |
| `reviewTime` | string | 审核时间 |
| `thumbnailUrl` | string | 压缩图地址；图片小于等于 `2MB` 时等于原图，大于 `2MB` 时返回压缩图地址 |
| `picColor` | string | 主色调，例如 `#AABBCC` |
| `viewCount` | number | 浏览次数 |
| `likeCount` | number | 点赞次数 |

### PicturePageResponse

```json
{
  "pageNum": 1,
  "pageSize": 10,
  "total": 23,
  "list": []
}
```

### PictureDeleteResponse

```json
{
  "id": "1921565896585154562"
}
```

## 图片上传接口

### POST /picture/upload

- 完整路径：`POST /api/picture/upload`
- Content-Type：`multipart/form-data`
- 鉴权：需要

表单字段：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `file` | file | 是 | 图片文件 |
| `id` | string | 否 | 已存在图片 ID，用于更新图片 |
| `picName` | string | 否 | 自定义图片名称 |
| `introduction` | string | 否 | 简介 |
| `category` | string | 否 | 分类 |
| `tags` | string | 否 | 支持 JSON 数组字符串，如 `["a","b"]`，或逗号分隔字符串，如 `a,b` |

后端行为：

- 必须先登录
- 文件不能为空
- 允许格式：`jpg`、`jpeg`、`png`、`webp`
- 文件大小上限：`30MB`
- 后端会先保存临时文件、提取元数据、上传到 COS、再写数据库
- 对象存储路径格式：`public/{userId}/{yyyy-MM-dd}_{16位随机hex}.{ext}`
- `url` 指向原图
- `thumbnailUrl` 在图片大小大于 `2MB` 时使用压缩图，否则等于原图
- 如果 `picName` 为空，则取原始文件名去掉扩展名
- 如果传了 `id`：
  - 会先查原记录
  - 只有图片所有者或管理员能更新
  - 原有 `createTime`、`userId`、`viewCount`、`likeCount` 会保留
- 审核逻辑：
  - 管理员上传：自动通过
  - 普通用户上传：进入待审核

成功响应：

- 返回 `PictureResponse`

常见失败场景：

- `400`：缺少文件、`id` 非法、`tags` 非法、格式不支持、文件过大
- `401`：未登录或 JWT 无效
- `403`：尝试更新他人的图片
- `404`：目标图片不存在
- `500`：临时文件失败、元数据提取失败、数据库失败、COS 上传失败

### POST /picture/upload/url

- 完整路径：`POST /api/picture/upload/url`
- Content-Type：`application/json`
- 鉴权：需要

请求体：

```json
{
  "id": "1921565896585154562",
  "fileUrl": "https://example.com/demo.webp",
  "picName": "remote-demo",
  "introduction": "remote upload",
  "category": "demo",
  "tags": ["remote", "sample"]
}
```

说明：

- `id` 既支持字符串，也兼容数字；前端建议始终传字符串
- `fileUrl` 必须是合法 `http/https` 地址
- 后端会先尝试 `HEAD`，再走 `GET`
- 如果 `HEAD` 返回正常，会校验：
  - `Content-Type` 必须以 `image/` 开头
  - `Content-Length <= 10MB`
- 远程图片大小上限：`10MB`

成功响应：

- 返回 `PictureResponse`

常见失败场景：

- `400`：`fileUrl` 为空、URL 非法、协议不支持、远程文件不是图片、远程文件过大、下载失败
- `401`：未登录或 JWT 无效
- `403`：尝试更新他人的图片
- `404`：目标图片不存在
- `500`：临时文件失败、元数据提取失败、数据库失败、COS 上传失败

## 图片详情接口

### GET /picture/get/vo?id=<pictureId>

- 完整路径：`GET /api/picture/get/vo?id=<pictureId>`
- 鉴权：
  - 审核通过的图片可匿名访问
  - 未通过图片需要本人或管理员

查询参数：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | string | 是 | 图片 ID |

后端行为：

- 查询未删除图片
- 审核通过图片公开可见
- 未通过图片仅本人或管理员可见
- 成功访问时会把 `viewCount + 1`
- 返回里会补 `user` 创建人摘要

### GET /picture/:id

这是 `/picture/get/vo` 的路径版别名，行为与返回结构一致，也会增加浏览数。

### GET /picture/get?id=<pictureId>

- 完整路径：`GET /api/picture/get?id=<pictureId>`
- 鉴权：仅管理员

说明：

- 返回结构仍是 `PictureResponse`
- 当前实现不会补 `user` 创建人摘要
- 不增加浏览数

## 图片删除接口

### POST /picture/delete

- 完整路径：`POST /api/picture/delete`
- Content-Type：`application/json`
- 鉴权：需要

请求体：

```json
{
  "id": "1921565896585154562"
}
```

说明：

- `id` 支持字符串，也兼容数字；前端建议始终传字符串
- 只有图片所有者或管理员可以删除
- 删除采用逻辑删除，接口成功后图片不会再出现在查询结果中
- 当图片地址属于当前 COS Host 且服务端本地已配置 COS 密钥时，会同步删除远端对象

成功响应：

```json
{
  "code": 200,
  "message": "成功",
  "data": {
    "id": "1921565896585154562"
  }
}
```

## 图片审核接口

### POST /picture/review

- 完整路径：`POST /api/picture/review`
- Content-Type：`application/json`
- 鉴权：仅管理员

请求体：

```json
{
  "id": "1921565896585154562",
  "reviewStatus": 2,
  "reviewMessage": "图片内容不符合要求"
}
```

规则：

- `id` 支持字符串，也兼容数字；前端建议始终传字符串
- `reviewStatus` 只能是：
  - `1`：通过
  - `2`：拒绝
- 当 `reviewStatus = 2` 时，`reviewMessage.trim()` 不能为空
- 返回管理员原始图片结构，不补 `user`
- 不增加浏览数

常见失败场景：

- `400`：`id` 非法、`reviewStatus` 非法、拒绝但 `reviewMessage` 为空
- `401`：未登录或 JWT 无效
- `403`：非管理员
- `404`：图片不存在
- `500`：数据库失败

## 图片分页列表接口

### POST /picture/list/page/vo

- 完整路径：`POST /api/picture/list/page/vo`
- Content-Type：`application/json`
- 鉴权：无

业务列表只返回：

- `isDelete = 0`
- `reviewStatus = 1`

请求体示例：

```json
{
  "pageNum": 1,
  "pageSize": 10,
  "id": "1921565896585154562",
  "name": "cover",
  "introduction": "summer",
  "category": "travel",
  "tags": ["sea"],
  "userId": "1921565896585154562",
  "searchText": "trip"
}
```

支持的主要过滤字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `pageNum` | number | 默认 `1` |
| `pageSize` | number | 默认 `10`，最大 `20` |
| `id` | string | 图片 ID |
| `name` | string | 图片名称，模糊匹配 |
| `introduction` | string | 简介，模糊匹配 |
| `category` | string | 分类，精确匹配 |
| `tags` | string[] | 每个 tag 都按 `like` 过滤 |
| `picSize` | number | 图片大小 |
| `picWidth` | number | 宽度 |
| `picHeight` | number | 高度 |
| `picScale` | number | 宽高比 |
| `picFormat` | string | 图片格式 |
| `userId` | string | 创建人 ID |
| `editTimeStart` | string | 起始时间，支持 `yyyy-MM-dd` 或 `yyyy-MM-dd HH:mm:ss` |
| `editTimeEnd` | string | 结束时间，支持 `yyyy-MM-dd` 或 `yyyy-MM-dd HH:mm:ss` |
| `searchText` | string | 同时模糊匹配 `name/introduction` |

成功响应：

- 返回 `PicturePageResponse`
- 业务列表会补 `list[].user`

### POST /picture/my/list/page

- 完整路径：`POST /api/picture/my/list/page`
- Content-Type：`application/json`
- 鉴权：需要

说明：

- 返回当前登录用户自己的作品分页
- 后端会忽略前端传入的 `userId`，强制改成当前登录用户 ID
- 与公开业务分页相比：
  - 会查询当前用户全部未删除作品
  - 支持按 `reviewStatus` 过滤自己的待审核、已通过、已拒绝作品
  - 返回里会补 `list[].user`

请求体示例：

```json
{
  "pageNum": 1,
  "pageSize": 10,
  "reviewStatus": 0,
  "searchText": "旅行"
}
```

支持的主要过滤字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `pageNum` | number | 默认 `1` |
| `pageSize` | number | 默认 `10`，最大 `20` |
| `id` | string | 图片 ID |
| `name` | string | 图片名称，模糊匹配 |
| `introduction` | string | 简介，模糊匹配 |
| `category` | string | 分类，精确匹配 |
| `tags` | string[] | 每个 tag 都按 `like` 过滤 |
| `picSize` | number | 图片大小 |
| `picWidth` | number | 宽度 |
| `picHeight` | number | 高度 |
| `picScale` | number | 宽高比 |
| `picFormat` | string | 图片格式 |
| `reviewStatus` | number | `0 = 待审核`，`1 = 已通过`，`2 = 已拒绝` |
| `reviewMessage` | string | 审核说明，模糊匹配 |
| `reviewerId` | string | 审核人 ID |
| `editTimeStart` | string | 起始时间，支持 `yyyy-MM-dd` 或 `yyyy-MM-dd HH:mm:ss` |
| `editTimeEnd` | string | 结束时间，支持 `yyyy-MM-dd` 或 `yyyy-MM-dd HH:mm:ss` |
| `searchText` | string | 同时模糊匹配 `name/introduction` |

成功响应：

- 返回 `PicturePageResponse`

### POST /picture/list/page

- 完整路径：`POST /api/picture/list/page`
- Content-Type：`application/json`
- 鉴权：仅管理员

与业务分页相比，管理员分页还支持：

- `reviewStatus`
- `reviewMessage`
- `reviewerId`

且不会补 `list[].user`。

## 首页轮播接口

### GET /picture/home/carousel

- 完整路径：`GET /api/picture/home/carousel`
- 鉴权：无

后端行为：

- 只返回已通过、未删除图片
- 按 `viewCount desc, id desc` 排序
- 最多返回 `6` 条
- 返回里会补 `user`

成功响应：

```json
{
  "code": 200,
  "message": "成功",
  "data": {
    "list": []
  }
}
```

## TypeScript 类型建议

```ts
export type SnowflakeId = string;
export type ReviewStatus = 0 | 1 | 2;

export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export interface UserSummary {
  id: SnowflakeId;
  userName: string;
  userAvatar: string;
  userProfile: string;
  userRole: string;
}

export interface UserProfileResponse {
  id: SnowflakeId;
  userName: string;
  userAvatar: string;
  userProfile: string;
  userRole: string;
  createTime: string;
  updateTime: string;
  pictureCount: number;
  approvedPictureCount: number;
  pendingPictureCount: number;
  rejectedPictureCount: number;
}

export interface PictureResponse {
  id: SnowflakeId;
  url: string;
  name: string;
  introduction?: string;
  category?: string;
  tags?: string[];
  picSize?: number;
  picWidth?: number;
  picHeight?: number;
  picScale?: number;
  picFormat?: string;
  userId: SnowflakeId;
  user?: UserSummary;
  createTime: string;
  editTime: string;
  updateTime: string;
  reviewStatus: ReviewStatus;
  reviewMessage?: string;
  reviewerId?: SnowflakeId;
  reviewTime?: string;
  thumbnailUrl?: string;
  picColor?: string;
  viewCount: number;
  likeCount: number;
}

export interface PicturePageResponse {
  pageNum: number;
  pageSize: number;
  total: number;
  list: PictureResponse[];
}
```

## 前端接入建议

- 所有 ID 在前端统一使用 `string`
- 不要把雪花 ID 存成 JS `number`
- 请求里即使后端兼容数字，前端仍建议始终传字符串
- 用户详情页可配合：
  - `GET /user/get/vo`
  - `GET /user/my`
  - `POST/PATCH /user/my`
  - `POST /picture/my/list/page`
- `GET /picture/get/vo` 和 `GET /picture/:id` 会增加浏览数，不要做高频轮询
- 图片详情页如果只是后台审核使用，优先调用管理员接口 `/picture/get`
- 上传前请先确认 COS 配置已在服务端配置完成，否则上传会失败
