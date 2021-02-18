# 开发指南

此指南是针对 gypsum 源码修改的指南，gypsum 模板使用指南请查看 [模板说明](./docs/template.md)

## 前端

首先运行新版 gypsum 作为后端，然后启动前端开发服务器

```shell
# 拉取前端代码
git pull https://github.com/yuudi/gypsum-web.git

cd gypsum-web

# 安装依赖
yarn install
```

如果你的后端服务不是 `127.0.0.1:9900`，那么请修改 `vue.config.js` 中 `proxy` 字段以便正确代理 api 请求

```
# 启动前端开发服务器
yarn serve
```

## 后端

首先提取前端静态资源，然后构建后端

```shell
# 提取静态资源到当前目录
./gypsum extract-web .
```

此时在当前下获得一个 `web` 文件夹

```shell
# 拉取后端代码
git pull https://github.com/yuudi/gypsum.git
cd gypsum
```

将刚才的 `web` 文件夹放在 `gypsum/gypsum/web` 位置

```shell
# 编译并启动
mkdir dist
go build -o "dist/gypsum" .
cd dist
./gypsum
```
