package config

// import (
// 	"github.com/MohammadBnei/go-ai-cli/api"
// 	"github.com/spf13/pflag"
// )

// type AI struct {
// 	Temperature, TopP float64
// 	TopK              int

// 	ModelName string
// 	ApiType   api.API_TYPE

// 	OpenAIKey      string
// 	HuggingfaceKey string

// 	OllamaHost string
// }

// type UI struct {
// 	MarkdownMode bool
// 	CodeMode     bool
// }

// type Prompt struct {
// 	SavedSystem   map[string]string
// 	DefaultSystem []string
// 	AutoLoad      bool
// }

// type config struct {
// 	UI
// 	AI
// 	Prompt
// }

// var Config *config = &config{}

// func (c *config) BindToFlags(flags pflag.FlagSet) {
// }

// func initConfig() {
// 	if cfgFile != "" {
// 		// Use config file from the flag.
// 		viper.SetConfigFile(cfgFile)
// 	} else {
// 		// Find home directory.
// 		home, err := os.UserHomeDir()
// 		cobra.CheckErr(err)

// 		// Search config in home directory with name ".go-ai-cli" (without extension).
// 		viper.AddConfigPath(home)
// 		viper.SetConfigName(".go-ai-cli")
// 	}

// 	viper.SetEnvPrefix("GO_AI_CLI") // set the environment variable prefix to GO_AI_CLI
// 	viper.AutomaticEnv()            // read in environment variables that match

// 	// If a config file is found, read it in.
// 	if err := viper.ReadInConfig(); err == nil {
// 		fmt.Println("Using config file:", viper.ConfigFileUsed())
// 	}

// 	bindFlagsToConfig()
// 	watchConfig()
// }

// func bindFlagsToConfig() {
// 	viper.BindPFlag("openai_key", RootCmd.PersistentFlags().Lookup("openai-key"))
// 	viper.BindPFlag("huggingface_key", RootCmd.PersistentFlags().Lookup("huggingface-key"))
// 	viper.BindPFlag("ollama_host", RootCmd.PersistentFlags().Lookup("ollama-host"))

// 	viper.BindPFlag("ai.temperature", RootCmd.PersistentFlags().Lookup(config.AI_TEMPERATURE))
// 	viper.BindPFlag("ai.top_p", RootCmd.PersistentFlags().Lookup(config.AI_TOP_P))
// 	viper.BindPFlag("ai.top_k", RootCmd.PersistentFlags().Lookup(config.AI_TOP_K))
// 	viper.BindPFlag("ai.model_name", RootCmd.PersistentFlags().Lookup(config.AI_MODEL_NAME))
// 	viper.BindPFlag("ai.api_type", RootCmd.PersistentFlags().Lookup(config.AI_API_TYPE))

// 	viper.BindPFlag("ui.markdown_mode", RootCmd.PersistentFlags().Lookup(config.UI_MARKDOWN_MODE))
// 	viper.BindPFlag("ui.code_mode", RootCmd.PersistentFlags().Lookup(config.UI_CODE_MODE))
// }

// func watchConfig() {
// 	viper.WatchConfig()
// 	viper.OnConfigChange(func(e fsnotify.Event) {
// 		fmt.Println("Config file changed:", e.Name)
// 		// Here you can add code to re-initialize anything that needs to be
// 		// re-loaded when the configuration changes.
// 	})
// }

// func init() {
// 	cobra.OnInitialize(initConfig)

// 	// Initialize all your flags here
// 	// Example:
// 	RootCmd.PersistentFlags().StringVarP(&config.Config.OpenAIKey, "openai-key", "o", "", "OpenAI API key")
// 	// More flags initialization...

// 	// After initializing flags, bind them to your configuration as needed
// }
