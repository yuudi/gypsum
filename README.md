# gypsum

石膏自定义

gypsum 是受到 [铃心自定义](http://myepk.club/) 的启发，基于 [ZeroBot](https://github.com/wdvxdr1123/ZeroBot) 的实现可视化控制台

[预览版](https://github.com/yuudi/gypsum/releases/latest)

![预览图](./imgs/preview.png)

交流 QQ 群：238627697

## 用法

**！！！警告：预览版本仅供测试，用于生产环境请谨慎**  
**！！！警告：预览版本接口尚未稳定，可能会进行不兼容更新**

1. 修改 `onebot` 的配置文件，启用`正向ws`
1. 启动一次 `gypsum`，生成 `gypsum_config.toml` 配置文件
1. 向 `gypsum_config.toml` 配置文件中填写`正向ws`连接参数、网页端口、账号、密码
1. 启动 `onebot` ，再启动 `gypsum`
1. 打开 `<你的ip地址>:9900`，开始使用

## todo

- [x] 接口鉴权
- [x] 通知事件
- [x] 定时任务
- [x] 暂停/启用
- [ ] 前端
    - [x] 简易前端
    - [ ] 用户友好的前端
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
- [ ] 分组
- [ ] 组导入导出
