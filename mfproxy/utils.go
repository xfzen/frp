package mfproxy

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
)

// 判断是否为 HTTP 请求
func getHTTPUrl(data []byte) (string, error) {
	// 使用 bufio.NewReader 从数据流中提取第一行
	reader := bufio.NewReader(bytes.NewReader(data))
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	// logx.Debugf("%v", string(data))

	// 检查是否是有效的HTTP请求
	requestParts := strings.Split(line, " ")
	if len(requestParts) < 3 {
		return "", errors.New("invalid HTTP request")
	}

	rawURL = requestParts[1]

	// 检查请求行是否包含 HTTP 方法（GET, POST 等）
	methods := []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "PATCH"}
	for _, method := range methods {
		if strings.HasPrefix(line, method) && strings.Contains(line, "HTTP/") {
			return rawURL, nil
		}
	}

	return "", errors.New("invalid HTTP request")
}

// getMFName 从 URL 中提取 mfname 查询参数的值
func getMFName(rawURL string) (string, error) {
	// logx.Debugf("getMFName: %v", rawURL)

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		logx.Errorf("getMFName url: %v, err: %v", rawURL, err)
		return "", err
	}

	queryParams := parsedURL.Query()

	mfname := queryParams.Get("mfname")
	// logx.Debugf("getMFName mfname: %v", mfname)
	if mfname == "" {
		return "", fmt.Errorf("mfname not found in URL")
	}

	return mfname, nil
}
