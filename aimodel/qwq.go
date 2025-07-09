package aimodel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/openai/openai-go"
	"io"
	"net/http"
)

type QWQ struct {
	model string
	token string
}

func NewQwqModel(model, key string) *QWQ {
	return &QWQ{
		model: model,
		token: key,
	}
}

func (Q QWQ) Chat(ctx context.Context, chatList ChatList) (StringResult, error) {
	url := "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"

	data := map[string]interface{}{
		"model":    Q.model,
		"messages": chatList,
		"stream":   false,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return StringResult{}, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return StringResult{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+Q.token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return StringResult{}, err
	}
	defer resp.Body.Close()

	//if resp.StatusCode != http.StatusOK {
	//	fmt.Println("Request failed with status:", resp.Status)
	//	return StringResult{}, fmt.Errorf("request failed with status: %s", resp.Status)
	//}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return StringResult{}, err
	}

	fmt.Println("Response:", string(body))
	var result DeepSeekResponse
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("Error unmarshalling response:", err)
		return StringResult{}, err
	}
	if len(result.Choices) == 0 {
		return StringResult{}, fmt.Errorf("no choices in response")
	}
	res := StringResult{
		Result: result.Choices[0].Message.Content,
		Reason: result.Choices[0].Message.ReasonContent,
	}
	return res, nil
}

func (Q QWQ) ChatFlow(ctx context.Context, chatData ChatList) (chan string, chan string, chan error, error) {
	//TODO implement me
	panic("implement me")
}

func (Q QWQ) chatListToOpenAIChatMessages(chatList ChatList) []openai.ChatCompletionMessageParamUnion {
	messages := make([]openai.ChatCompletionMessageParamUnion, len(chatList))
	for i, chat := range chatList {
		switch chat.Role {
		case EAIChatRoleUser:
			if userContent, ok := chat.Content.([]UserContent); ok {
				var openUserContent []openai.ChatCompletionContentPartUnionParam
				for _, content := range userContent {
					switch content.Type {
					case "text":
						openUserContent = append(openUserContent, openai.TextContentPart(content.Text))
					case "image_url":
						openUserContent = append(openUserContent, openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
							URL: content.ImageUrl,
						}))
					}
				}
				messages[i] = openai.UserMessage(openUserContent)
			} else {
				messages[i] = openai.UserMessage(chat.Content.(string))
			}
		case EAIChatRoleAssistant:
			messages[i] = openai.AssistantMessage(chat.Content.(string))
		case EAIChatRoleSystem:
			messages[i] = openai.SystemMessage(chat.Content.(string))
		default:
			messages[i] = openai.UserMessage(chat.Content.(string))
		}
	}
	return messages
}
