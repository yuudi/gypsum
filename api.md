# 接口文档

base path: `/api/v1`

请求成功时，状态码是 `2xx` 或 `4xx`，状态码为 `4xx` 时，会有 `message` 字段说明错误原因

## 消息规则

对象结构：消息规则

| 字段         | 类型            | 含义                                                                                  |
| ------------ | --------------- | ------------------------------------------------------------------------------------- |
| activate     | boolean         | 当前规则是否启用                                                                      |
| message_type | integer\*       | 匹配的消息类型                                                                        |
| group_id     | integer         | 匹配群，`0`表示所有                                                                   |
| user_id      | integer         | 匹配 QQ 号，`0`表示所有                                                               |
| matcher_type | integer         | 匹配方式<br/>0 完全匹配，1 关键词匹配，1 前缀匹配，3 后缀匹配，4 命令匹配，5 正则匹配 |
| patterns     | array\<string\> | 匹配表达式的数组                                                                      |
| response     | string          | 回复模板                                                                              |
| priority     | integer         | 优先级                                                                                |
| block        | boolean         | 是否阻止后续规则                                                                      |

消息类型编号为

| 类型         | 编号 |
| ------------ | ---- |
| 好友消息     | 0x01 |
| 群临时消息   | 0x02 |
| 其他临时消息 | 0x04 |
| 公众号消息   | 0x08 |
| 群普通消息   | 0x10 |
| 群匿名消息   | 0x20 |
| 群系统通知   | 0x40 |
| 讨论组消息   | 0x80 |

如需同时匹配多种消息可用`位或`运算，例如：0x07 匹配所有私聊消息

### 列出所有规则

GET `/rules`

返回一个对象，key 是整数（即`rule_id`，不一定连续），value 是规则

### 查看规则

GET `/rules/{rule_id}`

返回一个规则

### 添加规则

POST `/rules`

请求体为一条规则

返回 `status 201` `code=0`

### 删除规则

DELETE `/rules/{rule_id}`

返回 `code=0`

### 修改规则

PUT `/rules/{rule_id}`

请求体为一条规则

返回 `code=0`

## 触发事件

对象结构：事件规则

| 字段         | 类型    | 含义                                                              |
| ------------ | ------- | ----------------------------------------------------------------- |
| activate     | boolean | 当前规则是否启用                                                  |
| group_id     | integer | 匹配群，`0`表示所有                                               |
| user_id      | integer | 匹配 QQ 号，`0`表示所有                                           |
| trigger_type | string  | 触发事件，具体见<a href="javascript:alert('咕咕咕')">模板文档</a> |
| response     | string  | 回复模板                                                          |
| priority     | integer | 优先级                                                            |
| block        | boolean | 是否阻止后续规则                                                  |

### 列出所有事件规则

GET `/triggers`

返回一个对象，key 是整数（即`trigger_id`，不一定连续），value 是规则

### 查看事件规则

GET `/triggers/{trigger_id}`

返回一个规则

### 添加事件规则

POST `/triggers`

请求体为一条规则

返回 `status 201` `code=0`

### 删除事件规则

DELETE `/triggers/{trigger_id}`

返回 `code=0`

### 修改事件规则

PUT `/triggers/{trigger_id}`

请求体为一条规则

返回 `code=0`

## 定时任务

对象结构：任务

| 字段      | 类型             | 含义                                                                            |
| --------- | ---------------- | ------------------------------------------------------------------------------- |
| activate  | boolean          | 当前任务是否启用                                                                |
| group_id  | array\<integer\> | 发送结果到群号                                                                  |
| user_id   | array\<integer\> | 发送结果到 QQ 号                                                                |
| once      | boolean          | 当前任务是否是一次性任务                                                        |
| cron_spec | string           | 计划任务表达式，详见[cron](https://pkg.go.dev/github.com/robfig/cron#hdr-Usage) |
| action    | string           | 执行任务模板                                                                    |

### 列出所有任务

GET `/jobs`

返回一个对象，key 是整数（即`job_id`，不一定连续），value 是规则

### 查看任务

GET `/jobs/{job_id}`

返回一个规则

### 添加任务

POST `/jobs`

请求体为一条任务

返回 `status 201` `code=0`

如果计划任务表达式语法错误，将返回 http 状态码 `422 Unprocessable Entity`

### 删除任务

DELETE `/jobs/{job_id}`

返回 `code=0`

### 修改任务

PUT `/jobs/{job_id}`

请求体为一条任务

返回 `code=0`

如果计划任务表达式语法错误，将返回 http 状态码 `422 Unprocessable Entity; code=2010`

## 静态资源

## 模板测试
