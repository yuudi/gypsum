# Lua 代码块

## 变量

### event

`event` 是收到的事件对象，具体结构可参照 [onebot 标准](https://github.com/howmanybots/onebot/blob/master/v11/specs/event)  
只有事件触发的模板才能使用 `event`，定时任务中的模板是没有 `event` 对象的。

一些常用的变量：

| 变量                  | 事件类型           | 含义                   |
| --------------------- | ------------------ | ---------------------- |
| event.message         | 消息               | 接收到的消息           |
| event.user_id         | 消息、部分通知     | 事件的触发者           |
| event.group_id        | 群消息、群通知     | 事件的触发群，可能为空 |
| event.sender.nickname | 消息               | 消息发送者的昵称       |
| event.comment         | 加好友加群请求邀请 | 验证消息               |

### state

对于各种匹配方式，state 给出了匹配结果

#### 完全匹配

`state.matched` 为匹配的消息

#### 关键词匹配

`state.keyword` 为匹配的关键词

#### 前缀匹配

`state.prefix` 为匹配的前缀  
`state.args` 为除去前缀后的剩余部分

#### 后缀匹配

`state.suffix` 为匹配的后缀  
`state.args` 为除去后缀后的剩余部分

#### 命令匹配

`state.command` 为匹配的命令  
`state.args` 为除去命令后的剩余部分

#### 正则匹配

`state.regex_matched` 为正则匹配结果数组

示例：

用正则表达式 `(\d+)d(\d+)` 匹配消息 `1d6` 时  
`state.regex_matched[1]` 为 `1d6`  
`state.regex_matched[2]` 为 `1`  
`state.regex_matched[3]` 为 `6`

> 注意：在 lua 中 state.regex_matched 的序号是从 1 开始的

## 函数

### write

将目标写入回复模板，此方法会对 CQ 码进行转义以保证安全。如需发送 CQ 码请使用 `write_safe`

参数：若干个任意参数

用法示例：

```lua
{% lua %}
write("你好")
{% endlua %}
```

### write_safe

将目标写入回复模板，此方法**不会**对 CQ 码进行转义。

参数：若干个任意参数

用法示例：

```lua
{% lua %}
write_safe("[CQ:at,qq=" .. event.user_id .. "] 你好")
{% endlua %}
```

### sleep

等待一段时间

参数：时间，秒

用法示例：

```lua
{% lua %}
sleep(10)
write("时间到！")
{% endlua %}
```

### res

获取资源的 URI，详见[资源管理](./resources.md)

参数：字符串，表示资源的名称

返回：资源的 URI

用法示例：

```lua
{% lua %}
write(image(res("123456789abcdef.jpg")))
{% endlua %}
```

## 模块

### bot

与收发消息相关的模块

#### bot.send

立刻发送一条消息，此方法默认会对 CQ 码进行转义。

| 参数位置 | 参数类型 | 默认值 | 参数含义     |
| -------- | -------- | ------ | ------------ |
| 1        | 字符串   |        | 要发送的消息 |
| 2        | Bool     | false  | 是否取消转义 |

返回：message_id

限制：仅限消息与事件（定时任务中无法使用）

用法示例：

```lua
{% lua %}
local bot = require("bot")

bot.send("处理中，请稍等……")
sleep(10)
write("处理完毕")
{% endlua %}
```

#### bot.send_private

向指定用户发送一条私聊消息，此方法默认会对 CQ 码进行转义。

| 参数位置 | 参数类型 | 默认值 | 参数含义       |
| -------- | -------- | ------ | -------------- |
| 1        | 数字     |        | 要发送的 QQ 号 |
| 2        | 字符串   |        | 发送的消息     |
| 3        | Bool     | false  | 是否取消转义   |

返回：message_id

用法示例：

```lua
{% lua %}
local bot = require("bot")

bot.send_private(event.user_id, "您的暗骰点数为：" .. math.random(1, 6))
{% endlua %}
暗骰已完成，请查看私聊
```

#### bot.send_group

向指定群发送一条消息，此方法默认会对 CQ 码进行转义。

返回：message_id

同上，略。

#### bot.get

获取下一个消息

| 参数位置 | 参数类型                   | 默认值          | 参数含义                                    |
| -------- | -------------------------- | --------------- | ------------------------------------------- |
| 1        | 数字                       | 当前用户        | 指定接收消息的用户，`0`表示任何用户         |
| 2        | 数字                       | 当前群/当前私聊 | 指定接收消息的群，`0`表示私聊               |
| 3        | 数字                       | 30 秒           | 超时时间（秒）                              |
| 4        | 函数，接收 Table 返回 Bool | 始终为 true     | 用于筛选消息的函数，接收的 Table 表示 event |

返回值：成功时返回对方的消息，超时返回 nil 与 `"timeout"` 字符串。错误时第一个返回值为 nil，第二个返回值为错误信息

限制：如果在定时任务中使用，则必须指定 QQ 号或群号，QQ 号与群号不得同时为 0

用法示例：

```lua
{% lua %}
local bot = require("bot")

send("请问您需要查找哪座城市的天气？")
reply, err = bot.get()
if(err != nil)
then
    write(reply + "的天气是晴天")
else
    write(err)
end
{% endlua %}
```

筛选函数的用法示例：

```lua
{% lua %}
local bot = require("bot")

function is_valid_yn(ev)  -- 注意此处 ev 与全局的 event 有区别
    if(ev.message == "y" or ev.message == "n")
    then
        return true
    else
        bot.send("y or n")  -- 验证函数中也可以发送消息
        return false
    end
end

bot.send("您确定吗？请回复 y 或 n")
reply, err = bot.get(nil, nil, nil, is_valid_yn)  -- 用 nil 表示默认值，即自动获取当前的对话

if(err != nil)
then
    if(reply == "y")
    then
        write("已确认")
    else
        write("已取消")
    end
end
{% endlua %}
```

> 提示：  
> 使用验证函数的方法会持续获取消息直到成功或超时，
> 如果希望只获取一次，应当直接用 bot.get() 直接获取消息后再做验证

#### bot.approve

同意一个事件

参数：无

限制：仅限加好友请求、加群请求、加群邀请

用法示例：

```lua
{% lua %}
local bot = require("bot")
if (event.user_id == 123456)
then
  bot.approve()
end
{% endlua %}
```

#### bot.withdraw

撤回消息

参数：无

限制：仅限群聊消息，需要 bot 是管理员

用法示例：

```lua
{% lua %}
local bot = require("bot")

if (string.find(event.message, "广告"))
then
  bot.withdraw()
  write("禁止发广告")
end
{% endlua %}
```

#### bot.set_title

设置群头衔

| 参数位置 | 参数类型 | 默认值      | 参数含义     |
| -------- | -------- | ----------- | ------------ |
| 1        | 字符串   |             | 头衔         |
| 2        | 数字     | 发送者的 qq | 设置的 qq 号 |

限制：仅限群聊消息、群聊事件，需要 bot 是群主

用法示例：

```lua
{% lua %}
local bot = require("bot")

bot.set_title("大佬")
{% endlua %}
你太厉害了，送给你“大佬”头衔
```

#### bot.group_ban

群内禁言

| 参数位置 | 参数类型 | 默认值      | 参数含义         |
| -------- | -------- | ----------- | ---------------- |
| 1        | 数字     |             | 时长，0 表示解除 |
| 2        | 数字     | 发送者的 qq | 禁言的 qq 号     |

限制：仅限群聊消息、群聊事件，需要 bot 是管理员

用法示例：

```lua
{% lua %}
local bot = require("bot")

bot.group_ban(60*5)
{% endlua %}
违反群规！禁言5分钟警告！
```

#### bot.api

调用 bot api，具体方法可参照 [onebot 标准](https://github.com/howmanybots/onebot/tree/master/v11/specs/api)

| 参数位置 | 参数类型 | 默认值 | 参数含义       |
| -------- | -------- | ------ | -------------- |
| 1        | 字符串   |        | 需要调用的方法 |
| 2        | Table    |        | 调用的各个参数 |

返回值：成功时返回 api 的返回值，失败时第一个返回值为 nil，第二个返回值为错误信息

api 返回值每种调用方法的返回值各不相同，具体可参照 [onebot 标准](https://github.com/howmanybots/onebot/tree/master/v11/specs/api)

用法示例：

```lua
{% lua %}
local bot = require("bot")

message = "收到来自" .. event.sender.nickname .. "的反馈：" .. event.message
args = {
    user_id = 123456,
    message = message
}
bot.api("send_private_msg", args)
{% endlua %}
已将您的反馈发送给主人，感谢支持
```

### database

将数据存储在 gypsum 的模块

#### database.put

| 参数位置 | 参数类型                       | 默认值 | 参数含义   |
| -------- | ------------------------------ | ------ | ---------- |
| 1        | 数字、字符串                   |        | 键值       |
| 2        | 数字、字符串、Bool、Nil、Table |        | 储存的数据 |

返回：成功时没有返回值，失败时返回值为错误信息。

#### database.get

| 参数位置 | 参数类型     | 默认值 | 参数含义             |
| -------- | ------------ | ------ | -------------------- |
| 1        | 数字、字符串 |        | 键值                 |
| 2        | 任意         | nil    | 键值不存在时的默认值 |

返回：成功时返回数据，失败时第一个返回值为 nil，第二个返回值为错误信息。

用法示例：

```lua
{% lua %}
local db = require("database")

key = "sign-up" .. event.user_id
usage = db.get(key, 0)
if(usage >= 1)
then
    write("失败，您已使用过了")
else
    db.put(key, usage+1)
    write("成功")
end
{% endlua %}
```

### json

进行 json 编码解码的模块，来自 [gopher-json](https://layeh.com/gopher-json)

#### json.decode

解析 json

参数：字符串

返回：成功时返回解析结果，失败时返回 nil 与错误信息

#### json.encode

编码为 json

参数：任意

返回：成功时返回 json 字符串，失败时返回 nil 与错误信息

用法示例：

```lua
{% lua %}
local json = require("json")

json_data = '{"code":0,"message":"ok"}'
table_data = json.decode(raw_data)
print("code is", table_data.code)

table_data.new_key = "new value"
new_json = json.encode(table_data)
print(new_json)
{% endlua %}
```

### http

进行 http 请求的模块，来自 [gluahttp](https://github.com/cjoudrey/gluahttp)

#### http.request

参数：第一个参数为字符串，表示请求方法。第二个参数字符串，表示请求地址。第三个参数为 table，表示选项。

选项 Table 的字段含义：

| 字段    | 类型                                |
| ------- | ----------------------------------- |
| query   | 形如 `key=value` 的字符串           |
| body    | 字符串                              |
| cookies | Table                               |
| headers | Table                               |
| timeout | 数字                                |
| auth    | Table，有 `user` 和 `pass` 两个字段 |

#### 其他

捷径：`get` `delete` `head` `patch` `post` `put` 可直接使用，参数为地址与选项。

用法示例：

```lua
{% lua %}
local http = require("http")
local json = require("json")

response = http.get("https://api.lolicon.app/setu/")
data = json.decode(response.body)
write_safe("[CQ:image,file=" .. data.data[1].url .. "]")
{% endlua %}
```

## 标准库

在 lua 代码块中可以使用 lua 标准库与 openlib 中的函数，可参考[lua 教程](https://wizardforcel.gitbooks.io/lua-doc/content/8.html)。

### 标准库中的常用函数

#### print

打印到控制台（即 gypsum 控制台，不是发送）

#### tonumber

转化为数字，失败时返回 nil

#### tostring

转化为字符串

#### string.match

`string.match(s, pattern)`

查找字符串 s 中符合 pattern 表达式的字符串

#### string.format

`string.format(formatstring, ···)`

格式化字符串，用法类似与 C 语言的格式化

#### string.sub

`string.sub(s, i [, j])`

截取子字符串，从 i 开始到 j 结束，i 与 j 可以是负数表示从后索引

注意：lua 的索引需要从`1`开始

#### math.max

返回最大的数

#### math.min

返回最小的数

#### math.random

`math.random([m [, n]])`

无参数时从 `[0,1)` 中取随机浮点数。一个参数时从 `[1, m]` 中随机取整数。两个参数时从 `[m, n]` 中随机取整数。

#### os.date

`os.date([format [, time]])`

格式化日期，使用 [strftime 标准](https://wizardforcel.gitbooks.io/w3school-c/content/190.html)，第二个参数省略则为现在。

#### os.time

获取 unix 时间戳（秒级）
