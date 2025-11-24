package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type LoggingConfig struct {
	Enabled               bool   `json:"enabled"`
	Directory             string `json:"directory"`
	LogAnthropicRequest   bool   `json:"log_anthropic_request"`
	LogOpenAIRequest      bool   `json:"log_openai_request"`
	LogOpenAIResponse     bool   `json:"log_openai_response"`
	LogAnthropicResponse  bool   `json:"log_anthropic_response"`
}

type Env struct {
	OpenRouterBaseUrl string            `json:"openrouter_base_url"`
	ModelMappings     map[string]string `json:"model_mappings"`
	DataLogging       LoggingConfig     `json:"data_logging"`
}

var env Env
var dataLogger *DataLogger

func init() {
	// Set default first, then load config, then check environment variable for override
	env.OpenRouterBaseUrl = "https://openrouter.ai/api/v1"
	loadConfig()
	// Allow environment variable to override config file
	if envVar := getEnv("OPENROUTER_BASE_URL", ""); envVar != "" {
		env.OpenRouterBaseUrl = envVar
	}
	// Initialize data logger
	dataLogger = NewDataLogger(env.DataLogging)
}

func loadConfig() {
	configFile := "config.json"
	if _, err := os.Stat(configFile); err == nil {
		file, err := os.Open(configFile)
		if err != nil {
			log.Printf("Failed to open config file: %v", err)
			return
		}
		defer file.Close()

		var config struct {
			OpenRouterBaseUrl string            `json:"openrouter_base_url"`
			ModelMappings     map[string]string `json:"model_mappings"`
			DataLogging       LoggingConfig     `json:"data_logging"`
		}
		
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&config); err != nil {
			log.Printf("Failed to decode config file: %v", err)
			return
		}
		
		if config.OpenRouterBaseUrl != "" {
			env.OpenRouterBaseUrl = config.OpenRouterBaseUrl
		}
		env.ModelMappings = config.ModelMappings
		env.DataLogging = config.DataLogging
		log.Printf("Loaded configuration with %d model mappings", len(env.ModelMappings))
		log.Printf("Data logging enabled: %v", env.DataLogging.Enabled)
	} else {
		// Default mappings if config file doesn't exist
		env.ModelMappings = map[string]string{
			"haiku":  "anthropic/claude-3.5-haiku",
			"sonnet": "anthropic/claude-sonnet-4",
			"opus":   "anthropic/claude-opus-4",
		}
		log.Printf("Using default model mappings")
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	r := gin.Default()

	// 静态页面路由
	r.GET("/", handleIndex)
	//r.GET("/terms", handleTerms)
	//r.GET("/privacy", handlePrivacy)
	//r.GET("/install.sh", handleInstallSh)

	// API路由
	r.POST("/v1/messages", handleMessages)

	// 启动服务器
	port := getEnv("PORT", "8080")
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}