// Package config provides functionality for loading and accessing application configuration.
// The configuration is loaded from a JSON file and stored in memory for easy access.
package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// OpenAI represents configuration settings for the OpenAI API integration.
type OpenAI struct {
	Key   string `json:"key"`   // API key for authentication with OpenAI services
	Model string `json:"model"` // The name of the model to be used for AI operations
}

// Channel represents configuration for a media channel that the application processes.
// It contains all necessary information to handle videos from this channel.
type Channel struct {
	ID                  string `json:"id"`                    // Unique identifier for the channel
	Name                string `json:"name"`                  // Display name of the channel
	ChannelID           string `json:"Channel_id"`            // Platform-specific channel identifier
	URL                 string `json:"url"`                   // URL to the channel
	Folder              string `json:"folder"`                // Local folder where channel content is stored
	VerticalVideoBase   string `json:"video_base_vertical"`   // Base template for vertical video format
	HorizontalVideoBase string `json:"video_base_horizontal"` // Base template for horizontal video format
	CoverVideoBase      string `json:"video_cover"`           // Base template for video covers
	Description         string `json:"description"`           // Channel description
	LastCheck           string `json:"last_check,omitempty"`  // Timestamp of the last content check
	Topics              string `json:"topics"`                // Topics or categories for the channel
	Excerpts            int    `json:"excerpts"`              // Number of excerpts to generate
	StretchTime         int    `json:"stretch_time"`          // Time to stretch content in seconds
	VideoLimit          int    `json:"video_limit"`           // Maximum number of videos to process
	Font                string `json:"font"`                  // Font to use for text overlays
	FontSize            string `json:"font_size"`             // Font size for text overlays
	FontColor           string `json:"font_color"`            // Font color for text overlays
	FontEffect          string `json:"font_effect"`           // Special effects to apply to text
	UploadToYouTube     bool   `json:"upload_to_youtube"`     // Whether to upload processed videos to YouTube
	YtdlpFormat         string `json:"ytdlp_format"`          // Format string for yt-dlp
}

// Config represents the main application configuration structure.
// It contains paths to required external tools and application settings.
type Config struct {
	YtDlp    string    `json:"ytdlp"`    // Path to the yt-dlp executable
	FFmpeg   string    `json:"ffmpeg"`   // Path to the FFmpeg executable
	FFprobe  string    `json:"ffprobe"`  // Path to the FFprobe executable
	OpenAI   OpenAI    `json:"openai"`   // OpenAI API configuration
	Channels []Channel `json:"channels"` // List of channels to process
}

// configInstance holds the singleton instance of loaded configuration
var configInstance *Config

// loadConfig reads and parses the configuration file from the specified path.
// It returns a pointer to the Config structure and any error encountered.
func loadConfig(filePath string) (*Config, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading JSON file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, fmt.Errorf("error parsing JSON file: %w", err)
	}

	return &config, nil
}

// init is automatically called when the package is imported.
// It loads the configuration from the default location.
func init() {
	var err error
	configInstance, err = loadConfig("config.json")
	if err != nil {
		log.Fatalf("Error loading JSON configuration file: %v", err)
	}
}

// GetChannels returns the list of configured channels.
func GetChannels() []Channel {
	return configInstance.Channels
}

// GetYtDlp returns the path to the yt-dlp executable.
func GetYtDlp() string {
	return configInstance.YtDlp
}

// GetFFmpeg returns the path to the FFmpeg executable.
func GetFFmpeg() string {
	return configInstance.FFmpeg
}

// GetFFprobe returns the path to the FFprobe executable.
func GetFFprobe() string {
	return configInstance.FFprobe
}

// GetOpenAIKey returns the OpenAI API key.
func GetOpenAIKey() string {
	return configInstance.OpenAI.Key
}

// GetOpenAIModel returns the name of the OpenAI model to use.
func GetOpenAIModel() string {
	return configInstance.OpenAI.Model
}
