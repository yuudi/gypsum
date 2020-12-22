# gypsum

石膏自定义

gypsum 是受到 [铃心自定义](http://myepk.club/) 的启发，基于 [ZeroBot](https://github.com/wdvxdr1123/ZeroBot) 的实现可视化控制台

![预览图](./imgs/preview.png)

## 用法

**！！！警告：目前版本仅供测试，用于生产环境请谨慎**

1. 修改 `onebot` 的配置文件，启用`正向ws`
1. 启动一次 `gypsum`，生成 `gypsum_config.toml` 配置文件
1. 向 `gypsum_config.toml` 配置文件中填写
1. 启动 `onebot` ，再启动 `gypsum`
1. 打开 `<你的ip地址>:9900`，开始使用

## todo

- [x] 修改删除规则
- [x] 鉴权
- [ ] 通知事件
- [ ] 定时器
- [ ] 写前端
- [ ] 接口文档
- [ ] 自动更新
- [ ] 更强大的回复模板
- [ ] 模板文档
