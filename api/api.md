# Conduit API 接口文档

## 1. 认证相关接口 (Auth)

### 1.1 用户注册 (Register)

**接口信息**
- 名称：用户注册
- 路径：`/users`
- 请求方法：POST
- 功能描述：创建新用户账号

**请求参数**
| 参数名 | 数据类型 | 是否必填 | 描述 |
|-------|---------|---------|------|
| user.email | string | 是 | 用户邮箱地址 |
| user.password | string | 是 | 用户密码 |
| user.username | string | 是 | 用户用户名 |

**请求头**
- Content-Type: application/json
- X-Requested-With: XMLHttpRequest

**请求示例**
```json
{
  "user": {
    "email": "user@example.com",
    "password": "password123",
    "username": "user123"
  }
}
```

**响应数据**
| 字段名 | 数据类型 | 描述 |
|-------|---------|------|
| user.email | string | 用户邮箱地址 |
| user.username | string | 用户用户名 |
| user.bio | string | 用户个人简介 |
| user.image | string | 用户头像URL |
| user.token | string | JWT认证令牌 |

**成功响应示例**
```json
{
  "user": {
    "email": "user@example.com",
    "username": "user123",
    "bio": "",
    "image": null,
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**错误码说明**
| 错误码 | 描述 |
|-------|------|
| 400 Bad Request | 邮箱已存在 |
| 400 Bad Request | 用户名已存在 |

### 1.2 用户登录 (Login)

**接口信息**
- 名称：用户登录
- 路径：`/users/login`
- 请求方法：POST
- 功能描述：用户登录并获取认证令牌

**请求参数**
| 参数名 | 数据类型 | 是否必填 | 描述 |
|-------|---------|---------|------|
| user.email | string | 是 | 用户邮箱地址 |
| user.password | string | 是 | 用户密码 |

**请求头**
- Content-Type: application/json
- X-Requested-With: XMLHttpRequest

**请求示例**
```json
{
  "user": {
    "email": "user@example.com",
    "password": "password123"
  }
}
```

**响应数据**
| 字段名 | 数据类型 | 描述 |
|-------|---------|------|
| user.email | string | 用户邮箱地址 |
| user.username | string | 用户用户名 |
| user.bio | string | 用户个人简介 |
| user.image | string | 用户头像URL |
| user.token | string | JWT认证令牌 |

**成功响应示例**
```json
{
  "user": {
    "email": "user@example.com",
    "username": "user123",
    "bio": "",
    "image": null,
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**错误码说明**
| 错误码 | 描述 |
|-------|------|
| 401 Unauthorized | 邮箱或密码错误 |

### 1.3 获取当前用户 (Current User)

**接口信息**
- 名称：获取当前用户
- 路径：`/user`
- 请求方法：GET
- 功能描述：获取当前登录用户的信息

**请求参数**
| 参数名 | 数据类型 | 是否必填 | 描述 |
|-------|---------|---------|------|
| N/A | N/A | N/A | N/A |

**请求头**
- Content-Type: application/json
- X-Requested-With: XMLHttpRequest
- Authorization: Token {{token}}

**请求示例**
```
GET /user
Authorization: Token eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应数据**
| 字段名 | 数据类型 | 描述 |
|-------|---------|------|
| user.email | string | 用户邮箱地址 |
| user.username | string | 用户用户名 |
| user.bio | string | 用户个人简介 |
| user.image | string | 用户头像URL |
| user.token | string | JWT认证令牌 |

**成功响应示例**
```json
{
  "user": {
    "email": "user@example.com",
    "username": "user123",
    "bio": "Software developer",
    "image": "https://example.com/avatar.jpg",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**错误码说明**
| 错误码 | 描述 |
|-------|------|
| 401 Unauthorized | 无效令牌 |

### 1.4 更新用户 (Update User)

**接口信息**
- 名称：更新用户
- 路径：`/user`
- 请求方法：PUT
- 功能描述：更新当前用户的信息

**请求参数**
| 参数名 | 数据类型 | 是否必填 | 描述 |
|-------|---------|---------|------|
| user.email | string | 否 | 用户邮箱地址 |
| user.username | string | 否 | 用户用户名 |
| user.password | string | 否 | 用户密码 |
| user.bio | string | 否 | 用户个人简介 |
| user.image | string | 否 | 用户头像URL |

**请求头**
- Content-Type: application/json
- X-Requested-With: XMLHttpRequest
- Authorization: Token {{token}}

**请求示例**
```json
{
  "user": {
    "email": "newemail@example.com",
    "bio": "Full-stack developer"
  }
}
```

**响应数据**
| 字段名 | 数据类型 | 描述 |
|-------|---------|------|
| user.email | string | 用户邮箱地址 |
| user.username | string | 用户用户名 |
| user.bio | string | 用户个人简介 |
| user.image | string | 用户头像URL |
| user.token | string | JWT认证令牌 |

**成功响应示例**
```json
{
  "user": {
    "email": "newemail@example.com",
    "username": "user123",
    "bio": "Full-stack developer",
    "image": "https://example.com/avatar.jpg",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**错误码说明**
| 错误码 | 描述 |
|-------|------|
| 401 Unauthorized | 无效令牌 |
| 400 Bad Request | 邮箱已存在 |

## 2. 文章相关接口 (Articles)

### 2.1 获取所有文章 (All Articles)

**接口信息**
- 名称：获取所有文章
- 路径：`/articles`
- 请求方法：GET
- 功能描述：获取文章列表，支持分页和筛选

**请求参数**
| 参数名 | 数据类型 | 是否必填 | 描述 |
|-------|---------|---------|------|
| page | integer | 否 | 页码，默认1 |
| limit | integer | 否 | 每页数量，默认20 |
| author | string | 否 | 按作者筛选 |
| tag | string | 否 | 按标签筛选 |
| favorited | string | 否 | 按收藏用户筛选 |

**请求头**
- Content-Type: application/json
- X-Requested-With: XMLHttpRequest

**请求示例**
```
GET /articles?page=1&limit=10&author=johnjacob
```

**响应数据**
| 字段名 | 数据类型 | 描述 |
|-------|---------|------|
| articles | array | 文章列表 |
| articlesCount | integer | 文章总数 |
| articles[].title | string | 文章标题 |
| articles[].slug | string | 文章唯一标识 |
| articles[].body | string | 文章内容 |
| articles[].createdAt | string | 创建时间（ISO 8601格式） |
| articles[].updatedAt | string | 更新时间（ISO 8601格式） |
| articles[].description | string | 文章描述 |
| articles[].tagList | array | 文章标签列表 |
| articles[].author | object | 作者信息 |
| articles[].author.username | string | 作者用户名 |
| articles[].author.bio | string | 作者简介 |
| articles[].author.image | string | 作者头像URL |
| articles[].author.following | boolean | 当前用户是否关注作者 |
| articles[].favorited | boolean | 当前用户是否收藏文章 |
| articles[].favoritesCount | integer | 文章收藏数 |

**成功响应示例**
```json
{
  "articles": [
    {
      "title": "How to train your dragon",
      "slug": "how-to-train-your-dragon",
      "body": "Very carefully.",
      "createdAt": "2023-01-01T00:00:00.000Z",
      "updatedAt": "2023-01-01T00:00:00.000Z",
      "description": "Ever wonder how?",
      "tagList": ["training", "dragons"],
      "author": {
        "username": "johnjacob",
        "bio": "Dragon trainer",
        "image": "https://example.com/avatar.jpg",
        "following": false
      },
      "favorited": false,
      "favoritesCount": 0
    }
  ],
  "articlesCount": 1
}
```

### 2.2 创建文章 (Create Article)

**接口信息**
- 名称：创建文章
- 路径：`/articles`
- 请求方法：POST
- 功能描述：创建新文章

**请求参数**
| 参数名 | 数据类型 | 是否必填 | 描述 |
|-------|---------|---------|------|
| article.title | string | 是 | 文章标题 |
| article.description | string | 是 | 文章描述 |
| article.body | string | 是 | 文章内容 |
| article.tagList | array | 否 | 文章标签列表 |

**请求头**
- Content-Type: application/json
- X-Requested-With: XMLHttpRequest
- Authorization: Token {{token}}

**请求示例**
```json
{
  "article": {
    "title": "How to train your dragon",
    "description": "Ever wonder how?",
    "body": "Very carefully.",
    "tagList": ["training", "dragons"]
  }
}
```

**响应数据**
| 字段名 | 数据类型 | 描述 |
|-------|---------|------|
| article.title | string | 文章标题 |
| article.slug | string | 文章唯一标识 |
| article.body | string | 文章内容 |
| article.createdAt | string | 创建时间（ISO 8601格式） |
| article.updatedAt | string | 更新时间（ISO 8601格式） |
| article.description | string | 文章描述 |
| article.tagList | array | 文章标签列表 |
| article.author | object | 作者信息 |
| article.author.username | string | 作者用户名 |
| article.author.bio | string | 作者简介 |
| article.author.image | string | 作者头像URL |
| article.author.following | boolean | 当前用户是否关注作者 |
| article.favorited | boolean | 当前用户是否收藏文章 |
| article.favoritesCount | integer | 文章收藏数 |

**成功响应示例**
```json
{
  "article": {
    "title": "How to train your dragon",
    "slug": "how-to-train-your-dragon",
    "body": "Very carefully.",
    "createdAt": "2023-01-01T00:00:00.000Z",
    "updatedAt": "2023-01-01T00:00:00.000Z",
    "description": "Ever wonder how?",
    "tagList": ["training", "dragons"],
    "author": {
      "username": "user123",
      "bio": "Software developer",
      "image": "https://example.com/avatar.jpg",
      "following": false
    },
    "favorited": false,
    "favoritesCount": 0
  }
}
```

**错误码说明**
| 错误码 | 描述 |
|-------|------|
| 401 Unauthorized | 无效令牌 |

### 2.3 获取文章详情 (Single Article)

**接口信息**
- 名称：获取文章详情
- 路径：`/articles/{slug}`
- 请求方法：GET
- 功能描述：获取指定文章的详细信息

**请求参数**
| 参数名 | 数据类型 | 是否必填 | 描述 |
|-------|---------|---------|------|
| slug | string | 是 | 文章唯一标识（路径参数） |

**请求头**
- Content-Type: application/json
- X-Requested-With: XMLHttpRequest
- Authorization: Token {{token}} (可选)

**请求示例**
```
GET /articles/how-to-train-your-dragon
```

**响应数据**
| 字段名 | 数据类型 | 描述 |
|-------|---------|------|
| article.title | string | 文章标题 |
| article.slug | string | 文章唯一标识 |
| article.body | string | 文章内容 |
| article.createdAt | string | 创建时间（ISO 8601格式） |
| article.updatedAt | string | 更新时间（ISO 8601格式） |
| article.description | string | 文章描述 |
| article.tagList | array | 文章标签列表 |
| article.author | object | 作者信息 |
| article.author.username | string | 作者用户名 |
| article.author.bio | string | 作者简介 |
| article.author.image | string | 作者头像URL |
| article.author.following | boolean | 当前用户是否关注作者 |
| article.favorited | boolean | 当前用户是否收藏文章 |
| article.favoritesCount | integer | 文章收藏数 |

**成功响应示例**
```json
{
  "article": {
    "title": "How to train your dragon",
    "slug": "how-to-train-your-dragon",
    "body": "Very carefully.",
    "createdAt": "2023-01-01T00:00:00.000Z",
    "updatedAt": "2023-01-01T00:00:00.000Z",
    "description": "Ever wonder how?",
    "tagList": ["training", "dragons"],
    "author": {
      "username": "johnjacob",
      "bio": "Dragon trainer",
      "image": "https://example.com/avatar.jpg",
      "following": false
    },
    "favorited": false,
    "favoritesCount": 0
  }
}
```

**错误码说明**
| 错误码 | 描述 |
|-------|------|
| 404 Not Found | 文章不存在 |

### 2.4 更新文章 (Update Article)

**接口信息**
- 名称：更新文章
- 路径：`/articles/{slug}`
- 请求方法：PUT
- 功能描述：更新指定文章的信息

**请求参数**
| 参数名 | 数据类型 | 是否必填 | 描述 |
|-------|---------|---------|------|
| slug | string | 是 | 文章唯一标识（路径参数） |
| article.title | string | 否 | 文章标题 |
| article.description | string | 否 | 文章描述 |
| article.body | string | 否 | 文章内容 |

**请求头**
- Content-Type: application/json
- X-Requested-With: XMLHttpRequest
- Authorization: Token {{token}}

**请求示例**
```json
{
  "article": {
    "body": "With two hands"
  }
}
```

**响应数据**
| 字段名 | 数据类型 | 描述 |
|-------|---------|------|
| article.title | string | 文章标题 |
| article.slug | string | 文章唯一标识 |
| article.body | string | 文章内容 |
| article.createdAt | string | 创建时间（ISO 8601格式） |
| article.updatedAt | string | 更新时间（ISO 8601格式） |
| article.description | string | 文章描述 |
| article.tagList | array | 文章标签列表 |
| article.author | object | 作者信息 |
| article.author.username | string | 作者用户名 |
| article.author.bio | string | 作者简介 |
| article.author.image | string | 作者头像URL |
| article.author.following | boolean | 当前用户是否关注作者 |
| article.favorited | boolean | 当前用户是否收藏文章 |
| article.favoritesCount | integer | 文章收藏数 |

**成功响应示例**
```json
{
  "article": {
    "title": "How to train your dragon",
    "slug": "how-to-train-your-dragon",
    "body": "With two hands",
    "createdAt": "2023-01-01T00:00:00.000Z",
    "updatedAt": "2023-01-01T00:00:00.000Z",
    "description": "Ever wonder how?",
    "tagList": ["training", "dragons"],
    "author": {
      "username": "user123",
      "bio": "Software developer",
      "image": "https://example.com/avatar.jpg",
      "following": false
    },
    "favorited": false,
    "favoritesCount": 0
  }
}
```

**错误码说明**
| 错误码 | 描述 |
|-------|------|
| 401 Unauthorized | 无效令牌 |
| 403 Forbidden | 无权限更新 |
| 404 Not Found | 文章不存在 |

### 2.5 删除文章 (Delete Article)

**接口信息**
- 名称：删除文章
- 路径：`/articles/{slug}`
- 请求方法：DELETE
- 功能描述：删除指定文章

**请求参数**
| 参数名 | 数据类型 | 是否必填 | 描述 |
|-------|---------|---------|------|
| slug | string | 是 | 文章唯一标识（路径参数） |

**请求头**
- Content-Type: application/json
- X-Requested-With: XMLHttpRequest
- Authorization: Token {{token}}

**请求示例**
```
DELETE /articles/how-to-train-your-dragon
Authorization: Token eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应数据**
| 字段名 | 数据类型 | 描述 |
|-------|---------|------|
| N/A | N/A | N/A |

**成功响应示例**
```
200 OK
```

**错误码说明**
| 错误码 | 描述 |
|-------|------|
| 401 Unauthorized | 无效令牌 |
| 403 Forbidden | 无权限删除 |
| 404 Not Found | 文章不存在 |

### 2.6 收藏文章 (Favorite Article)

**接口信息**
- 名称：收藏文章
- 路径：`/articles/{slug}/favorite`
- 请求方法：POST
- 功能描述：收藏指定文章

**请求参数**
| 参数名 | 数据类型 | 是否必填 | 描述 |
|-------|---------|---------|------|
| slug | string | 是 | 文章唯一标识（路径参数） |

**请求头**
- Content-Type: application/json
- X-Requested-With: XMLHttpRequest
- Authorization: Token {{token}}

**请求示例**
```
POST /articles/how-to-train-your-dragon/favorite
Authorization: Token eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应数据**
| 字段名 | 数据类型 | 描述 |
|-------|---------|------|
| article.title | string | 文章标题 |
| article.slug | string | 文章唯一标识 |
| article.body | string | 文章内容 |
| article.createdAt | string | 创建时间（ISO 8601格式） |
| article.updatedAt | string | 更新时间（ISO 8601格式） |
| article.description | string | 文章描述 |
| article.tagList | array | 文章标签列表 |
| article.author | object | 作者信息 |
| article.author.username | string | 作者用户名 |
| article.author.bio | string | 作者简介 |
| article.author.image | string | 作者头像URL |
| article.author.following | boolean | 当前用户是否关注作者 |
| article.favorited | boolean | 当前用户是否收藏文章 |
| article.favoritesCount | integer | 文章收藏数 |

**成功响应示例**
```json
{
  "article": {
    "title": "How to train your dragon",
    "slug": "how-to-train-your-dragon",
    "body": "Very carefully.",
    "createdAt": "2023-01-01T00:00:00.000Z",
    "updatedAt": "2023-01-01T00:00:00.000Z",
    "description": "Ever wonder how?",
    "tagList": ["training", "dragons"],
    "author": {
      "username": "johnjacob",
      "bio": "Dragon trainer",
      "image": "https://example.com/avatar.jpg",
      "following": false
    },
    "favorited": true,
    "favoritesCount": 1
  }
}
```

**错误码说明**
| 错误码 | 描述 |
|-------|------|
| 401 Unauthorized | 无效令牌 |
| 404 Not Found | 文章不存在 |

### 2.7 取消收藏文章 (Unfavorite Article)

**接口信息**
- 名称：取消收藏文章
- 路径：`/articles/{slug}/favorite`
- 请求方法：DELETE
- 功能描述：取消收藏指定文章

**请求参数**
| 参数名 | 数据类型 | 是否必填 | 描述 |
|-------|---------|---------|------|
| slug | string | 是 | 文章唯一标识（路径参数） |

**请求头**
- Content-Type: application/json
- X-Requested-With: XMLHttpRequest
- Authorization: Token {{token}}

**请求示例**
```
DELETE /articles/how-to-train-your-dragon/favorite
Authorization: Token eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应数据**
| 字段名 | 数据类型 | 描述 |
|-------|---------|------|
| article.title | string | 文章标题 |
| article.slug | string | 文章唯一标识 |
| article.body | string | 文章内容 |
| article.createdAt | string | 创建时间（ISO 8601格式） |
| article.updatedAt | string | 更新时间（ISO 8601格式） |
| article.description | string | 文章描述 |
| article.tagList | array | 文章标签列表 |
| article.author | object | 作者信息 |
| article.author.username | string | 作者用户名 |
| article.author.bio | string | 作者简介 |
| article.author.image | string | 作者头像URL |
| article.author.following | boolean | 当前用户是否关注作者 |
| article.favorited | boolean | 当前用户是否收藏文章 |
| article.favoritesCount | integer | 文章收藏数 |

**成功响应示例**
```json
{
  "article": {
    "title": "How to train your dragon",
    "slug": "how-to-train-your-dragon",
    "body": "Very carefully.",
    "createdAt": "2023-01-01T00:00:00.000Z",
    "updatedAt": "2023-01-01T00:00:00.000Z",
    "description": "Ever wonder how?",
    "tagList": ["training", "dragons"],
    "author": {
      "username": "johnjacob",
      "bio": "Dragon trainer",
      "image": "https://example.com/avatar.jpg",
      "following": false
    },
    "favorited": false,
    "favoritesCount": 0
  }
}
```

**错误码说明**
| 错误码 | 描述 |
|-------|------|
| 401 Unauthorized | 无效令牌 |
| 404 Not Found | 文章不存在 |

## 3. 评论相关接口 (Comments)

### 3.1 创建评论 (Create Comment for Article)

**接口信息**
- 名称：创建评论
- 路径：`/articles/{slug}/comments`
- 请求方法：POST
- 功能描述：为指定文章创建评论

**请求参数**
| 参数名 | 数据类型 | 是否必填 | 描述 |
|-------|---------|---------|------|
| slug | string | 是 | 文章唯一标识（路径参数） |
| comment.body | string | 是 | 评论内容 |

**请求头**
- Content-Type: application/json
- X-Requested-With: XMLHttpRequest
- Authorization: Token {{token}}

**请求示例**
```json
{
  "comment": {
    "body": "Thank you so much!"
  }
}
```

**响应数据**
| 字段名 | 数据类型 | 描述 |
|-------|---------|------|
| comment.id | integer | 评论ID |
| comment.body | string | 评论内容 |
| comment.createdAt | string | 创建时间（ISO 8601格式） |
| comment.updatedAt | string | 更新时间（ISO 8601格式） |
| comment.author | object | 作者信息 |
| comment.author.username | string | 作者用户名 |
| comment.author.bio | string | 作者简介 |
| comment.author.image | string | 作者头像URL |
| comment.author.following | boolean | 当前用户是否关注作者 |

**成功响应示例**
```json
{
  "comment": {
    "id": 1,
    "body": "Thank you so much!",
    "createdAt": "2023-01-01T00:00:00.000Z",
    "updatedAt": "2023-01-01T00:00:00.000Z",
    "author": {
      "username": "user123",
      "bio": "Software developer",
      "image": "https://example.com/avatar.jpg",
      "following": false
    }
  }
}
```

**错误码说明**
| 错误码 | 描述 |
|-------|------|
| 401 Unauthorized | 无效令牌 |
| 404 Not Found | 文章不存在 |

### 3.2 获取文章评论 (All Comments for Article)

**接口信息**
- 名称：获取文章评论
- 路径：`/articles/{slug}/comments`
- 请求方法：GET
- 功能描述：获取指定文章的所有评论

**请求参数**
| 参数名 | 数据类型 | 是否必填 | 描述 |
|-------|---------|---------|------|
| slug | string | 是 | 文章唯一标识（路径参数） |

**请求头**
- Content-Type: application/json
- X-Requested-With: XMLHttpRequest
- Authorization: Token {{token}} (可选)

**请求示例**
```
GET /articles/how-to-train-your-dragon/comments
```

**响应数据**
| 字段名 | 数据类型 | 描述 |
|-------|---------|------|
| comments | array | 评论列表 |
| comments[].id | integer | 评论ID |
| comments[].body | string | 评论内容 |
| comments[].createdAt | string | 创建时间（ISO 8601格式） |
| comments[].updatedAt | string | 更新时间（ISO 8601格式） |
| comments[].author | object | 作者信息 |
| comments[].author.username | string | 作者用户名 |
| comments[].author.bio | string | 作者简介 |
| comments[].author.image | string | 作者头像URL |
| comments[].author.following | boolean | 当前用户是否关注作者 |

**成功响应示例**
```json
{
  "comments": [
    {
      "id": 1,
      "body": "Thank you so much!",
      "createdAt": "2023-01-01T00:00:00.000Z",
      "updatedAt": "2023-01-01T00:00:00.000Z",
      "author": {
        "username": "user123",
        "bio": "Software developer",
        "image": "https://example.com/avatar.jpg",
        "following": false
      }
    }
  ]
}
```

**错误码说明**
| 错误码 | 描述 |
|-------|------|
| 404 Not Found | 文章不存在 |

### 3.3 删除评论 (Delete Comment for Article)

**接口信息**
- 名称：删除评论
- 路径：`/articles/{slug}/comments/{commentId}`
- 请求方法：DELETE
- 功能描述：删除指定评论

**请求参数**
| 参数名 | 数据类型 | 是否必填 | 描述 |
|-------|---------|---------|------|
| slug | string | 是 | 文章唯一标识（路径参数） |
| commentId | integer | 是 | 评论ID（路径参数） |

**请求头**
- Content-Type: application/json
- X-Requested-With: XMLHttpRequest
- Authorization: Token {{token}}

**请求示例**
```
DELETE /articles/how-to-train-your-dragon/comments/1
Authorization: Token eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应数据**
| 字段名 | 数据类型 | 描述 |
|-------|---------|------|
| N/A | N/A | N/A |

**成功响应示例**
```
200 OK
```

**错误码说明**
| 错误码 | 描述 |
|-------|------|
| 401 Unauthorized | 无效令牌 |
| 403 Forbidden | 无权限删除 |
| 404 Not Found | 评论不存在 |

## 4. 用户资料相关接口 (Profiles)

### 4.1 获取用户资料 (Profile)

**接口信息**
- 名称：获取用户资料
- 路径：`/profiles/{username}`
- 请求方法：GET
- 功能描述：获取指定用户的资料信息

**请求参数**
| 参数名 | 数据类型 | 是否必填 | 描述 |
|-------|---------|---------|------|
| username | string | 是 | 用户名（路径参数） |

**请求头**
- Content-Type: application/json
- X-Requested-With: XMLHttpRequest
- Authorization: Token {{token}} (可选)

**请求示例**
```
GET /profiles/johnjacob
```

**响应数据**
| 字段名 | 数据类型 | 描述 |
|-------|---------|------|
| profile.username | string | 用户名 |
| profile.bio | string | 个人简介 |
| profile.image | string | 头像URL |
| profile.following | boolean | 当前用户是否关注该用户 |

**成功响应示例**
```json
{
  "profile": {
    "username": "johnjacob",
    "bio": "Dragon trainer",
    "image": "https://example.com/avatar.jpg",
    "following": false
  }
}
```

**错误码说明**
| 错误码 | 描述 |
|-------|------|
| 404 Not Found | 用户不存在 |

### 4.2 关注用户 (Follow Profile)

**接口信息**
- 名称：关注用户
- 路径：`/profiles/{username}/follow`
- 请求方法：POST
- 功能描述：关注指定用户

**请求参数**
| 参数名 | 数据类型 | 是否必填 | 描述 |
|-------|---------|---------|------|
| username | string | 是 | 用户名（路径参数） |

**请求头**
- Content-Type: application/json
- X-Requested-With: XMLHttpRequest
- Authorization: Token {{token}}

**请求示例**
```
POST /profiles/johnjacob/follow
Authorization: Token eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应数据**
| 字段名 | 数据类型 | 描述 |
|-------|---------|------|
| profile.username | string | 用户名 |
| profile.bio | string | 个人简介 |
| profile.image | string | 头像URL |
| profile.following | boolean | 当前用户是否关注该用户 |

**成功响应示例**
```json
{
  "profile": {
    "username": "johnjacob",
    "bio": "Dragon trainer",
    "image": "https://example.com/avatar.jpg",
    "following": true
  }
}
```

**错误码说明**
| 错误码 | 描述 |
|-------|------|
| 401 Unauthorized | 无效令牌 |
| 404 Not Found | 用户不存在 |

### 4.3 取消关注用户 (Unfollow Profile)

**接口信息**
- 名称：取消关注用户
- 路径：`/profiles/{username}/follow`
- 请求方法：DELETE
- 功能描述：取消关注指定用户

**请求参数**
| 参数名 | 数据类型 | 是否必填 | 描述 |
|-------|---------|---------|------|
| username | string | 是 | 用户名（路径参数） |

**请求头**
- Content-Type: application/json
- X-Requested-With: XMLHttpRequest
- Authorization: Token {{token}}

**请求示例**
```
DELETE /profiles/johnjacob/follow
Authorization: Token eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应数据**
| 字段名 | 数据类型 | 描述 |
|-------|---------|------|
| profile.username | string | 用户名 |
| profile.bio | string | 个人简介 |
| profile.image | string | 头像URL |
| profile.following | boolean | 当前用户是否关注该用户 |

**成功响应示例**
```json
{
  "profile": {
    "username": "johnjacob",
    "bio": "Dragon trainer",
    "image": "https://example.com/avatar.jpg",
    "following": false
  }
}
```

**错误码说明**
| 错误码 | 描述 |
|-------|------|
| 401 Unauthorized | 无效令牌 |
| 404 Not Found | 用户不存在 |

## 5. 标签相关接口 (Tags)

### 5.1 获取所有标签 (All Tags)

**接口信息**
- 名称：获取所有标签
- 路径：`/tags`
- 请求方法：GET
- 功能描述：获取所有文章标签

**请求参数**
| 参数名 | 数据类型 | 是否必填 | 描述 |
|-------|---------|---------|------|
| N/A | N/A | N/A | N/A |

**请求头**
- Content-Type: application/json
- X-Requested-With: XMLHttpRequest

**请求示例**
```
GET /tags
```

**响应数据**
| 字段名 | 数据类型 | 描述 |
|-------|---------|------|
| tags | array | 标签列表 |

**成功响应示例**
```json
{
  "tags": ["training", "dragons", "javascript", "react"]
}
```

## 6. 错误码说明

| 错误码 | 描述 | 常见原因 |
|-------|------|---------|
| 400 Bad Request | 请求参数错误 | 邮箱已存在、用户名已存在、参数格式错误 |
| 401 Unauthorized | 未授权访问 | 无效令牌、邮箱或密码错误 |
| 403 Forbidden | 禁止访问 | 无权限执行操作 |
| 404 Not Found | 资源不存在 | 文章不存在、用户不存在、评论不存在 |
| 500 Internal Server Error | 服务器内部错误 | 服务器处理请求时发生错误 |

## 7. 认证说明

所有需要认证的接口都需要在请求头中添加 `Authorization` 字段，格式为：
```
Authorization: Token {token}
```

其中 `{token}` 是用户登录或注册时获取的 JWT 令牌。

## 8. 分页说明

对于返回列表的接口（如获取文章列表），支持以下分页参数：
- `page`：页码，默认值为 1
- `limit`：每页数量，默认值为 20

例如：
```
GET /articles?page=2&limit=10
```

## 9. 排序说明

对于返回列表的接口，默认按创建时间倒序排序（最新的在前）。

## 10. 速率限制

API 默认对每个用户实施速率限制，具体限制如下：
- 每小时最多 1000 个请求
- 每分钟最多 60 个请求

超过限制的请求会返回 `429 Too Many Requests` 错误。

## 11. 跨域支持

API 支持跨域请求，通过以下响应头实现：
- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS`
- `Access-Control-Allow-Headers: Content-Type, Authorization, X-Requested-With`

## 12. 版本控制

当前 API 版本为 v1，通过 URL 路径前缀 `/api` 标识。

## 13. 部署环境

API 部署在以下环境：
- 开发环境：http://localhost:8080/api
- 测试环境：https://test-api.conduit.com/api
- 生产环境：https://api.conduit.com/api

## 14. 联系方式

如有 API 使用问题，请联系：
- 邮箱：api-support@conduit.com
- 文档：https://docs.conduit.com/api

## 15. 变更日志

### v1.0.0 (2023-01-01)
- 初始版本发布
- 支持用户认证、文章管理、评论管理、用户关注等功能

### v1.0.1 (2023-02-15)
- 修复了部分接口的错误处理
- 优化了认证流程

### v1.0.2 (2023-03-30)
- 添加了速率限制
- 改进了文档结构

---

此文档基于 Conduit API 规范编写，旨在为开发人员提供清晰、准确的接口信息，便于集成和使用。