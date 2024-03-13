# go-ai-cli

`go-ai-cli` is a versatile command-line interface that enables users to interact with various AI models for text generation, speech-to-text conversion, image generation, and web scraping. It is designed to be extensible with any AI service. This tool is ideal for developers, content creators, and anyone interested in leveraging AI capabilities directly from their terminal. This tool works well with Ollama.

## Features

- **Text Generation**: Utilize OpenAI's GPT-3 for generating text based on prompts.
- **Speech to Text**: Convert spoken language into text.
- **Image Generation**: Create images from textual descriptions.
- **Modular Design**: Easily extendable to incorporate additional AI models and services.

## Installation

### Using Go

```sh
go install github.com/MohammadBnei/go-ai-cli@latest
```

with portaudio (recommended) :

```sh
go install -tags portaudio github.com/MohammadBnei/go-ai-cli@latest
```


### Pre-compiled Binaries

Download the appropriate binary for your operating system from the [releases page](https://github.com/MohammadBnei/go-ai-cli/releases/).

## Configuration

Before using `go-ai-cli`, configure it with your OpenAI API key:

```sh
go-ai-cli config --OPENAI_KEY=<YOUR_API_KEY>
```

You can also specify the AI model to use:

```sh
go-ai-cli config --model=<MODEL_NAME>
```

To list available models:

```sh
go-ai-cli config -l
```

## Usage

To start the interactive prompt:

```sh
go-ai-cli prompt
```

Within the prompt, you have several commands at your disposal:

- `ctrl+d`: Quit the application.
- `ctrl+h`: Display help information.
- `ctrl+g`: Open options page.
- `ctrl+f`: Add a file to the messages. The file content won't be sent to the model until you submit a prompt.

## Advanced Configuration

The configuration file is located at `$HOME/.go-ai-cli.yaml`. You can customize various settings, including the default AI model and API keys for different services.

## Contributing

Contributions are welcome! Please fork the repository, make your changes, and submit a pull request. Ensure your code follows the [Go style guide](https://golang.org/doc/effective_go.html).

## License

`go-ai-cli` is open-source software licensed under the [MIT License](https://opensource.org/licenses/MIT).