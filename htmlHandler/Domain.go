package htmlHandler

import (
	"github.com/playwright-community/playwright-go"
	"net/url"
	"strings"
)

func GetTopLevelDomain(page playwright.Page) (string, error) {
	// 获取页面的 URL
	pageURL := page.URL()
	parsedURL, err := url.Parse(pageURL)
	if err != nil {
		return "", err
	}

	// 提取主机名
	host := parsedURL.Host

	// 分割主机名并提取一级域名
	parts := strings.Split(host, ".")
	if len(parts) < 2 {
		return host, nil // 如果主机名格式不正确，直接返回
	}
	topLevelDomain := strings.Join(parts[len(parts)-2:], ".")
	return topLevelDomain, nil
}
