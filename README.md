# API Server For tldraw-no-wait-transfer-example

## 运行方式
### 方式一
#### Mac用户 执行 `BUCKET=<your_bucket> AK=<your-bitiful-s3-accesskey> SK=<your-bitiful-s3-secretkey> ./api-server-mac`
#### Windows用户 执行 `BUCKET=<your_bucket> AK=<your-bitiful-s3-accesskey> SK=<your-bitiful-s3-secretkey> ./api-server-win.exe`

### 方式二 (如果你本地已经安装了golang的开发环境)
#### 1. `go mod tidy && go mod vendor`
#### 2. `BUCKET=<your_bucket> AK=<your-bitiful-s3-accesskey> SK=<your-bitiful-s3-secretkey> go run main.go`

--- 

### *编译命令参考*
#### *编译为 Mac 可执行文件*
*GOOS=darwin GOARCH=amd64 go build -o api-server-mac main.go*
#### *编译为 Windows 可执行文件*
*GOOS=windows GOARCH=amd64 go build -o api-server-win.exe main.go*