package aimodel

type UserPromptType interface {
	string | UserContent | []UserContent
}

func UserPrompt[T UserPromptType](prompt T) ChatList {
	return ChatList{
		{
			Role:    EAIChatRoleUser,
			Content: prompt,
		},
	}
}

func SystemUserPrompt[T UserPromptType](system string, userContent T) ChatList {
	return ChatList{
		{
			Role:    EAIChatRoleSystem,
			Content: system,
		},
		{
			Role:    EAIChatRoleUser,
			Content: userContent,
		},
	}
}
