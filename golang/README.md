# S3 获取预签名接口 Server - golang version

## 运行方式
## 1. 直接使用二进制文件运行
### macOS - Apple Silicon
执行
```shell
BUCKET=<your_bucket> AK=<your-bitiful-s3-accesskey> SK=<your-bitiful-s3-secretkey> ./api-server-mac
```

### Windows
执行
```shell
BUCKET=<your_bucket> AK=<your-bitiful-s3-accesskey> SK=<your-bitiful-s3-secretkey> ./api-server-win.exe
```

## 2. 使用 go 编译运行
```shell
go mod tidy && go mod vendor`
BUCKET=<your_bucket> AK=<your-bitiful-s3-accesskey> SK=<your-bitiful-s3-secretkey> go run main.go
```

--- 

### 其他编译命令参考
#### x86 架构 macOS 下编译
```shell
GOOS=darwin GOARCH=amd64 go build -o api-server-mac main.go
```

#### 编译为 Windows 可执行文件
```shell
GOOS=windows GOARCH=amd64 go build -o api-server-win.exe main.go
```
