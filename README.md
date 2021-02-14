# gypsum

石膏自定义

交流 QQ 群：238627697

gypsum 是受到 [铃心自定义](http://myepk.club/) 的启发，基于 [ZeroBot](https://github.com/wdvxdr1123/ZeroBot) 的实现可视化控制台

[预览版](https://github.com/yuudi/gypsum/releases/latest)

![预览图](./docs/preview.png)

## 用法

**！！！警告：预览版本仅供测试，用于生产环境请谨慎**  
**！！！警告：预览版本接口尚未稳定，可能会进行不兼容更新**

gypsym 需要配合 onebot 使用，例如：[go-cqhttp](https://go-cqhttp.org/)、[onebot-mirai](https://github.com/yyuueexxiinngg/onebot-kotlin)、[node-onebot](https://github.com/takayama-lily/node-onebot)、[XQ-HTTP](https://discourse.xianqubot.com/t/topic/50)等

1. 修改 `onebot` 的配置文件，启用`正向ws`
1. 启动一次 `gypsum`，生成 `gypsum_config.toml` 配置文件
1. 向 `gypsum_config.toml` 配置文件中填写`正向ws`连接参数、网页端口、账号、密码
1. 启动 `onebot` ，再启动 `gypsum`
1. 打开 `<你的ip地址>:9900`，开始使用

### Docker

```shell
curl -sL https://github.com/yuudi/gypsum/raw/master/scripts/download.Dockerfile | docker build -t gypsum:latest -f - .
docker run --rm -v ${PWD}/gypsum:/gypsum gypsum
# 修改 gypsum/gypsum_data/gypsum_config.toml 文件后
docker rum -d -v ${PWD}/gypsum:/gypsum --name gypsum gypsum
# 最好同时将 gypsum 目录挂载至 gocqhttp 容器，以便共享文件
```

## todo

### 1.0

- [x] 接口鉴权
- [x] 通知事件
- [x] 定时任务
- [x] 暂停/启用
- [x] 前端
  - [x] 用户友好的前端
- [x] 静态资源上传
- [ ] 程序自动更新
- [x] 回复模板
  - [x] 更强大的回复模板
  - [x] 模板中使用 Lua
    - [x] Lua 调用 bot API
    - [ ] Lua 访问 KV 数据库
    - [ ] Lua 发起网络请求
  - [ ] 模板文档
  - [ ] 模板测试
- [ ] 内置频率控制器
- [ ] 内置积分系统
- [x] 分组
- [x] 组导入导出

### 1.1+

- [ ] 更强大的前端编辑器