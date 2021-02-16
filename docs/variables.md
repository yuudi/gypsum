# 模板变量

## event

`event` 是收到的事件对象，具体结构可参照按照 [onebot标准](https://github.com/howmanybots/onebot/blob/master/v11/specs/event)  
只有事件触发的模板才能使用 `event`，定时任务中的模板是没有 `event` 对象的。

## state

对于各种匹配方式，state 给出了匹配结果

### 完全匹配

`state.matched` 为匹配的消息

### 关键词匹配

`state.keyword` 为匹配的关键词

### 前缀匹配

`state.prefix` 为匹配的前缀  
`state.args` 为除去前缀后的剩余部分

### 后缀匹配

`state.suffix` 为匹配的后缀  
`state.args` 为除去后缀后的剩余部分

### 命令匹配

`state.command` 为匹配的命令  
`state.args` 为除去命令后的剩余部分

### 正则匹配

`state.regex_matched` 为正则匹配结果数组

示例：

用正则表达式 `(\d+)d(\d+)` 匹配消息 `1d6` 时  
`state.regex_matched[0]` 为 `1d6`  
`state.regex_matched[1]` 为 `1`  
`state.regex_matched[2]` 为 `6`

> 注意：在 lua 中 state.regex_matched 的序号是从 1 开始的
