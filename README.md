<p align="center">
  <img src="https://github.com/rogersilvasouza/godeogoker/blob/main/gopher.png" alt="Godeogoker Mascot" width="200" height="200">
</p>

<h1 align="center">Godeogoker</h1>

<p align="center">
  <img src="https://img.shields.io/badge/platform-macOS-lightgrey" alt="Platform">
  <img src="https://img.shields.io/badge/language-Go-blue" alt="Language">
  <img src="https://img.shields.io/badge/license-MIT-green" alt="License">
</p>

<p align="center">
  <b>Your friendly Go-powered YouTube video chopper and uploader! 🎬 ✂️ 🚀</b>
</p>

---

## 🤖 What is Godeogoker?

Godeogoker (Go-Video-Chopper) is a powerful **command-line tool** that automates the tedious work of creating clips from longer YouTube videos. Think of it as your personal video assistant who:

1. 📥 **Downloads** videos from YouTube
2. 🧠 **Analyzes** content using OpenAI to find the perfect cut points
3. ✂️ **Segments** videos into topic-based clips
4. 🎨 **Formats** them with intros, overlays, and descriptions
5. 📤 **Uploads** everything back to YouTube automatically

All through the magic of your terminal! No GUI needed - just pure command-line efficiency.

## ✨ Features

- **100% CLI-based** workflow - perfect for automation and scripting
- Smart content analysis powered by OpenAI
- Multi-channel support
- Customizable video formatting with overlays and branding
- Automatic YouTube uploads

## 🚀 Quick Start

## Project Overview

Godeogoker provides an end-to-end workflow for content creators and video editors:

1. **Download** videos from YouTube using yt-dlp
2. **Segment** videos into meaningful cuts based on topics
3. **Analyze** content using OpenAI's API to extract insights and generate descriptions
4. **Create** optimized video clips from the original content
5. **Upload** the resulting clips back to YouTube automatically

This automated pipeline saves hours of manual video editing and content repurposing work.

> **⚠️ DISCLAIMER:**
> - **OpenAI API Usage**: Using the OpenAI integration incurs costs according to their [pricing structure](https://openai.com/pricing). Be sure to monitor your usage to avoid unexpected charges.
> - **YouTube API**: This application uses YouTube's API which has daily quota limits. Please respect these limits to avoid service interruptions. See [YouTube API Quotas](https://developers.google.com/youtube/v3/getting-started#quota) for more information.

## System Requirements

This application has been developed and tested exclusively on macOS. Support for other operating systems is not guaranteed.

## Prerequisites

Before installing godeogoker, ensure you have the following:

- Go 1.24 or later
- Homebrew package manager
- OpenAI API key
- Google OAuth2 credentials for a desktop application

## Installation Guide for macOS

### 1. Install Dependencies

First, install the required dependencies using Homebrew:

```bash
# Install ffmpeg for video processing
brew install ffmpeg

# Install yt-dlp for YouTube video downloading
brew install yt-dlp
```

### 2. Build and Install godeogoker

```bash
# Get the dependencies
go mod tidy
go mod vendor

# Build the application
go build -x

# Verify the build
file godeogoker

# Make the binary executable and move it to your path
sudo chmod +x godeogoker
sudo mv godeogoker /usr/local/bin/godeogoker
```

### 3. Configuration Setup

```bash
# Create your configuration file from the example
cp config.json.example config.json

# Edit the configuration file with your preferred editor
# Replace API keys and credentials as needed
vim config.json  # or use any editor of your choice
```

#### Configuration File Explained

The `config.json` file contains all the settings needed for godeogoker to operate:

```json
{
    "ytdlp": "/usr/local/bin/yt-dlp",      // Path to yt-dlp executable
    "ffmpeg": "/usr/local/bin/ffmpeg",     // Path to ffmpeg executable
    "ffprobe": "/usr/local/bin/ffprobe",   // Path to ffprobe executable
    "openai": {
        "key": "sk-",                      // Your OpenAI API key
        "model": "gpt-4o-mini-2024-07-18"  // OpenAI model to use
    },
    "channels": [
        {
            "id": "",                      // Unique identifier for this channel configuration
            "name": "",                    // Display name for the channel
            "channel_id": "",              // YouTube channel ID
            "url": "",                     // YouTube channel URL
            "folder": "",                  // Local folder to store downloads
            "video_base_vertical": "",     // Base template for video vertical
            "video_base_horizontal": "",   // Base template for video horizontal
            "video_cover": "",             // Cover image for videos
            "font": "",                    // Font to use for text overlays
            "font_size": "64",             // Font size for text overlays
            "font_color": "#FFFFFF",       // Font color for text overlays
            "font_effect": "...",          // Font effects for text overlays
            "description": "",             // Default description template
            "topics": "one,two,three",     // Topics to focus on when cutting
            "excerpts": 3,                 // Number of excerpts to generate
            "stretch_time": 1,             // Time factor for stretching clips
            "video_limit": 15,             // Maximum videos to process
            "upload_to_youtube": false     // Upload automatically to youtube
        },
        // Add more channel configurations here
    ]
}
```

**Finding Program Paths:**
To find the correct paths for your system, use the `which` command in your terminal:
```bash
which yt-dlp
which ffmpeg
which ffprobe
```
Update the values in your config file with the outputs from these commands.

**YouTube Channel ID:**
The `channel_id` is a unique identifier for each YouTube channel (e.g., MrBeast's is "UCX6OQ3DkcsbYNE6H8uQQuVA"). To find a channel ID:
- Use online tools like [Comment Picker](https://commentpicker.com/youtube-channel-id.php) or [YTCH ID](https://www.ytch-id.com/)
- Or view the channel page source and search for "channelId"

**Video Limit Setting:**
The `video_limit` parameter controls how many videos will be downloaded from the YouTube channel's XML feed. While the maximum is 15, it's recommended to use a lower value (like 3-5) when first testing to avoid quickly exhausting your API quotas.

**Note:** To process multiple YouTube channels, simply add additional objects to the `channels` array in your configuration file.

## 🚀 Performance Considerations

For optimal performance, godeogoker processes videos in 720p resolution by default. This provides a good balance between quality and processing speed.

**Important Processing Note:**
Videos longer than 20 minutes are automatically split into 20-minute segments to improve processing efficiency and reduce memory usage. These segments are processed individually and then recombined as needed.

**Benchmark Information:**
- A system with an Intel i5 processor and 8GB RAM typically takes:
  - ~3 minutes to process a 10-minute video
  - ~3 minutes for downloading and uploading (depends on internet speed)
  - Total ~6 minutes per video

- Systems with ARM architecture (like Apple M-series) can process videos up to twice as fast
- Faster internet connections will significantly reduce download and upload times
- Processing time scales roughly with video length and complexity

## 🔑 Getting API Credentials

### OpenAI API Key

1. Create an account or log in at [OpenAI's website](https://platform.openai.com/signup)
2. Navigate to [API Keys](https://platform.openai.com/account/api-keys)
3. Click "Create new secret key" and save it securely
4. Add this key to your `config.json` file

#### Model Selection and Cost Considerations

The default model in the configuration is `gpt-4o-mini-2024-07-18`, which offers a good balance between performance and cost. For most videos, processing with this model costs approximately $0.01 (1 cent) per video.

If you prefer different models:
- GPT-4.1 and GPT-o1 provide better results but significantly increase costs
- GPT-3.5-turbo is cheaper but provides less accurate analysis

Choose your model based on your budget and quality requirements.

### Google OAuth2 Credentials

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the YouTube Data API v3 from the API Library
4. Go to "Credentials" and click "Create Credentials" → "OAuth client ID"
5. Select "Desktop app" as the application type
6. Add your personal email to the "Test users" section if your app isn't in production yet
7. Download the credentials JSON file
8. Add the client ID and client secret to your `config.json` file

### Authentication Flow

When you run `godeogoker login`, you'll be directed to authenticate with YouTube. After granting permissions, you'll be redirected to a URL like:

```
http://localhost/?state=state&code=CODEHERE&scope=https://www.googleapis.com/auth/youtube.upload%20https://www.googleapis.com/auth/youtube.readonly
```

You'll need to copy the value from the `code=` parameter and paste it back into the CLI prompt. Godeogoker will handle the rest of the OAuth flow automatically!

**Important:** The authentication token obtained through this process is valid for only one hour. After this period, you'll need to run the `godeogoker login` command again to refresh your credentials.

## 🛠️ Usage

Godeogoker offers several command options:

```bash
# Generate YouTube channel credentials
godeogoker login

# Process all channels in your config.json
godeogoker exec

# Process a specific channel by ID
godeogoker exec {channel_id}

# Force regeneration of all content for a specific channel
godeogoker exec {channel_id} --force

# Process a specific video within a channel
godeogoker exec {channel_id} -v={youtube_video_id}
```

## 🤝 Contributing

Love cutting videos and writing Go? We'd love your contributions!

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is licensed under the MIT License - see below for details:

```
MIT License

Copyright (c) 2024 Godeogoker Contributors

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```