<p align="center">
  <img src="https://github.com/rogersilvasouza/godeogoker/blob/main/gopher.png" alt="Godeogoker mascot" width="200" height="200">
</p>

<h1 align="center">Godeogoker</h1>

<p align="center">
  <img src="https://img.shields.io/badge/platform-macOS-lightgrey" alt="Platform: macOS">
  <img src="https://img.shields.io/badge/language-Go-blue" alt="Language: Go">
  <img src="https://img.shields.io/badge/license-MIT-green" alt="License: MIT">
</p>

<p align="center">
  <b>A Go CLI for turning long YouTube videos into topic-based clips.</b>
</p>

## What it does

Godeogoker automates a video repurposing pipeline:

1. discovers recent videos from a YouTube channel RSS feed;
2. downloads videos and Portuguese automatic subtitles with `yt-dlp`;
3. splits videos longer than 20 minutes;
4. asks OpenAI to identify relevant excerpts and generate metadata;
5. cuts and formats horizontal and vertical videos with FFmpeg;
6. optionally uploads the generated videos to YouTube.

The project has been developed and tested on macOS.

## Requirements

- Go 1.25.8 or a compatible newer release
- [Homebrew](https://brew.sh/)
- [FFmpeg](https://ffmpeg.org/) and FFprobe
- [yt-dlp](https://github.com/yt-dlp/yt-dlp)
- an OpenAI API key
- Google OAuth desktop-app credentials with the YouTube Data API v3 enabled

OpenAI requests may incur charges, and YouTube uploads consume API quota. Review the
[OpenAI pricing](https://openai.com/api/pricing/) and
[YouTube quota documentation](https://developers.google.com/youtube/v3/getting-started#quota)
before processing large batches.

## Quick start

Install the external tools:

```bash
brew install ffmpeg yt-dlp
```

Build and install the CLI:

```bash
git clone https://github.com/rogersilvasouza/godeogoker.git
cd godeogoker
go build -o godeogoker .
install -m 755 godeogoker /usr/local/bin/godeogoker
```

If `/usr/local/bin` is not writable by your user, run only the `install` command
with `sudo`.

Create the local application configuration:

```bash
cp config.json.example config.json
```

Edit `config.json` and set the paths returned by:

```bash
which yt-dlp ffmpeg ffprobe
```

At minimum, configure the OpenAI key and model, a writable output folder, the
YouTube channel ID, and the desired `yt-dlp` format. `config.json` is ignored by
Git.

## Configuration

The complete configuration template is available in
[`config.json.example`](config.json.example). Each channel supports:

| Field | Purpose |
| --- | --- |
| `id` | Short identifier used by the CLI, such as `my-channel`. |
| `name` | Display name shown during processing. |
| `channel_id` | YouTube channel ID. |
| `folder` | Directory where downloaded and generated files are stored. |
| `topics` | Topics OpenAI should prioritize when selecting excerpts. |
| `excerpts` | Desired number of excerpts per analyzed subtitle file. |
| `stretch_time` | Approximate clip length in minutes. |
| `video_limit` | Maximum number of recent feed entries to process. |
| `ytdlp_format` | Format expression passed to `yt-dlp`. |
| `video_base_vertical` | Optional background used for vertical output. |
| `video_base_horizontal` | Optional background used for branded horizontal output. |
| `video_cover` | Optional base image used to generate covers. |
| `upload_to_youtube` | Upload generated videos when authentication is available. |

The `font`, `font_size`, `font_color`, and `font_effect` fields customize generated
covers. To process multiple channels, add more objects to the `channels` array.

Start with `video_limit` set to `1` while validating your configuration and API
usage.

## Google credentials and login

1. Create a project in the [Google Cloud Console](https://console.cloud.google.com/).
2. Enable the YouTube Data API v3.
3. Configure the OAuth consent screen and add your account as a test user when
   required.
4. Create OAuth credentials for a **Desktop app**.
5. Download the JSON file and save it in the repository root as
   `credentials.json`.
6. Run:

```bash
godeogoker login
```

Open the URL printed by the CLI, authorize access, copy the `code` query parameter
from the redirect URL, and paste it into the prompt. The token is saved locally as
`youtube-token.json`. Both credential files are ignored by Git.

If the saved access token expires, run `godeogoker login` again.

## Usage

```text
godeogoker login
godeogoker exec
godeogoker exec <channel-id>
godeogoker exec <channel-id> --force
godeogoker exec <channel-id> -v=<youtube-video-id>
godeogoker help
```

- `exec` without a channel processes every configured channel.
- `--force` deletes the existing output directory for a video before processing it.
- `-v=` processes one YouTube video using the selected channel configuration.

Generated `.mp4` files are ignored by Git.

## Install a release

Prebuilt artifacts, when available for your macOS architecture, are published on
the [releases page](https://github.com/rogersilvasouza/godeogoker/releases).
Building from source is recommended when no matching artifact is available.

## Development

Format, analyze, and build the project before opening a pull request:

```bash
go fmt ./...
go mod tidy
go vet ./...
go build ./...
```

The repository currently has no automated test suite. Changes to the video
pipeline should be checked with a low `video_limit` and a disposable output
directory.

## License

This project is distributed under the [MIT License](LICENSE).
