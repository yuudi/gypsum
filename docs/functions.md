# 模板函数

## at

接受若干个 QQ 号，转化为 at，需要配合 `safe` 过滤器使用

参数：数字或字符串

用法示例：

```jinja
{{ at(2854196310) | safe }} 你好！
```

## at_sender

at 发送者，需要配合 `safe` 过滤器使用

参数：无

限制：仅限群聊消息、群聊事件

用法示例：

```jinja
{{ at_sender | safe }} 你好！
```

## approve

同意一个事件

参数：无

限制：仅限加好友请求、加群请求、加群邀请

用法示例：

```jinja
{% if event.user_id == 123456 %}
{{ approve }}
{% endif %}
```

## group_ban

群内禁言

参数：数字或字符串，

限制：仅限群聊消息、群聊事件

用法示例：

```jinja
{{ group_ban(60*5) }}
违反群规！禁言5分钟警告！
```

## image

接受一个图片网址，转化为图片发送，需要配合 `safe` 过滤器使用

参数：第一个参数为 uri 字符串。第二个参数为整数表示是否缓存，默认值 1

用法示例：

```jinja
{{ image("https://home.baidu.com/Public/img/logo.png") | safe }} 请使用百度
```

```jinja
{{ image("https://moebi.org/pic.php", 0) | safe }} 这是随机图片
```

## res

接受一个资源文件，转化为 uri，一般配合 image 使用  
在 file 模式下会转化为资源文件的绝对路径，在 http 模式下会生成为资源文件的网址

参数：字符串，一般在资源文件页面能找到

用法示例：

```jinja
{{ image(res("0123456789abcdef.jpg")) | safe }}
```

## sleep

等待一段时间

参数：数字或可转化为数字的字符串，单位为秒

用法示例：

```jinja
{{ sleep(10) }}
时间到！
```

## url_encode

将字符串进行 url 编码

参数：字符串

返回：字符串

用法示例：

```jinja
为您找到的结果：
https://baidu.com/s?wd={{ url_encode(state.args) }}
```

## random_int

获取随机整数

参数：0\~2 个整数，没有参数时范围为 0\~99，有一个参数 `a` 时范围为 0\~a，有两个参数 `a` `b` 时范围为 a\~b

用法示例：

```jinja
您的骰子点数为：{{ random_int(1, 6) }}
```

## db_put

向数据库中写一个值

参数：两个参数均为整数或字符串，第一个参数为键值，第二个参数为数据

用法示例：见下一部分

## db_get

从数据库中读一个值

参数：两个参数均为整数或字符串，第一个参数为键值，第二个参数为默认值（可选）

返回值：读取出的数据

用法示例：

```jinja
{% if db_get("setu", 0) > 5 %}
您今天使用次数太多了，请明天再来
{% else %}
这是您今天的结果
{{ db_put("setu", db_get("setu", 0)) }}
{% endif %}
```
