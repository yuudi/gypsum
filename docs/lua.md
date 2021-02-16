# Lua 代码块

## 变量

### event

`event` 是收到的事件对象，具体结构可参照 [onebot标准](https://github.com/howmanybots/onebot/blob/master/v11/specs/event)  
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

将目标写入回复模板，此方法**不会**对 CQ 码进行转义

参数：若干个任意参数

用法示例：

```lua
{% lua %}
write_safe("[CQ:at,qq=" .. event.user_id .. "] 你好")
{% endlua %}
```

## 模块

### bot

调用 bot，具体方法可参照 [onebot标准](https://github.com/howmanybots/onebot/tree/master/v11/specs/api)

方法：api

参数：第一个参数为字符串，表示需要调用的方法。第二个参数为表（table），表示调用的各个参数

返回值：成功时返回 api 的返回值，失败时第一个返回值为 nil，第二个返回值为错误信息

api 返回值每种调用方法的返回值各不相同，具体可参照 [onebot标准](https://github.com/howmanybots/onebot/tree/master/v11/specs/api)

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

方法：put

参数：第一个参数为数字或字符串，表示键值。第二个参数为数据，数据的类型只能是：数字、字符串、bool、nil、table 之一。

返回：成功时没有返回值，失败时返回值为错误信息。

方法：get

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

用法示例：

```lua
{% lua %}
local json = require("json")

json_data = '{"code":0,"message":"ok"}'
table_data = json.decode(raw_data)
print("code is", table_data.code)
{% endlua %}
```

### http

进行 http 请求的模块，来自 [gluahttp](https://github.com/cjoudrey/gluahttp)

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
