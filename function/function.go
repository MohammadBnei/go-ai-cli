package function

import (
	"github.com/sashabaranov/go-openai"
)

type FunctionDefinition[T any] struct {
	Definition openai.FunctionDefinition
	Id         string
	Function   func(T) error
}
