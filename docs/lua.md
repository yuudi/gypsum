# Lua 代码块

## 变量

### event

`event` 是收到的事件对象，具体结构可参照 [onebot 标准](https://github.com/howmanybots/onebot/blob/master/v11/specs/event)  
只有事件触发的模板才能使用 `event`，定时任务中的模板是没有 `event` 对象的。

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

## 模块

### bot

与收发消息相关的模块

#### send

立刻发送一条消息，此方法默认会对 CQ 码进行转义。

参数：第一个参数为字符串，表示要发送的消息。第二个参数为布尔，表示是否取消转义，省略则为 false。

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

#### send_private

向指定用户发送一条私聊消息，此方法默认会对 CQ 码进行转义。

参数：第一个参数为数字，表示要发送的 QQ 号。第二个参数为字符串，表示要发送的消息。第三个参数为布尔，表示是否取消转义，省略则为 false。

用法示例：

```lua
{% lua %}
local bot = require("bot")

bot.send_private(event.user_id, "您的暗骰点数为：" .. math.random(1, 6))
{% endlua %}
暗骰已完成，请查看私聊
```

#### send_group

向指定群发送一条消息，此方法默认会对 CQ 码进行转义。

同上，略。

#### get

获取下一个消息

参数：第一个参数为数字，表示指定用户，省略时为当前用户，`0`表示所有用户。第二个参数为数字，表示指定群，省略时为当前群，`0`表示私聊。第三个参数为数字，表示超时时间秒，省略则为 30 秒。

返回值：成功时返回对方的消息，超时返回两个 nil。错误时第一个返回值为 nil，第二个返回值为错误信息

限制：如果在定时任务中使用，则必须指定 QQ 号与群号

用法示例：

```lua
{% lua %}
local bot = require("bot")

send("请问您需要查找哪座城市的天气？")
reply = bot.get()
if(reply==nil)
then
    -- 返回值为 nil，表示用户没有发送消息
    write("查询已超时")
else
    write(reply + "的天气是晴天")
end
{% endlua %}
```

#### approve

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

#### withdraw

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

#### set_title

设置群头衔

参数：第一个参数是字符串，表示头衔。后续参数是数字，表示 qq 号，省略则使用发送者的 qq。

限制：仅限群聊消息、群聊事件，需要 bot 是群主

用法示例：

```lua
{% lua %}
local bot = require("bot")

bot.set_title("大佬")
{% endlua %}
你太厉害了，送给你“大佬”头衔
```

#### group_ban

群内禁言

参数：第一个参数是数字，表示时长，0 表示解除。后续参数是数字，表示 qq 号，省略则使用发送者的 qq。

限制：仅限群聊消息、群聊事件，需要 bot 是管理员

用法示例：

```lua
{% lua %}
local bot = require("bot")

bot.group_ban(60*5)
{% endlua %}
违反群规！禁言5分钟警告！
```

#### api

调用 bot api，具体方法可参照 [onebot 标准](https://github.com/howmanybots/onebot/tree/master/v11/specs/api)

参数：第一个参数为字符串，表示需要调用的方法。第二个参数为表（table），表示调用的各个参数

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

#### put

参数：第一个参数为数字或字符串，表示键值。第二个参数为数据，数据的类型只能是：数字、字符串、bool、nil、table 之一。

返回：成功时没有返回值，失败时返回值为错误信息。

#### get

参数：第一个参数为数字或字符串，表示键值。第二个参数为任意类型的数据，表示未命中时的默认值，可省略。

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

#### decode

解析 json

参数：字符串

返回：成功时返回解析结果，失败时返回 nil 与错误信息

#### encode

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

#### request

参数：第一个参数为字符串，表示请求方法。第二个参数字符串，表示请求地址。第三个参数为 table，表示选项。

选项的值：

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

#### 读写文件

示例：

```lua
{% lua %}
file1 = io.open("my_file.txt", "r")
content = file1.read("a")  -- 读取整个文件
file1.close()

file2 = io.open("new_file.txt", "w")
file2.write(content)
file2.close()
{% endlua %}
```
