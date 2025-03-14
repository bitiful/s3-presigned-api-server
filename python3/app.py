#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import boto3
import botocore
from flask import Flask, request, jsonify
from botocore.config import Config
import os
import logging
from urllib.parse import urlencode

# 配置常量
S3_ENDPOINT = "https://s3.bitiful.net"
BUCKET = os.environ.get("BUCKET", "")
AK = os.environ.get("AK", "")
SK = os.environ.get("SK", "")

app = Flask(__name__)

def get_s3_client(key, secret):
    """
    创建S3客户端并配置自定义端点
    """
    session = boto3.session.Session()
    return session.client(
        's3',
        region_name='cn-east-1',
        endpoint_url=S3_ENDPOINT,
        aws_access_key_id=key,
        aws_secret_access_key=secret
    )

def get_s3_resource(key, secret):
    """
    创建S3资源并配置自定义端点
    """
    session = boto3.session.Session()
    return session.resource(
        's3',
        region_name='cn-east-1',
        endpoint_url=S3_ENDPOINT,
        aws_access_key_id=key,
        aws_secret_access_key=secret
    )

@app.route('/presigned-url', methods=['GET'])
def presigned_url():
    """
    生成预签名URL
    请求示例: http://127.0.0.1:1998/presigned-url?key=tmp/test&content-length=231703
    """
    # 获取请求参数
    key = request.args.get('key', '')
    content_length_str = request.args.get('content-length', '')
    no_wait = request.args.get('no-wait', '')
    max_requests = request.args.get('max-requests', '')
    expire_seconds = request.args.get('expire', '')
    force_download = request.args.get('force-download', '').lower() in ['true', '1', 't', 'y', 'yes']
    limit_rate = request.args.get('limit-rate', '')

    # 验证key参数
    if not key:
        return '', 400

    # 验证content-length参数（如果提供）
    content_length = None
    if content_length_str:
        try:
            content_length = int(content_length_str)
            if content_length <= 0 or content_length > 1024 * 1024 * 1024:  # 大于0且不超过1GB
                return '', 400
        except ValueError:
            return '', 400

    # 创建S3客户端
    s3_client = get_s3_client(AK, SK)

    # 准备额外参数
    additional_params = {}

    # 开启 simul-transfer 即传即收
    if no_wait:
        try:
            no_wait_value = int(no_wait)
            if no_wait_value > 0:
                if no_wait_value > 10:
                    no_wait_value = 10
                additional_params['no-wait'] = str(no_wait_value)
        except ValueError:
            pass

    # 最大下载次数
    if max_requests:
        try:
            max_requests_value = int(max_requests)
            if max_requests_value > 0:
                additional_params['x-bitiful-max-requests'] = str(max_requests_value)
        except ValueError:
            pass

    # 单线程限速
    if limit_rate:
        try:
            limit_rate_value = int(limit_rate)
            if limit_rate_value > 0:
                additional_params['x-bitiful-limit-rate'] = str(limit_rate_value)
        except ValueError:
            pass

    # 强制下载
    if force_download:
        additional_params['response-content-disposition'] = 'attachment'

    # 设置过期时间
    expiration = 3600  # 默认1小时
    if expire_seconds:
        try:
            expire_seconds_value = int(expire_seconds)
            if expire_seconds_value > 0:
                expiration = expire_seconds_value
        except ValueError:
            pass

    # 生成GET预签名URL
    get_params = {
        'Bucket': BUCKET,
        'Key': key,
    }
    
    # 创建GET URL
    get_url = s3_client.generate_presigned_url(
        'get_object',
        Params=get_params,
        ExpiresIn=expiration
    )
    
    # 如果有额外参数，添加到URL
    if additional_params:
        # 解析现有URL并添加额外参数
        if '?' in get_url:
            base_url, query_string = get_url.split('?', 1)
            # 保留原有参数
            from urllib.parse import parse_qs
            existing_params = parse_qs(query_string)
            # 合并参数
            for key, value in additional_params.items():
                existing_params[key] = [value]
            # 重新编码参数
            new_query = urlencode(existing_params, doseq=True)
            get_url = f"{base_url}?{new_query}"
        else:
            get_url = f"{get_url}?{urlencode(additional_params)}"

    # 创建PUT URL
    put_params = {
        'Bucket': BUCKET,
        'Key': key,
    }
    
    # 只有当指定了content_length时才设置ContentLength字段
    if content_length is not None:
        put_params['ContentLength'] = content_length
    
    put_url = s3_client.generate_presigned_url(
        'put_object',
        Params=put_params,
        ExpiresIn=expiration
    )

    # 返回结果
    response = jsonify({
        'get-url': get_url,
        'put-url': put_url
    })
    
    # 设置CORS头
    response.headers.add('Access-Control-Allow-Origin', '*')
    response.headers.add('Access-Control-Allow-Methods', '*')
    response.headers.add('Access-Control-Allow-Headers', '*')
    
    return response

if __name__ == '__main__':
    # 验证环境变量
    if not BUCKET or not AK or not SK:
        logging.fatal("bucket, ak, sk should not be empty")
        exit(1)

    # 启动服务器
    logging.basicConfig(level=logging.INFO)
    logging.info("server started at :1998")
    app.run(host='0.0.0.0', port=1998, debug=False)
