package aimodel

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type DeepSeekResponse struct {
	Id      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role          string `json:"role"`
			Content       string `json:"content"`
			ReasonContent string `json:"reason_content,omitempty"` // Reason content is optional
		} `json:"message"`
		Logprobs     interface{} `json:"logprobs"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens        int `json:"prompt_tokens"`
		CompletionTokens    int `json:"completion_tokens"`
		TotalTokens         int `json:"total_tokens"`
		PromptTokensDetails struct {
			CachedTokens int `json:"cached_tokens"`
		} `json:"prompt_tokens_details"`
		PromptCacheHitTokens  int `json:"prompt_cache_hit_tokens"`
		PromptCacheMissTokens int `json:"prompt_cache_miss_tokens"`
	} `json:"usage"`
	SystemFingerprint string `json:"system_fingerprint"`
}

type DeepSeekFlowResponse struct {
	Id                string `json:"id"`
	Object            string `json:"object"`
	Created           int    `json:"created"`
	Model             string `json:"model"`
	SystemFingerprint string `json:"system_fingerprint"`
	Choices           []struct {
		Index int `json:"index"`
		Delta struct {
			Content          string `json:"content"`
			ReasoningContent string `json:"reasoning_content"`
		} `json:"delta"`
		Logprobs     interface{} `json:"logprobs"`
		FinishReason interface{} `json:"finish_reason"`
	} `json:"choices"`
}

type DeepSeek struct {
	model  string
	apiKey string
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

//func NewDeepSeekModel() IAiModel {
//	apiToken, _ := g.Cfg().Get(context.Background(), "model_token.deepseek_token")
//	return DeepSeek{
//		model:  "deepseek-reasoner",
//		apiKey: apiToken.String(),
//	}
//}

func NewDeepSeekWithModel(model, apiKey string) IAiModel {
	return DeepSeek{
		model:  model,
		apiKey: apiKey,
	}
}

func (d DeepSeek) Chat(ctx context.Context, chatList ChatList) (StringResult, error) {
	url := "https://api.deepseek.com/chat/completions"

	data := map[string]interface{}{
		"model":       d.model,
		"messages":    chatList,
		"stream":      false,
		"temperature": 0.6,
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
	req.Header.Set("Authorization", "Bearer "+d.apiKey)

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

func (d DeepSeek) ChatFlow(ctx context.Context, chatList ChatList) (chan string, chan string, chan error, error) {
	url := "https://api.deepseek.com/chat/completions"

	data := map[string]interface{}{
		"model":       d.model,
		"messages":    chatList,
		"stream":      true,
		"temperature": 0.6,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, nil, nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, nil, nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, nil, nil, fmt.Errorf("request failed with status: %s", resp.Status)
	}

	out := make(chan string)
	errCh := make(chan error)
	think := make(chan string, 1) // Channel to send reasoning content
	go func() {
		defer resp.Body.Close()
		defer close(out)
		defer close(errCh)
		defer close(think)

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "data: ") {
				line = strings.TrimPrefix(line, "data: ")
				if line == "[DONE]" {
					out <- line
					break
				}
				var response DeepSeekFlowResponse
				if err := json.Unmarshal([]byte(line), &response); err != nil {
					errCh <- err
					return
				}
				if len(response.Choices) > 0 {
					if response.Choices[0].Delta.Content != "" {
						out <- response.Choices[0].Delta.Content
					} else if response.Choices[0].Delta.ReasoningContent != "" {
						think <- response.Choices[0].Delta.ReasoningContent
					}

				}
			}
		}
		if err := scanner.Err(); err != nil {
			errCh <- err
			return
		}
	}()
	return out, think, errCh, nil
}
