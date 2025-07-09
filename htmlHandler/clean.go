package htmlHandler

import (
	"bytes"
	"golang.org/x/net/html"
	"io"
	"strings"
)

func CleanHTML(r io.Reader) (string, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return "", err
	}

	// 扩展标签黑名单（移除整个节点）
	removeTags := map[string]bool{
		"script": true, "style": true, "meta": true, "head": true,
		"link": true, "iframe": true, "svg": true, "picture": true,
		"video": true, "audio": true, "source": true, "track": true,
		"canvas": true, "map": true, "object": true, "embed": true,
		"applet": true, "frame": true, "frameset": true, "noframes": true,
		"noscript": true, "template": true, "path": true, "datalist": true,
	}

	// 需要保留内容但移除标签属性的元素
	stripAttributes := map[string]bool{
		"img":      true, // 保留标签但处理属性
		"input":    true, // 保留输入框但精简属性
		"button":   true, // 保留按钮但精简属性
		"form":     true, // 保留表单但精简属性
		"select":   true, // 保留下拉框
		"option":   true, // 保留选项
		"textarea": true, // 保留文本框
	}

	// 允许保留的HTML属性
	allowedAttributes := map[string]bool{
		"name": true, "type": true, "value": true,
		"placeholder": true, "alt": true, "title": true, "selected": true,
		"checked": true, "disabled": true, "readonly": true, "multiple": true,
		"id": true, "class": true,
	}

	var clean func(*html.Node)
	clean = func(n *html.Node) {
		// 从后向前遍历子节点（避免删除时指针问题）
		children := make([]*html.Node, 0)
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			children = append(children, c)
		}

		for i := len(children) - 1; i >= 0; i-- {
			c := children[i]
			switch {
			// 删除黑名单标签
			case c.Type == html.ElementNode && removeTags[c.Data]:
				n.RemoveChild(c)

			// 删除注释
			case c.Type == html.CommentNode:
				n.RemoveChild(c)

			// 处理需要精简属性的标签
			case c.Type == html.ElementNode && stripAttributes[c.Data]:
				// 特殊处理图片：清空src且保留alt
				if c.Data == "img" {
					newAttrs := []html.Attribute{}
					for _, attr := range c.Attr {
						if attr.Key == "alt" || attr.Key == "title" {
							newAttrs = append(newAttrs, attr)
						}
					}
					c.Attr = newAttrs
				} else {
					// 其他标签保留允许的属性
					newAttrs := []html.Attribute{}
					for _, attr := range c.Attr {
						if allowedAttributes[attr.Key] {
							newAttrs = append(newAttrs, attr)
						}
					}
					c.Attr = newAttrs
				}
				clean(c) // 递归处理子节点

			// 处理普通元素
			case c.Type == html.ElementNode:
				// 移除所有非白名单属性
				newAttrs := []html.Attribute{}
				for _, attr := range c.Attr {
					if allowedAttributes[attr.Key] {
						newAttrs = append(newAttrs, attr)
					}
				}
				c.Attr = newAttrs
				clean(c) // 递归处理子节点

			// 保留文本节点
			case c.Type == html.TextNode:
				// 可选：压缩连续空白字符
				c.Data = strings.Join(strings.Fields(c.Data), " ")

			// 其他类型节点直接删除
			default:
				n.RemoveChild(c)
			}
		}
	}

	clean(doc)

	// 序列化时跳过无关内容
	var buf bytes.Buffer
	html.Render(&buf, doc)
	return buf.String(), nil
}
