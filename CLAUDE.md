# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Godeogoker** (Go-Video-Chopper) is a CLI application that automates viral video clip creation from YouTube videos. It downloads videos, uses OpenAI for content analysis, cuts videos into topic-based clips, adds subtitles/overlays, and uploads back to YouTube.

## Architecture

### Main Components
- **`main.go`** - CLI entry point with lipgloss styling and command parsing
- **`internal/auth/auth.go`** - OAuth2 authentication for YouTube API
- **`internal/config/main.go`** - JSON-based configuration management  
- **`internal/videos/main.go`** - Core video processing pipeline (~1000+ lines)

### Processing Pipeline
1. **Discovery** - Fetches video IDs from YouTube RSS feeds
2. **Download** - Uses yt-dlp for videos/subtitles
3. **Segmentation** - Splits videos >20min into chunks
4. **AI Analysis** - OpenAI identifies interesting segments
5. **Video Cutting** - Creates clips using moviego/ffmpeg
6. **Enhancement** - Adds subtitles, overlays, branding
7. **Upload** - Automatic YouTube upload (optional)

## Development Commands

### Build and Install
```bash
# Get dependencies
go mod tidy
go mod vendor

# Build application
go build -x

# Install locally
sudo chmod +x godeogoker
sudo mv godeogoker /usr/local/bin/godeogoker
```

### Configuration Setup
```bash
# Create config from template
cp config.json.example config.json

# Find system tool paths
which yt-dlp ffmpeg ffprobe
```

### CLI Commands
```bash
# OAuth authentication
godeogoker login

# Process all channels
godeogoker exec

# Process specific channel
godeogoker exec {channel_id}

# Force regeneration
godeogoker exec {channel_id} --force

# Process single video
godeogoker exec {channel_id} -v={youtube_video_id}
```

## Dependencies

### External Tools (required)
- **yt-dlp** - YouTube video downloading
- **ffmpeg/ffprobe** - Video processing
- Install via: `brew install ffmpeg yt-dlp`

### Go Dependencies
- `github.com/charmbracelet/lipgloss` - CLI styling
- `github.com/mowshon/moviego` - Video manipulation
- `golang.org/x/oauth2` + `google.golang.org/api` - YouTube API

### API Requirements
- **OpenAI API key** - Content analysis (default: gpt-4o-mini-2024-07-18)
- **Google OAuth2 credentials** - YouTube API access (desktop app type)

## Configuration

Uses JSON config file with:
- Tool paths (yt-dlp, ffmpeg, ffprobe)
- OpenAI settings (key, model)
- Channel configurations (multiple supported)
- Video processing parameters

Key settings:
- `video_limit`: Controls how many videos to process (max 15, recommend 3-5 for testing)
- `ytdlp_format`: Default is 720p for performance balance
- `excerpts`: Number of clips to generate per video
- `upload_to_youtube`: Auto-upload toggle

## Performance Notes

- Videos >20 minutes auto-split into segments
- Default 720p processing for speed/quality balance
- ARM processors (Apple M-series) ~2x faster than Intel
- Typical processing: ~3min for 10min video + upload time
- OpenAI costs: ~$0.01 per video with gpt-4o-mini

## Platform Support

**macOS only** - developed and tested exclusively on macOS. Other platforms not supported.

## Testing

No automated tests present. Manual testing workflow involves:
1. Configure with test channel
2. Run with low `video_limit` 
3. Monitor OpenAI/YouTube API quotas
4. Verify output video quality

## Common Issues

- **Authentication expires** - OAuth tokens valid 1 hour, re-run `login`
- **API quota limits** - Monitor OpenAI usage and YouTube daily quotas
- **Processing failures** - Check external tool paths in config
- **Video format issues** - Ensure ffmpeg/yt-dlp are latest versions