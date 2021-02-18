# gypsum Cli

## 命令

### daemon

`gypsum daemon`

启动 gypsum 守护（默认方式）

### run

`gypsum run`

启动 gypsum 服务，此方法应当仅由守护进程使用 ，返回值为 `5` 时应当由守护进程重启

### init

`gypsum init`

初始化 gypsum 配置文件，若已存在则覆盖

### extract-web

`gypsum extract-web <path>`

提取 gypsum 内置网页文件到指定路径

### update

更新 gypsum

`gypsum update [<version>] [--mirror=<mirror_host>] [--force]`

参数： version 指定版本，可填 `stable` `beta` `v1.0.0`

选项： mirror 指定下载镜像（将替换 `github.com`），force 强制更新

示例：

```shell
gypsum update v1.0.0 --mirror="download.fastgit.org"
```
