package function

import (
	"github.com/MohammadBnei/go-openai-cli/tool"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type SaveFunctionData struct {
	Content       string
	Filename      string
	WithTimestamp bool
}

var SaveFileFunctionDef = FunctionDefinition[*SaveFunctionData]{
	Id: "saveFile",
	Function: func(data *SaveFunctionData) error {
		return tool.SaveToFile([]byte(data.Content), data.Filename)
	},
	Definition: openai.FunctionDefinition{
		Name:        "saveFile",
		Description: `Saves the provided content to a file."`,
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"data": {
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"content": {
							Type:        jsonschema.String,
							Description: "The content to be saved to the file.",
						},
						"filename": {
							Type:        jsonschema.String,
							Description: "The name of the file, including its path, to save the content to. If empty, the user will be prompted to provide one.",
						},
						"withTimestamp": {
							Type:        jsonschema.Boolean,
							Description: "If true, prepends the current timestamp to the file content.",
						},
					},
					Required: []string{"content"},
				},
			},
		},
	},
}
