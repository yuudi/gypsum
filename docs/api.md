# 开发接口文档

base path: `/api/v1`

请求成功时，状态码是 `200` 或 `201`，`code` 值为 `0`（部分特殊见说明）  
当状态码为 `4xx` 时，会有 `message` 字段说明错误原因  
当状态码为 `5xx` 时，如果 `Content-Type` 不是 `application/json`，则为服务故障

## 鉴权

首先获取 `password_salt`

GET `/gypsum/information`

返回 `password_salt` `logged_in`

如果 logged_in 值为 `true` 则表示登录尚未过期，无需重复登录

然后将用户输入的密码与 `password_salt` 连接，取 sha256 值的 16 进制小写，作为验证密码

PUT `/gypsum/login`

请求体为 json，字段为 `password`，值为验证密码

返回 200 `code=0` 并得到 cookie

## 组

对象结构：组

| 字段           | 类型              | 含义                                        |
| -------------- | ----------------- | ------------------------------------------- |
| display_name   | string            | 显示名称                                    |
| plugin_name    | string            | （仅导入的组）插件名                        |
| plugin_version | integer           | （仅导入的组）插件数字版本（大于 0 的整数） |
| items          | array\<object\*\> | 项目                                        |

对象结构：项目

| 字段         | 类型    | 含义                                                                                                           |
| ------------ | ------- | -------------------------------------------------------------------------------------------------------------- |
| item_type    | string  | 项目类型<br>`rule` 消息规则<br>`trigger` 触发事件<br>`scheduler` 定时任务<br>`resource` 静态资源<br>`group` 组 |
| display_name | string  | 显示名称                                                                                                       |
| item_id      | integer | 项目编号                                                                                                       |

### 列出所有组

GET `/groups`

返回一个对象，key 是整数（即`group_id`，不一定连续），value 是`组`

### 查看组

GET `/groups/{group_id}`

返回一个`组`

### 添加组

POST `/groups`  
POST `/groups/{group_id}/groups`

如果使用第二种路由，则需要指定 `group_id` 为上级组（后面同理）

请求体为 `json`，只有 `display_name` 字段，例如：`{"display_name":"my group"}`

返回 `status 201` `code=0`

### 移动组项目

PUT `/groups/{group_id}/items/{item_type}/{item_id}`

将一个项目移动至一个组  
（一个项目只能属于一个组，`group_id=0` 表示不属于任何组）

如果将组移动至其子组中，将返回 http 状态码 `422 Unprocessable Entity`

### 导出组

GET `/groups/{group_id}/archive`

参数：

`plugin_name` 导出插件的名称，用于导入时识别相同插件，使用域名加路径（不带`http://`），如无域名则可用 `github.com` 加用户名加插件名，如 `github.com/yuudi/gypsum`  
`plugin_version` 导出插件的数字版本，用于导入时识别版本，任意递增数字即可，如时间戳

例如 `GET /api/v1/groups/{group_id}/archive?plugin_name=github.com%2Fyuudi%2Fgypsum&plugin_version=1`

返回一个二进制文件（扩展名是 .gypsum，本身是一个 zip 压缩包）

### 导入组

POST `/groups`  
POST `/groups/{group_id}/groups`

请求体为二进制文件，即由`导出`获得的文件。请求头需设置 `Content-Type: application/zip`，否则会被视为[添加组](#添加组)

返回 `status 201` `code=0` 或 `status 415`

### 删除组

DELETE `/groups/{group_id}`

请求体为 `json`，`move_to` 值表示组中项目移动到的新组，默认值 `0`。例如：`{"move_to"=2}`。不可直接删除所有项目。

### 修改组

只能修改组名

PATCH `/groups/{group_id}`

请求体为 `json`，只有 `display_name` 字段，例如：`{"display_name":"new group name"}`

## 消息规则

对象结构：消息规则

| 字段         | 类型             | 含义                                                                                                             |
| ------------ | ---------------- | ---------------------------------------------------------------------------------------------------------------- |
| display_name | string           | 显示名称                                                                                                         |
| activate     | boolean          | 当前规则是否启用                                                                                                 |
| message_type | integer\*        | 匹配的消息类型                                                                                                   |
| groups_id    | array\<integer\> | 匹配群，留空表示所有                                                                                             |
| users_id     | array\<integer\> | 匹配 QQ 号，留空表示所有                                                                                         |
| matcher_type | integer          | 匹配方式<br/>`0` 完全匹配<br/>`1` 关键词匹配<br/>`2` 前缀匹配<br/>`3` 后缀匹配<br/>`4` 命令匹配<br/>`5` 正则匹配 |
| only_at_me   | boolean          | 是否只有被 at 才会触发                                                                                           |
| patterns     | array\<string\>  | 匹配表达式的数组                                                                                                 |
| response     | string           | 回复模板                                                                                                         |
| priority     | integer          | 优先级                                                                                                           |
| block        | boolean          | 是否阻止后续规则                                                                                                 |

消息类型编号为

| 类型         | 编号   |
| ------------ | ------ |
| 好友消息     | 0x0001 |
| 群临时消息   | 0x0002 |
| 其他临时消息 | 0x0004 |
| 公众号消息   | 0x0008 |
| 群普通消息   | 0x0010 |
| 群匿名消息   | 0x0020 |
| 群系统通知   | 0x0040 |
| 讨论组消息   | 0x0080 |

如需同时匹配多种消息可用`位或`运算，例如：0x07 匹配所有私聊消息

### 列出所有规则

GET `/rules`

返回一个对象，key 是整数（即`rule_id`，不一定连续），value 是`规则`

### 查看规则

GET `/rules/{rule_id}`

返回一个`规则`

### 添加规则

POST `/rules`  
POST `/groups/{group_id}/rules`

请求体为一条`规则`，如果匹配方式是正则匹配，那么 `patterns` 数组长度必须为 1

返回 `status 201` `code=0`

如果正则表达式语法错误，将返回 http 状态码 `422 Unprocessable Entity`

### 删除规则

DELETE `/rules/{rule_id}`

返回 `code=0`

### 修改规则

PUT `/rules/{rule_id}`

请求体为一条`规则`，如果匹配方式是正则匹配，那么 `patterns` 数组长度必须为 1

返回 `code=0`

如果正则表达式语法错误，将返回 http 状态码 `422 Unprocessable Entity`

## 触发事件

对象结构：事件规则

| 字段         | 类型              | 含义                     |
| ------------ | ----------------- | ------------------------ |
| display_name | string            | 显示名称                 |
| activate     | boolean           | 当前规则是否启用         |
| groups_id    | array\<integer\>  | 匹配群，留空表示所有     |
| users_id     | array\<integer\>  | 匹配 QQ 号，留空表示所有 |
| trigger_type | \*array\<string\> | 触发事件                 |
| response     | string            | 回复模板                 |
| priority     | integer           | 优先级                   |
| block        | boolean           | 是否阻止后续规则         |

触发事件是一个字符串数组，含有 1 个或 2 个元素，格式为 `["<detail-type>", "<sub-type>"]`

其中：  
`detail-type` 为 onebot 协议中 `post_type` 或 `request_type` 的内容  
`sub-type` 为 onebot 协议中 `sub_type` 的内容，可省略

例如：  
`["group_increase","approve"]` 匹配 `群成员增加` 中的 `管理员同意入群` 事件  
`["group_increase"]` 匹配所有 `群成员增加` 事件

### 列出所有事件规则

GET `/triggers`

返回一个对象，key 是整数（即`trigger_id`，不一定连续），value 是规则

### 查看事件规则

GET `/triggers/{trigger_id}`

返回一个`规则`

### 添加事件规则

POST `/triggers`  
POST `/groups/{group_id}/triggers`

请求体为一条`规则`

返回 `status 201` `code=0`

### 删除事件规则

DELETE `/triggers/{trigger_id}`

返回 `code=0`

### 修改事件规则

PUT `/triggers/{trigger_id}`

请求体为一条`规则`

返回 `code=0`

## 定时任务

对象结构：任务

| 字段         | 类型             | 含义                                                                            |
| ------------ | ---------------- | ------------------------------------------------------------------------------- |
| display_name | string           | 显示名称                                                                        |
| activate     | boolean          | 当前任务是否启用                                                                |
| group_id     | array\<integer\> | 发送结果到群号                                                                  |
| user_id      | array\<integer\> | 发送结果到 QQ 号                                                                |
| once         | boolean          | 当前任务是否是一次性任务                                                        |
| cron_spec    | string           | 计划任务表达式，详见[cron](https://pkg.go.dev/github.com/robfig/cron#hdr-Usage) |
| action       | string           | 执行任务模板                                                                    |

### 列出所有任务

GET `/jobs`

返回一个对象，key 是整数（即`job_id`，不一定连续），value 是`规则`

### 查看任务

GET `/jobs/{job_id}`

返回一个`任务`

### 添加任务

POST `/jobs`  
POST `/groups/{group_id}/jobs`

请求体为一条`任务`

返回 `status 201` `code=0`

如果计划任务表达式语法错误，将返回 http 状态码 `422 Unprocessable Entity`

### 删除任务

DELETE `/jobs/{job_id}`

返回 `code=0`

### 修改任务

PUT `/jobs/{job_id}`

请求体为一条`任务`

返回 `code=0`

如果计划任务表达式语法错误，将返回 http 状态码 `422 Unprocessable Entity; code=2010`

## 静态资源

对象结构：资源

| 字段       | 类型   | 含义                         |
| ---------- | ------ | ---------------------------- |
| file_name  | string | 文件名称（不含扩展名）       |
| ext        | string | 文件扩展名（包含点号）       |
| sha256_sum | string | 文件散列值，十六进制小写字母 |

### 列出所有资源

GET `/resources`

返回一个对象，key 是整数（即`resource_id`，不一定连续），value 是`资源`

### 查看资源

GET `/resources/{resource_id}`

返回一个`资源`

GET `/resources/{sha256_sum}`

返回 `status 302` 至 `resource_id` 格式的的 URI

### 下载资源

GET `/resources/{resource_id}/content`

返回资源的二进制文件，文件名包含在标头 `Content-Disposition` 字段中

资源可用 `ETag` 与 `If-None-Match` 标记缓存，缓存匹配时返回 `status 304`（可由浏览器自动处理）

### 上传资源

POST `/resources/{file_name}{ext}`  
POST `/groups/{group_id}/resources/{file_name}{ext}`

文件名与扩展名没有分隔符，例如：`POST /api/v1/resources/%e8%a1%a8%e6%83%85%e5%8c%85.jpg`

请求体为二进制文件

返回 `status 201` `code=0`：成功，返回 `resource_id`  
返回 `status 200` `code=1`：资源已经存在，无需重复上传，返回已有的 `resource_id`

上传资源前，可以先通过 `GET /resources/{sha256_sum}` 查询资源是否已存在（非必须）

### 删除资源

DELETE `/resources/{resource_id}`

返回 `code=0`

### 修改资源

只能修改资源的文件名，扩展名与散列值无法修改

PATCH `/resources/{resource_id}`

请求体为 `json`，只有 `file_name` 字段，例如：`{"file_name":"a better name"}`

## 模板测试

### 测试模板

POST `/debug`

注意：即使是测试中，模板也会被执行一次。所以数据库操作、bot api 调用都会实际执行。

| 字段         | 类型      | 含义                                                                                                                           |
| ------------ | --------- | ------------------------------------------------------------------------------------------------------------------------------ |
| event        | object    | （仅消息测试与通知测试）onebot 事件                                                                                            |
| debug_type   | string    | `message` 或 `notice` 或 `schedule`                                                                                            |
| matcher_type | \*integer | （仅消息测试）匹配方式<br/>`0` 完全匹配<br/>`1` 关键词匹配<br/>`2` 前缀匹配<br/>`3` 后缀匹配<br/>`4` 命令匹配<br/>`5` 正则匹配 |
| pattern      | string    | （仅消息测试）匹配表达式                                                                                                       |
| response     | string    | 回复模板                                                                                                                       |

\* 见[消息规则](#消息规则)

| 字段    | 类型    | 含义                                                      |
| ------- | ------- | --------------------------------------------------------- |
| code    | integer | `0` 表示成功，其他表示失败，失败信息在 `message` 字段获取 |
| matched | boolean | 消息测试中表示是否成功匹配消息，其他情况始终为 `true`     |
| reply   | string  | 发送的消息                                                |

## bot

### 获取所有群 （进行中）

### 获取所有好友 （进行中）

### 获取群成员 （进行中）

## gypsum

### 获取版本信息

GET `/gypsum/information`

返回 `version` `commit` `password_salt` `logged_in` `platform`

### 更新

GET `/gypsum/update`

获取更新状态

PUT `/gypsum/update`

开始更新，更新完毕会自动重启

| 字段          | 类型    | 含义                                    |
| ------------- | ------- | --------------------------------------- |
| new_version   | string  | 指定版本，可填 `stable` `beta` `v1.0.0` |
| mirror        | string  | 指定下载镜像站（将替换 `github.com`）   |
| forced_update | boolean | 强制更新                                |
