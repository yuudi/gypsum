# 模板语法

gypsum 的模板语法使用 [pongo2](https://github.com/flosch/pongo2)，语法类似于 `jinja2` 和 `Django`

---

你可以在模板中直接编写回复

```jinja2
你好
```

---

你可以在模板中可以直接使用 [CQ码](https://github.com/howmanybots/onebot/blob/master/v11/specs/message/string.md#cq-%E7%A0%81%E6%A0%BC%E5%BC%8F)

```jinja2
[CQ:at,qq=all] 大家好
```

---

也可以用 `{{` `}}` 符号访问变量

```jinja2
你好，{{ event.sender.nickname }}
```

其中：`event` 是收到的事件对象，具体结构可参照按照 [onebot 标准](https://github.com/howmanybots/onebot/blob/master/v11/specs/event)。
注意只有事件触发的模板才能使用 `event`，定时任务中的模板是没有 `event` 对象的。

---

默认情况下，变量中的 CQ码会被转义以保证安全，如果不希望模板的结果触发 CQ码，请使用 `safe` 过滤器。

```jinja2
你发送的消息是：{{ event.message }}
CQ码解析后为：{{ event.message | safe }}
```

---

在模板变量中可以使用一些由 `gypsum` 提供的函数

```jinja2
{{ at(event.user_id) | safe }}，你好
```

这里 `at` 是一个函数，接受了发送者的 user_id 之后，将其包装成一个 `@发送者` 的消息节点

注意这里比如使用 `safe` 过滤器，将包装后的 CQ码 转化为实际发送的 `at`

所有可用的函数可参考 [模板文档 - 函数](https://github.com/yuudi/gypsum/wiki/Templating)

---

在消息触发方式中，还可以使用 `state` 对象

例如：使用前缀匹配方式中，可以获得 `state.args` 参数

```jinja2
为您找到的结果：
https://baidu.com/s?wd={{ url_encode(state.args) }}
```

每种触发方式产生的 `state` 对象不同，具体可参考 [模板文档 - 变量](https://github.com/yuudi/gypsum/wiki/Templating)

---

在模板中，可以使用条件、循环等控制语句

```jinja2
{% if event.user_id == 123456 %}
你好，主人
{% else %}
你好，欢迎来到 123456 的领域
{% endif %}
```

---

除了标准的 jinja 模板，gypsum 还提供了 `Lua 代码块` 标签，以供实现复杂逻辑

```lua
开始执行 Lua
{% lua %}
a = "你好"
b = "世界"
c = "控制台"
write(a..b)
print(a..c)
{% endlua %}
```

你可能注意到了，在 Lua 代码块中，使用 `print` 函数会将结果输出到控制台，使用 `write` 函数会将结果写入模板。
在 Lua 代码块中所有可用的函数可参考 [模板文档 - Lua 代码块](https://github.com/yuudi/gypsum/wiki/Lua)
