package service

import (
	"encoding/json"

	"github.com/sashabaranov/go-openai"
)

var ExempleFunctionDef = &openai.FunctionDefinition{
	Name: "Exemple function",

	Description: "This is an example function",
	Parameters: json.RawMessage(`{
		"type": "object",
		"properties": {
				"location": {
						"type": "string",
						"description": "The city and state, e.g. San Francisco, CA",
				},
				"format": {
						"type": "string",
						"enum": ["celsius", "fahrenheit"],
						"description": "The temperature unit to use. Infer this from the users location.",
				},
				"num_days": {
						"type": "integer",
						"description": "The number of days to forecast",
				}
		},
		"required": ["location", "format", "num_days"]
	}`),
}
