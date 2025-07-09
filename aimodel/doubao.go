package aimodel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"io"
	"net/http"
)

type Doubao struct {
	token string
	model string
}

func NewDoubao(model, token string) *Doubao {
	return &Doubao{
		model: model,
		token: token,
	}
}

func (d Doubao) Chat(ctx context.Context, chatList ChatList) (StringResult, error) {
	url := "https://ark.cn-beijing.volces.com/api/v3/chat/completions"
	token := d.token

	// 构造messages
	var messages []map[string]interface{}
	for _, chat := range chatList {
		msg := map[string]interface{}{
			"role":    chat.Role,
			"content": chat.Content,
		}
		messages = append(messages, msg)
	}

	body := map[string]interface{}{
		"model":    d.model,
		"messages": messages,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return StringResult{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return StringResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return StringResult{}, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return StringResult{}, err
	}
	var response DeepSeekResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return StringResult{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	if len(response.Choices) == 0 {
		g.Log().Errorf(ctx, "Doubao response has no choices: %s", string(respBody))
		return StringResult{}, fmt.Errorf("no choices in response")
	}
	// 直接返回字符串结果
	return StringResult{
		ID:     "",
		Result: response.Choices[0].Message.Content,
		Reason: response.Choices[0].Message.ReasonContent,
	}, nil
}

func (d Doubao) ChatFlow(ctx context.Context, chatData ChatList) (chan string, chan string, chan error, error) {
	//TODO implement me
	panic("implement me")
}
