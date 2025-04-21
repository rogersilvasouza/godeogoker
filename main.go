// Package main is the entry point for the godeogoker application.
// This CLI application allows users to download videos from channels
// after authenticating with Google.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/rogersilvasouza/godeogoker/internal/auth"
	"github.com/rogersilvasouza/godeogoker/internal/config"
	"github.com/rogersilvasouza/godeogoker/internal/videos"
)

// Define styles for the CLI interface
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF5F87")).
			Border(lipgloss.RoundedBorder()).
			Padding(0, 1).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Italic(true).
			Foreground(lipgloss.Color("#5F87FF"))

	commandStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#5FFFAF"))

	optionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF87"))

	descriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#D7D7D7"))

	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF0000"))

	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FF00"))
)

// main is the entry point of the application.
// It parses command-line arguments and routes to the appropriate handlers.
func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	switch args[0] {
	case "login":
		fmt.Println(subtitleStyle.Render("🔑 Starting Google authentication process..."))
		if err := auth.Login(); err != nil {
			fmt.Println(errorStyle.Render(fmt.Sprintf("Login error: %v", err)))
			os.Exit(1)
		}
		fmt.Println(successStyle.Render("🎉 Login successful! You're ready to download videos!"))
	case "exec":
		fmt.Println(subtitleStyle.Render("🚀 Preparing to download awesome content..."))
		handleExec(args[1:])
	case "help":
		printExtendedHelp()
	default:
		printUsage()
		os.Exit(1)
	}
}

// printUsage displays styled help information showing available commands and options.
func printUsage() {
	fmt.Println(commandStyle.Render("Usage:"), descriptionStyle.Render("godeogoker <command> [options]"))
	fmt.Println()
	fmt.Println(commandStyle.Render("Commands:"))
	fmt.Println(optionStyle.Render("  - login:"), descriptionStyle.Render("Authenticate with Google (you'll need this first!)"))
	fmt.Println(optionStyle.Render("  - exec [channelID] [--force] [-v=videoID]:"), descriptionStyle.Render("Download videos"))
	fmt.Println(descriptionStyle.Render("    [channelID]: Optional. Specific channel ID for download"))
	fmt.Println(descriptionStyle.Render("    [--force]: Optional. Force reprocessing even if folder exists"))
	fmt.Println(descriptionStyle.Render("    [-v=videoID]: Optional. Specific video ID for processing"))
	fmt.Println(optionStyle.Render("  - help:"), descriptionStyle.Render("Show extended help with examples"))
	fmt.Println()
	fmt.Println(subtitleStyle.Render("💡 Tip:"), descriptionStyle.Render("Start with 'godeogoker login' to authenticate!"))
}

// printExtendedHelp displays detailed help information with examples
func printExtendedHelp() {
	fmt.Println(titleStyle.Render("🎬 Godeogoker - Video Downloader"))
	fmt.Println(subtitleStyle.Render("Your friendly assistant to download and organize videos from your favorite channels"))
	fmt.Println()

	fmt.Println(commandStyle.Render("How It Works:"))
	fmt.Println(descriptionStyle.Render("1. First, authenticate with Google using 'godeogoker login'"))
	fmt.Println(descriptionStyle.Render("2. Then download videos with 'godeogoker exec'"))
	fmt.Println(descriptionStyle.Render("3. Videos are organized by channel in your configured download directory"))
	fmt.Println()

	fmt.Println(commandStyle.Render("Examples:"))
	fmt.Println(optionStyle.Render("- Download videos from all configured channels:"))
	fmt.Println(descriptionStyle.Render("  godeogoker exec"))
	fmt.Println()

	fmt.Println(optionStyle.Render("- Download videos from a specific channel:"))
	fmt.Println(descriptionStyle.Render("  godeogoker exec mrbeast"))
	fmt.Println()

	fmt.Println(optionStyle.Render("- Download a specific video:"))
	fmt.Println(descriptionStyle.Render("  godeogoker exec mrbeast -v=0e3GPea1Tyg"))
	fmt.Println()

	fmt.Println(optionStyle.Render("- Force reprocessing of existing videos:"))
	fmt.Println(descriptionStyle.Render("  godeogoker exec --force"))
	fmt.Println()

	fmt.Println(commandStyle.Render("Troubleshooting:"))
	fmt.Println(descriptionStyle.Render("- If you encounter authentication issues, try 'godeogoker login' again"))
	fmt.Println(descriptionStyle.Render("- Make sure your channel IDs are correct in the configuration"))
	fmt.Println()

	fmt.Println(subtitleStyle.Render("📝 Fun fact:"), descriptionStyle.Render("The average YouTube channel produces about 600 hours of content yearly!"))
}

// handleExec processes the exec command with its arguments.
// It parses flags and options, then initiates the video download process
// for either a specific channel or all configured channels.
func handleExec(args []string) {
	force := false
	var videoID string

	i := 0
	for i < len(args) {
		switch {
		case args[i] == "--force":
			force = true
			args = append(args[:i], args[i+1:]...)
		case strings.HasPrefix(args[i], "-v=") || strings.HasPrefix(args[i], "--v="):
			videoID = strings.Split(args[i], "=")[1]
			args = append(args[:i], args[i+1:]...)
		default:
			i++
		}
	}

	channels := config.GetChannels()

	if len(args) > 0 {
		channelID := args[0]
		channelFound := false

		for _, channel := range channels {
			if channel.ID == channelID {
				if videoID != "" {
					channel.ChannelID = "v=" + videoID
				}
				fmt.Println(subtitleStyle.Render(fmt.Sprintf("📥 Downloading videos for channel: %s", channel.Name)))
				videos.DownloadVideo(channel, force)
				channelFound = true
				break
			}
		}

		if !channelFound {
			fmt.Println(errorStyle.Render(fmt.Sprintf("Error: Channel with ID '%s' not found", channelID)))
			os.Exit(1)
		}
	} else {
		fmt.Println(subtitleStyle.Render("🎯 Starting batch download for all channels..."))
		for _, channel := range channels {
			if videoID != "" {
				channel.ChannelID = "v=" + videoID
			}
			fmt.Println(subtitleStyle.Render(fmt.Sprintf("📥 Downloading videos for channel: %s", channel.Name)))
			videos.DownloadVideo(channel, force)
		}
	}

	fmt.Println(successStyle.Render("🎉 Download completed successfully! Enjoy your videos!"))
}
