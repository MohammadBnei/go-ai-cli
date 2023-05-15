# Go-OpenAI-CLI

Go-OpenAI-CLI is a command-line interface that provides access to OpenAI's GPT-3 language generation service. With this tool, users can send a prompt to the OpenAI API and receive a generated response, which can then be printed on the command-line or saved to a markdown file. 

This project is useful for quickly generating text for various purposes such as creative writing, chatbots, virtual assistants, or content generation for websites. 

## Installation

To install and use the Go-OpenAI-CLI with golang :

```sh
go install github.com/MohammadBnei/go-openai-cli@latest
```

To install the compiled binaries, go to the the [release page](https://github.com/MohammadBnei/go-openai-cli/releases/) and select the exec matching your operating system. 

## Usage

First, set up your OpenAI API key and default model by running:
```
go-openai-cli config --OPENAI_KEY=<YOUR_API_KEY> --model=<DEFAULT_MODEL>
```

To get a list of available models, run:
```
go-openai-cli config -l
```

To send a prompt to OpenAI GPT, run:
```
go-openai-cli prompt
```

You will be prompted to enter your text prompt. After submitting your prompt, OpenAI will process your input and generate a response.

### Command in prompt 
```
q: quit
h: help
s: save the response to a file
f: add a file to the messages (won't send to openAi until you send a prompt)
c: clear message list
```

## Configuration

To store your OpenAI API key and model, run the following command:
```
go-openai-cli config --OPENAI_KEY=<YOUR_API_KEY> --model=<MODEL>
```

### Flags
- `--OPENAI_KEY`: Your OpenAI API key.
- `--model`: The default model to use.
- `-l, --list-model`: List available models.
- `--config`: The config file location.

The configuration file is located in `$HOME/.go-openai-cli.yaml`.

## Contributing

To contribute to this project, fork the repository, make your changes, and submit a pull request. Please also ensure that your code adheres to the accepted [Go style guide](https://golang.org/doc/effective_go.html). 

## License

This project is licensed under the [MIT License](https://opensource.org/licenses/MIT).