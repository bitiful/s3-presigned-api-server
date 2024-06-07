package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const s3Endpoint = "https://s3.bitiful.net"

var bucket = ""
var ak = ""
var sk = ""

func getS3Client(key, secret string) *s3.Client {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if service == "S3" {
			return aws.Endpoint{
				URL: s3Endpoint,
			}, nil
		}
		return aws.Endpoint{}, fmt.Errorf("unknown service requested")
	})
	customProvider := credentials.NewStaticCredentialsProvider(key, secret, "")
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithCredentialsProvider(customProvider), config.WithEndpointResolverWithOptions(customResolver))
	if err != nil {
		return nil
	}
	cfg.Region = "cn-east-1"
	s3client := s3.NewFromConfig(cfg)
	return s3client
}

func main() {
	bucket = os.Getenv("BUCKET")
	ak = os.Getenv("AK")
	sk = os.Getenv("SK")

	if bucket == "" || ak == "" || sk == "" {
		log.Fatal("bucket, ak, sk should not be empty")
		return
	}
	// 定义路由处理函数
	http.HandleFunc("/presigned-url", presignedUrl)
	// 启动 HTTP 服务器
	log.Printf("server started at :1998")
	_ = http.ListenAndServe(":1998", nil)
}

func presignedUrl(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	contentLength, _ := strconv.ParseInt(r.URL.Query().Get("content-length"), 10, 64)
	noWait, _ := strconv.ParseInt(r.URL.Query().Get("no-wait"), 10, 64)
	maxRetries, _ := strconv.ParseInt(r.URL.Query().Get("max-retries"), 10, 64)

	if key == "" || contentLength <= 0 || contentLength > 1024*1024*1024 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	client := getS3Client(ak, sk)
	presignClient := s3.NewPresignClient(client)

	// gen presigned url include custom params
	additionalParams := map[string]string{} // no-wait等待上传的超时时间 (单位：秒)
	if noWait > 0 {
		if noWait > 10 {
			noWait = 10
		}
		additionalParams["no-wait"] = fmt.Sprintf("%d", noWait)
	}
	if maxRetries > 0 {
		additionalParams["x-bitiful-max-retries"] = fmt.Sprintf("%d", maxRetries)
	}

	ctx := context.TODO()
	if len(additionalParams) > 0 {
		ctx = context.WithValue(ctx, "bitiful-additional-params", additionalParams)
	}
	getObjectReq, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, func(presignOptions *s3.PresignOptions) {
		presignOptions.Expires = time.Hour // 有效期1小时
		presignOptions.ClientOptions = append(presignOptions.ClientOptions, func(options *s3.Options) {
			options.APIOptions = append(options.APIOptions, RegisterPresignedUrlAddParamsMiddleware)
		})
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// put url
	putObjectReq, err := presignClient.PresignPutObject(context.Background(), &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(key),
		ContentLength: aws.Int64(contentLength),
	}, func(presignOptions *s3.PresignOptions) {
		presignOptions.Expires = time.Hour
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, _ := json.Marshal(map[string]string{
		"get-url": getObjectReq.URL,
		"put-url": putObjectReq.URL,
	})

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}

// 获取预签名url增加自定义参数 参考链接 https://aws.github.io/aws-sdk-go-v2/docs/middleware/
func RegisterPresignedUrlAddParamsMiddleware(stack *middleware.Stack) error {
	return stack.Build.Add(presignedUrlAddParamsMiddleware, middleware.After)
}

var presignedUrlAddParamsMiddleware = middleware.BuildMiddlewareFunc("Bitiful:PresignedUrlAddParams", func(ctx context.Context, input middleware.BuildInput, next middleware.BuildHandler) (out middleware.BuildOutput, metadata middleware.Metadata, err error) {
	bitifulAdditionalParams := ctx.Value("bitiful-additional-params")
	if bitifulAdditionalParams == nil {
		return next.HandleBuild(ctx, input)
	}

	req, ok := input.Request.(*smithyhttp.Request)
	if !ok {
		return out, metadata, fmt.Errorf("unknown transport type %T", req)
	}

	bitifulAdditionalParamsMap, ok := bitifulAdditionalParams.(map[string]string)
	if !ok {
		return next.HandleBuild(ctx, input)
	}

	query := req.URL.Query()
	for key, value := range bitifulAdditionalParamsMap {
		query.Set(key, value)
	}
	req.URL.RawQuery = query.Encode()

	return next.HandleBuild(ctx, input)
})
