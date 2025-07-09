package htmlHandler

import (
	md "github.com/JohannesKaufmann/html-to-markdown"
	"strings"
)

func ToMarkdown(html string) (string, error) {
	// 1. 首先清理HTML（使用之前优化的cleanHTML函数）
	cleanedHTML, err := CleanHTML(strings.NewReader(html))
	if err != nil {
		return "", err
	}

	// 2. 创建Markdown转换器
	converter := md.NewConverter("", true, nil)

	// 4. 转换清理后的HTML为Markdown
	return converter.ConvertString(cleanedHTML)
}
