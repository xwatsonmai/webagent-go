package aimodel

import "github.com/pkg/errors"

func Builder(modelName string, think bool, key string) (IAiModel, error) {
	switch modelName {
	case "deepseek":
		if think {
			return NewDeepSeekWithModel("deepseek-reasoner", key), nil
		}
		return NewDeepSeekWithModel("deepseek-chat", key), nil
	case "qwq":
		return NewQwqModel("qwen-max-latest", key), nil
	case "doubao":
		model := "doubao-seed-1-6-250615"
		if think {
			model = "doubao-seed-1.6-thinking"
		}
		return NewDoubao(model, key), nil
	}
	return nil, errors.New("unsupported model type: " + modelName)
}
