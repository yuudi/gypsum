# 接口文档

base path: `/api/v1`

对象结构：规则

| 字段         | 类型          | 含义                                                                               |
| ------------ | ------------- | ---------------------------------------------------------------------------------- |
| group_id     | integar       | 匹配群，`0`表示所有                                                                |
| user_id      | integar       | 匹配 QQ 号，`0`表示所有                                                            |
| matcher_type | integar       | 匹配方式，0 完全匹配，1 关键词匹配，1 前缀匹配，3 后缀匹配，4 命令匹配，5 正则匹配 |
| patterns     | array\<string\> | 匹配表达式的数组                                                                   |
| response     | string        | 回复模板                                                                           |
| priority     | integar       | 优先级                                                                             |
| block        | boolean       | 是否阻止后续规则                                                                   |

## 列出所有规则

GET `/rules`

返回一个对象，key 是整数（即`rule_id`，不一定连续），value 是规则

## 查看规则

GET `/rules/{rule_id}`

返回一个规则

## 添加规则

POST `/rules`

请求体为一条规则

返回 `status 201` `code=0`

## 删除规则

DELETE `/rules/{rule_id}`

返回 `code=0`

## 修改规则

PUT `/rules/{rule_id}`

请求体为一条规则

返回 `code=0`
