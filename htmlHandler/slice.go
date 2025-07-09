package htmlHandler

func Slice(html string) ([]string, int) {
	var result []string
	runes := []rune(html)
	length := len(runes)
	const chunkSize = 15000
	const overlap = 200

	for start := 0; start < length; {
		end := start + chunkSize
		if end > length {
			end = length
		}
		chunk := runes[start:end]
		result = append(result, string(chunk))
		if end == length {
			break
		}
		// 下一个切片起点往前 overlap 个字符
		start = end - overlap
		if start < 0 {
			start = 0
		}
	}
	return result, len(result)
}
