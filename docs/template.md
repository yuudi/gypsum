# 模板语法

目录：

[简介](#简介)  
[模板变量](#模板变量)  
[模板函数](#模板函数)  
[模板标签](#模板标签)

## 简介

gypsum 的模板语法使用 [pongo2](https://github.com/flosch/pongo2)，语法类似于 `jinja2` 和 `Django`

学习资料：[jinja2 模板文档](http://docs.jinkan.org/docs/jinja2/templates.html)

---

你可以在模板中直接编写回复

```jinja
你好
```

---

你可以在模板中可以直接使用 [CQ 码](https://github.com/howmanybots/onebot/blob/master/v11/specs/message/string.md#cq-%E7%A0%81%E6%A0%BC%E5%BC%8F)

```jinja
[CQ:at,qq=all] 大家好
```

---

也可以用 `{{` `}}` 符号访问变量

```jinja
你好，{{ event.sender.nickname }}
```

其中：`event` 是收到的事件对象，具体结构可参照按照 [onebot 标准](https://github.com/howmanybots/onebot/blob/master/v11/specs/event)。  
注意只有事件触发的模板才能使用 `event`，定时任务中的模板是没有 `event` 对象的。

---

默认情况下，变量中的 CQ 码会被转义以保证安全，如果不希望模板的结果触发 CQ 码，请使用 `safe` 过滤器。

```jinja
你发送的消息是：{{ event.message }}
CQ码解析后为：{{ event.message | safe }}
```

---

在模板变量中可以使用一些由 `gypsum` 提供的函数

```jinja
{{ at(event.user_id) }}，你好
```

这里 `at` 是一个函数，接受了发送者的 user_id 之后，将其包装成一个 `@发送者` 的消息节点

> `at` 函数返回的值具有 `safe` 属性，无需 `safe` 过滤器

所有可用的函数可参考 [模板文档 - 函数](#模板函数)

---

在消息触发方式中，还可以使用 `state` 对象

例如：使用前缀匹配方式中，可以获得 `state.args` 参数

```jinja
为您找到的结果：
https://baidu.com/s?wd={{ url_encode(state.args) }}
```

每种触发方式产生的 `state` 对象不同，具体可参考 [模板文档 - 变量](#模板变量)

---

在模板中，可以使用条件、循环等控制语句

```jinja
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
在 Lua 代码块中所有可用的函数可参考 [模板文档 - Lua 代码块](./lua.md)

## 常用技巧

### 赋值

```jinja
{% set some_value = db_get("key", 0) %}
value is {{ some_value }}
```

### 判断

```jinja
{% if event.user_id == 123456 %}
你是主人
{% else %}
你不是主人
{% endif %}
```

### 循环

```jinja
4个骰子的点数分别为：
{%- for i in "1234" %}
第 {{ i }} 个：{{ random_int(1, 6) }}
{%- endfor %}
```

### 字符串包含

```jinja
{% if "宫廷玉液酒" in event.message %}
口令正确
{% else %}
口令错误
{% endif %}
```

### 裁剪空白符

在标签内侧加上一个 `-` 符号可以裁剪多余的空格或换行

```jinja
你是
{{- event.sender.nickname }}
```

> 标签和减号之间不能有空格

### 注释

```jinja
{# 这里是单行注释 #}

{% comment %}
这里是
多行注释
{% endcomment %}
```

## 模板变量

### event

`event` 是收到的事件对象，具体结构可参照按照 [onebot 标准](https://github.com/howmanybots/onebot/blob/master/v11/specs/event)  
只有事件触发的模板才能使用 `event`，定时任务中的模板是没有 `event` 对象的。

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
`state.regex_matched[0]` 为 `1d6`  
`state.regex_matched[1]` 为 `1`  
`state.regex_matched[2]` 为 `6`

> 注意：在 lua 中 state.regex_matched 的序号是从 1 开始的

## 模板函数

### at

接受若干个 QQ 号，转化为 at

参数：数字或字符串

用法示例：

```jinja
{{ at(2854196310) }} 你好！
```

### at_sender

at 发送者

参数：无

限制：仅限群聊消息、群聊事件

用法示例：

```jinja
{{ at_sender }} 你好！
```

### approve

同意一个事件

参数：无

限制：仅限加好友请求、加群请求、加群邀请

用法示例：

```jinja
{% if event.user_id == 123456 %}
{{ approve }}
{% endif %}
```

### withdraw

撤回消息

参数：无

限制：仅限群聊消息，需要 bot 是管理员

用法示例：

```jinja
{% if "广告" in event.message %}
{{ withdraw }}
禁止发广告
{% endif %}
```

### set_title

设置群头衔

参数：第一个参数是字符串，表示头衔。后续参数是数字，表示 qq 号，省略则使用发送者的 qq。

限制：仅限群聊消息、群聊事件，需要 bot 是群主

用法示例：

```jinja
{{ set_title("大佬") }}
你太厉害了，送给你“大佬”头衔
```

### group_ban

群内禁言

参数：第一个参数是数字，表示时长，0 表示解除。后续参数是数字，表示 qq 号，省略则使用发送者的 qq。

限制：仅限群聊消息、群聊事件，需要 bot 是管理员

用法示例：

```jinja
{{ group_ban(60*5) }}
违反群规！禁言5分钟警告！
```

### image

接受一个图片文件地址或网址，转化为图片发送

参数：第一个参数为 uri 字符串。第二个参数为整数表示是否缓存，默认值 1

用法示例：

```jinja
{{ image("https://home.baidu.com/Public/img/logo.png") }} 请使用百度
```

```jinja
{{ image("https://moebi.org/pic.php", 0) }} 这是随机图片
```

### record

语音，用法同 image

### res

接受一个资源文件，转化为 uri，一般配合 image 使用  
在 file 模式下会转化为资源文件的绝对路径，在 http 模式下会生成为资源文件的网址

参数：字符串，一般在资源文件页面能找到

用法示例：

```jinja
{{ image(res("0123456789abcdef.jpg")) }}
```

### sleep

等待一段时间

参数：数字或可转化为数字的字符串，单位为秒

用法示例：

```jinja
{{ sleep(10) }}
时间到！
```

### url_encode

将字符串进行 url 编码

参数：字符串

返回：字符串

用法示例：

```jinja
为您找到的结果：
https://baidu.com/s?wd={{ url_encode(state.args) }}
```

### file_get_contents

读取文件内容

参数：字符串，文件路径或网址

返回：字符串

用法示例：

```jinja
请朗读并背诵全文：

{{ file_get_contents("article.txt") }}
```

### random_line

随机取一行

参数：字符串

返回：字符串

用法示例：

```jinja
{{ random_line(file_get_contents("saved_reply.txt")) }}
```

### random_file

从文件夹中随机取一个文件

参数：字符串，文件夹路径

返回：字符串，文件的路径

用法示例：

```jinja
{{ image(random_file("/home/me/setu/")) }}
```

### parse_json

解析 json 字符串

返回：任何

用法示例：

```jinja
他的名称是：
{{- parse_json(file_get_contents("infomation.json")).name }}
```

### random_int

获取随机整数

参数：0\~2 个整数，没有参数时范围为 0\~99，有一个参数 `a` 时范围为 0\~a，有两个参数 `a` `b` 时范围为 a\~b

用法示例：

```jinja
您的骰子点数为：{{ random_int(1, 6) }}
```

### db_put

向数据库中写一个值

参数：两个参数均为整数或字符串，第一个参数为键值，第二个参数为数据

用法示例：见下一部分

### db_get

从数据库中读一个值

参数：两个参数均为整数或字符串，第一个参数为键值，第二个参数为默认值（可选）

返回值：读取出的数据

用法示例：

```jinja
{% set times = db_get("usage", 0) + 1 %}
{% if times > 3 %}
您今天使用次数太多了，请明天再来
{% else %}
这是您今天的次数：{{ times }}
{{ db_put("usage", times) }}
{% endif %}
```

## 模板标签

### comment

注释

用法示例：

```jinja
你好
{% comment %}
这里是注释，不会执行
{% endcomment %}
```

### random_choice

随机选择一个块

用法示例：

```jinja
{% random_choice %}
欢迎来到旅店
{% otherwise %}
快进来歇歇脚吧
{% otherwise %}
外面可真冷
{% end_random_choice %}
```

### send_private

发送私聊消息

参数：数字，表示 QQ 号

用法示例：

```jinja
{% send_private 123456 %}
收到来自{{ event.sender.nickname }}的反馈：
{{ state.args }}
{% end_send %}
您的反馈已发送给主人，感谢支持
```

### send_group

发送群聊消息

同上，略
