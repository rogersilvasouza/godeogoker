package videos

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mowshon/moviego"
	"github.com/rogersilvasouza/godeogoker/internal/auth"
	"github.com/rogersilvasouza/godeogoker/internal/config"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
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

// Feed represents the YouTube RSS feed structure
// with entries containing video information
type Feed struct {
	XMLName xml.Name `xml:"feed"`
	Entries []struct {
		ID    string `xml:"id"`
		Title string `xml:"title"`
	} `xml:"entry"`
}

// GetLastVideos retrieves video IDs from a YouTube channel using its RSS feed.
// If channel.ChannelID starts with "v=", it processes a specific video instead.
// It respects the video limit set in the channel configuration.
func GetLastVideos(channel config.Channel) []string {
	fmt.Println(titleStyle.Render("Getting videos from channel: " + channel.Name))

	if strings.HasPrefix(channel.ChannelID, "v=") {
		videoID := strings.TrimPrefix(channel.ChannelID, "v=")
		fmt.Println(subtitleStyle.Render("Processing specific video: " + videoID))
		return []string{videoID}
	}

	feedURL := fmt.Sprintf("https://www.youtube.com/feeds/videos.xml?channel_id=%s", channel.ChannelID)
	fmt.Println(descriptionStyle.Render("Fetching RSS feed: " + feedURL))

	resp, err := http.Get(feedURL)
	if err != nil {
		fmt.Println(errorStyle.Render("Error requesting RSS feed: " + err.Error()))
		log.Fatalf("Error requesting RSS feed: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(errorStyle.Render("Error reading RSS feed: " + err.Error()))
		log.Fatalf("Error reading RSS feed response: %v", err)
	}

	var feed Feed
	if err := xml.Unmarshal(body, &feed); err != nil {
		fmt.Println(errorStyle.Render("Error parsing RSS feed: " + err.Error()))
		log.Fatalf("Error unmarshalling RSS: %v", err)
	}

	if len(feed.Entries) == 0 {
		fmt.Println(errorStyle.Render("No videos found for channel: " + channel.Name))
		log.Fatalf("No videos found for channel: %s", channel.Name)
	}

	fmt.Println(subtitleStyle.Render(fmt.Sprintf("Total videos found: %d", len(feed.Entries))))

	videoLimit := channel.VideoLimit
	if videoLimit == 0 || videoLimit > len(feed.Entries) {
		videoLimit = len(feed.Entries)
	}

	fmt.Println(subtitleStyle.Render(fmt.Sprintf("Processing %d videos", videoLimit)))

	var videoIDs []string
	for i := 0; i < videoLimit; i++ {
		videoID := extractVideoID(feed.Entries[i].ID)
		fmt.Println(optionStyle.Render(fmt.Sprintf("Video %d: %s (ID: %s)", i+1, feed.Entries[i].Title, videoID)))
		videoIDs = append(videoIDs, videoID)
	}

	return videoIDs
}

func extractVideoID(rssID string) string {
	videoID := rssID[strings.LastIndex(rssID, ":")+1:]
	return videoID
}

func splitLongVideo(videoFileName string, subtitleFileName string) ([]string, []string, error) {
	const segmentDuration = 1200
	var videoSegments []string
	var subtitleSegments []string

	ffprobePath := config.GetFFprobe()
	cmd := exec.Command(ffprobePath, "-v", "quiet", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", videoFileName)
	output, err := cmd.Output()
	if err != nil {
		return nil, nil, fmt.Errorf("error getting video duration: %v", err)
	}

	duration, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		return nil, nil, fmt.Errorf("error converting video duration: %v", err)
	}

	if duration <= float64(segmentDuration) {
		return []string{videoFileName}, []string{subtitleFileName}, nil
	}

	numSegments := int(math.Ceil(duration / float64(segmentDuration)))
	ffmpegPath := config.GetFFmpeg()

	for i := 0; i < numSegments; i++ {
		startTime := i * segmentDuration
		segmentVideoFile := fmt.Sprintf("%s.part%d.mp4", videoFileName[:len(videoFileName)-4], i+1)
		segmentSubtitleFile := fmt.Sprintf("%s.part%d.srt", subtitleFileName[:len(subtitleFileName)-4], i+1)

		cmd = exec.Command(ffmpegPath,
			"-i", videoFileName,
			"-ss", fmt.Sprintf("%d", startTime),
			"-t", fmt.Sprintf("%d", segmentDuration),
			"-c", "copy",
			"-y",
			segmentVideoFile)

		if err := cmd.Run(); err != nil {
			return nil, nil, fmt.Errorf("error splitting video segment %d: %v", i+1, err)
		}

		if subtitleEntries, err := parseVTTFile(subtitleFileName + ".pt.vtt"); err == nil {
			subtitleText := getSubtitlesForTimeRange(subtitleEntries, startTime, startTime+segmentDuration)
			if err := ioutil.WriteFile(segmentSubtitleFile, []byte(subtitleText), 0644); err != nil {
				log.Printf("Error creating subtitle file for segment %d: %v", i+1, err)
			} else {
				subtitleSegments = append(subtitleSegments, segmentSubtitleFile)
			}
		}

		videoSegments = append(videoSegments, segmentVideoFile)
	}

	return videoSegments, subtitleSegments, nil
}

func DownloadVideo(channel config.Channel, force bool) {
	fmt.Println(titleStyle.Render("Processing channel: " + channel.Name))

	videoIDs := GetLastVideos(channel)

	for i, videoID := range videoIDs {
		fmt.Println(titleStyle.Render(fmt.Sprintf("Processing video %d/%d (ID: %s)", i+1, len(videoIDs), videoID)))

		outputDir := channel.Folder + "/" + videoID

		if !force {
			if _, err := os.Stat(outputDir); err == nil {
				fmt.Println(subtitleStyle.Render("Video already processed. Skipping. Use force=true to reprocess."))
				continue
			}
		} else {
			if _, err := os.Stat(outputDir); err == nil {
				fmt.Println(subtitleStyle.Render("Removing existing processed files..."))
				if err := os.RemoveAll(outputDir); err != nil {
					fmt.Println(errorStyle.Render("Error removing directory: " + err.Error()))
					continue
				}
			}
		}

		videoFileName := outputDir + "/" + fmt.Sprintf("%s.mp4", videoID)
		subtitleFileName := outputDir + "/" + fmt.Sprintf("%s.srt", videoID)
		videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
		ytDlpPath := config.GetYtDlp()

		if err := os.MkdirAll(outputDir+"/horizontal", 0755); err != nil {
			fmt.Println(errorStyle.Render("Error creating output directory: " + err.Error()))
			continue
		}

		if _, err := os.Stat(videoFileName); os.IsNotExist(err) {
			fmt.Println(commandStyle.Render("Downloading video..."))
			cmd := exec.Command(
				ytDlpPath,
				"--ignore-errors",
				"--merge-output-format", "mp4",
				"--geo-bypass",
				"--no-check-certificate",
				"--force-generic-extractor",
				"--format", channel.YtdlpFormat,
				"--concurrent-fragments", "8",
				"-o",
				videoFileName,
				videoURL,
			)

			if _, err := cmd.CombinedOutput(); err != nil {
				fmt.Println(errorStyle.Render(fmt.Sprintf("Error downloading video: %v", err)))
				continue
			}
			fmt.Println(successStyle.Render("Video downloaded successfully"))
		} else {
			fmt.Println(subtitleStyle.Render("Video file already exists. Skipping download."))
		}

		if _, err := os.Stat(subtitleFileName); os.IsNotExist(err) {
			fmt.Println(commandStyle.Render("Downloading subtitles..."))
			cmd := exec.Command(
				ytDlpPath,
				"--write-auto-sub",
				"--sub-lang", "pt",
				"--skip-download",
				"--output", subtitleFileName,
				videoURL,
			)
			if _, err := cmd.CombinedOutput(); err != nil {
				fmt.Println(errorStyle.Render(fmt.Sprintf("Error downloading subtitles: %v", err)))
				continue
			}
			fmt.Println(successStyle.Render("Subtitles downloaded successfully"))
		} else {
			fmt.Println(subtitleStyle.Render("Subtitle file already exists. Skipping download."))
		}

		fmt.Println(commandStyle.Render("Processing video segments..."))
		videoSegments, subtitleSegments, err := splitLongVideo(videoFileName, subtitleFileName)
		if err != nil {
			fmt.Println(errorStyle.Render("Error splitting video: " + err.Error()))
			continue
		}

		for i, segmentVideoFile := range videoSegments {
			fmt.Println(subtitleStyle.Render(fmt.Sprintf("Processing segment %d/%d", i+1, len(videoSegments))))

			segmentSubtitleFile := subtitleSegments[i]
			fmt.Println(commandStyle.Render("Finding interesting cuts in this segment..."))
			cuts := GetCuts(segmentSubtitleFile, channel.Topics, channel.Excerpts, channel.StretchTime)

			if len(cuts) > 0 {
				fmt.Println(successStyle.Render(fmt.Sprintf("Found %d interesting cuts", len(cuts))))

				video, err := moviego.Load(segmentVideoFile)
				if err != nil {
					fmt.Println(errorStyle.Render("Error loading video segment: " + err.Error()))
					continue
				}

				videoDuration := video.Duration()

				for j, cut := range cuts {
					fmt.Println(optionStyle.Render(fmt.Sprintf("Processing cut %d/%d: %s", j+1, len(cuts), cut.Title)))

					tempOutputFileName := fmt.Sprintf("%s/temp_%s.mp4", outputDir, cut.Title)
					outputFileName := fmt.Sprintf("%s/horizontal/%s.mp4", outputDir, cut.Title)

					if videoDuration < float64(cut.Begin) || videoDuration < float64(cut.End) {
						fmt.Println(errorStyle.Render("Cut time exceeds video duration. Skipping."))
						continue
					}

					fmt.Println(descriptionStyle.Render(fmt.Sprintf("Creating clip from %d to %d seconds", cut.Begin, cut.End)))
					if err := video.SubClip(float64(cut.Begin), float64(cut.End)).Output(tempOutputFileName).Run(); err != nil {
						fmt.Println(errorStyle.Render("Error creating clip: " + err.Error()))
						continue
					}

					subtitleEntries, err := parseVTTFile(subtitleFileName + ".pt.vtt")
					if err != nil {
						fmt.Println(subtitleStyle.Render("Creating clip without subtitles"))
						os.Rename(tempOutputFileName, outputFileName)
						continue
					}

					cutSubtitleFileName := fmt.Sprintf("%s/temp_%s.srt", outputDir, cut.Title)
					subtitleText := getSubtitlesForTimeRange(subtitleEntries, cut.Begin, cut.End)

					if err := ioutil.WriteFile(cutSubtitleFileName, []byte(subtitleText), 0644); err != nil {
						fmt.Println(errorStyle.Render("Error writing subtitle file: " + err.Error()))
						os.Rename(tempOutputFileName, outputFileName)
						continue
					}

					// Extract clean text from subtitles for metadata generation
					var subtitleContent string
					for _, entry := range subtitleEntries {
						if (entry.StartTime >= time.Duration(cut.Begin)*time.Second) &&
							(entry.EndTime <= time.Duration(cut.End)*time.Second) {
							subtitleContent += " " + cleanSubtitleText(entry.Text)
						}
					}
					subtitleContent = strings.TrimSpace(subtitleContent)

					// Generate SEO-optimized metadata
					fmt.Println(commandStyle.Render("Generating metadata..."))
					metadata, err := GenerateMetadata(cut.Title, subtitleContent, channel.Topics)
					if err == nil && metadata != nil {
						metadataFile := fmt.Sprintf("%s/horizontal/%s.json", outputDir, cut.Title)
						metadataJSON, _ := json.MarshalIndent(metadata, "", "  ")
						ioutil.WriteFile(metadataFile, metadataJSON, 0644)
						fmt.Println(successStyle.Render("Metadata generated successfully"))
					} else {
						fmt.Println(errorStyle.Render(fmt.Sprintf("Error generating metadata: %v", err)))
					}

					fmt.Println(commandStyle.Render("Adding subtitles to video..."))
					ffmpegPath := config.GetFFmpeg()
					cmd := exec.Command(
						ffmpegPath,
						"-i", tempOutputFileName,
						"-vf", "subtitles="+cutSubtitleFileName+":force_style='FontSize=22,Alignment=2'",
						"-c:a", "aac",
						"-c:v", "libx264",
						"-preset", "ultrafast",
						"-tune", "fastdecode",
						"-crf", "28",
						"-threads", "0",
						"-y",
						outputFileName,
					)

					if err := cmd.Run(); err != nil {
						fmt.Println(errorStyle.Render("Error adding subtitles: " + err.Error()))
						os.Rename(tempOutputFileName, outputFileName)
					} else {
						fmt.Println(successStyle.Render("Subtitles added successfully"))
					}

					os.Remove(tempOutputFileName)
					os.Remove(cutSubtitleFileName)

					if channel.CoverVideoBase != "" {
						fmt.Println(commandStyle.Render("Generating cover image..."))
						coverOutputDir := outputDir + "/covers"
						if _, err := os.Stat(coverOutputDir); os.IsNotExist(err) {
							os.Mkdir(coverOutputDir, 0755)
						}

						coverOutputFileName := fmt.Sprintf("%s/%s.jpg", coverOutputDir, cut.Title)

						words := strings.Fields(cut.Title)
						formattedTitle := cut.Title
						if len(words) > 3 {
							var lines []string
							for i := 0; i < len(words); i += 3 {
								end := i + 3
								if end > len(words) {
									end = len(words)
								}
								lines = append(lines, strings.Join(words[i:end], " "))
							}
							formattedTitle = strings.Join(lines, "\n")
						}

						fontSize := "36"
						if channel.FontSize != "" {
							fontSize = channel.FontSize
						}

						fontColor := "white"
						if channel.FontColor != "" {
							fontColor = channel.FontColor
						}

						fontName := ""
						fontParam := ""
						if channel.Font != "" {
							fontName = channel.Font
							fontParam = ":fontfile=" + fontName
						}

						fontEffect := ""
						if channel.FontEffect != "" {
							fontEffect = channel.FontEffect
						}

						cmd := exec.Command(
							ffmpegPath,
							"-i", channel.CoverVideoBase,
							"-vf", fmt.Sprintf("drawtext=text='%s':fontsize=%s:fontcolor=%s%s:x=(w-text_w)/2:y=(h-text_h)/2%s",
								formattedTitle, fontSize, fontColor, fontParam, fontEffect),
							"-frames:v", "1",
							"-y",
							coverOutputFileName,
						)

						if err := cmd.Run(); err != nil {
							fmt.Println(errorStyle.Render("Error generating cover image: " + err.Error()))
						} else {
							fmt.Println(successStyle.Render("Cover image generated successfully"))
						}
					}

					if channel.VerticalVideoBase != "" {
						fmt.Println(commandStyle.Render("Creating vertical version..."))
						verticalOutputDir := outputDir + "/vertical"
						if _, err := os.Stat(verticalOutputDir); os.IsNotExist(err) {
							os.Mkdir(verticalOutputDir, 0755)
						}

						verticalOutputFileName := fmt.Sprintf("%s/%s.mp4", verticalOutputDir, cut.Title)
						cmd := exec.Command(
							ffmpegPath,
							"-i", channel.VerticalVideoBase,
							"-i", outputFileName,
							"-filter_complex", "[0:v]loop=loop=-1:size=1:start=0[loopbg];[1:v]scale=1080:-1[scaled];[loopbg][scaled]overlay=(W-w)/2:(H-h)/2:shortest=1[outv]",
							"-map", "[outv]",
							"-map", "1:a",
							"-c:a", "aac",
							"-c:v", "libx264",
							"-preset", "ultrafast",
							"-tune", "fastdecode",
							"-crf", "28",
							"-threads", "0",
							"-shortest",
							"-y",
							verticalOutputFileName,
						)
						if err := cmd.Run(); err != nil {
							fmt.Println(errorStyle.Render("Error creating vertical version: " + err.Error()))
						} else {
							fmt.Println(successStyle.Render("Vertical version created successfully"))
						}
					}

					if channel.HorizontalVideoBase != "" {
						fmt.Println(commandStyle.Render("Creating horizontal version..."))
						horizontalOutputDir := outputDir + "/horizontal-yt"
						if _, err := os.Stat(horizontalOutputDir); os.IsNotExist(err) {
							os.Mkdir(horizontalOutputDir, 0755)
						}

						horizontalOutputFileName := fmt.Sprintf("%s/%s.mp4", horizontalOutputDir, cut.Title)
						cmd := exec.Command(
							ffmpegPath,
							"-i", channel.HorizontalVideoBase,
							"-i", outputFileName,
							"-filter_complex", "[0:v]loop=loop=-1:size=1:start=0[loopbg];[1:v]scale=1080:-1[scaled];[loopbg][scaled]overlay=(W-w)/2:(H-h)/2:shortest=1[outv]",
							"-map", "[outv]",
							"-map", "1:a",
							"-c:a", "aac",
							"-c:v", "libx264",
							"-preset", "ultrafast",
							"-tune", "fastdecode",
							"-crf", "28",
							"-threads", "0",
							"-shortest",
							"-y",
							horizontalOutputFileName,
						)
						if err := cmd.Run(); err != nil {
							fmt.Println(errorStyle.Render("Error creating horizontal version: " + err.Error()))
						} else {
							fmt.Println(successStyle.Render("horizontal version created successfully"))
						}
					}

					// After processing the video, upload it to YouTube
					if channel.UploadToYouTube && metadata != nil {
						// Upload horizontal video
						fmt.Println(commandStyle.Render("Uploading horizontal video to YouTube..."))
						outputFileName := fmt.Sprintf("%s/horizontal-yt/%s.mp4", outputDir, cut.Title)
						err := UploadToYouTube(
							outputFileName,
							metadata.Title,
							metadata.Description,
							metadata.Tags,
							"unlisted",
						)

						if err != nil {
							fmt.Println(errorStyle.Render(fmt.Sprintf("YouTube upload failed: %v", err)))
						} else {
							fmt.Println(successStyle.Render("Video uploaded to YouTube successfully"))
						}

						// Upload vertical video if it exists
						verticalFileName := fmt.Sprintf("%s/vertical/%s.mp4", outputDir, cut.Title)
						if _, err := os.Stat(verticalFileName); err == nil {
							fmt.Println(commandStyle.Render("Uploading vertical video to YouTube..."))
							err := UploadToYouTube(
								verticalFileName,
								metadata.Title+" (Vertical)",
								metadata.Description,
								metadata.Tags,
								"unlisted",
							)

							if err != nil {
								fmt.Println(errorStyle.Render(fmt.Sprintf("Vertical video upload failed: %v", err)))
							} else {
								fmt.Println(successStyle.Render("Vertical video uploaded to YouTube successfully"))
							}
						}
					}
				}
			} else {
				fmt.Println(subtitleStyle.Render("No interesting cuts found in this segment"))
			}
		}

		if len(videoSegments) > 1 {
			fmt.Println(commandStyle.Render("Cleaning up temporary files..."))
			for _, file := range append(videoSegments, subtitleSegments...) {
				if file != videoFileName && file != subtitleFileName {
					os.Remove(file)
				}
			}
			fmt.Println(successStyle.Render("Cleanup completed"))
		}
	}

	fmt.Println(titleStyle.Render("Processing completed for channel: " + channel.Name))
}

type Cut struct {
	Title string `json:"title"`
	Begin int    `json:"begin"`
	End   int    `json:"end"`
}

type CutsResponse struct {
	Cuts []Cut `json:"cuts"`
}

func GetCuts(subtleFileName string, topics string, excerpts int, stretchTime int) []Cut {
	isSegment := strings.Contains(subtleFileName, ".part")

	var vttPath string
	if isSegment {
		basePath := strings.Split(subtleFileName, ".part")[0]
		vttPath = basePath + ".srt.pt.vtt"
	} else {
		vttPath = subtleFileName + ".pt.vtt"
	}

	subtleContent, err := ioutil.ReadFile(vttPath)
	if err != nil {
		log.Printf("Error reading subtitle file: %v", err)
		return nil
	}

	subtleContentString := string(subtleContent)

	url := "https://api.openai.com/v1/chat/completions"
	method := "POST"

	systemPrompt := fmt.Sprintf(`You are a professional video editor specialized in analyzing video subtitles and identifying compelling segments about the topics "%s".
	Your task is to locate multiple excerpts (at least %d, if possible) that contain relevant discussions about these topics.

	While each excerpt should target around %d minute(s) in length, you should prioritize natural cutting points where conversations
	or ideas reach logical conclusions. This means your cuts can be 1-2 minutes longer or shorter than the target time
	if that produces a better quality clip with complete thoughts and discussions.

	Focus on segments that are self-contained, meaningful, and engaging. Cut at natural conversational breaks, not mid-sentence.

	Return only a JSON object in the format: {"cuts": [{"title": "Descriptive title of the cut", "begin": start time in seconds (integer), "end": end time in seconds (integer)}]}`, topics, excerpts, stretchTime)

	userPrompt := fmt.Sprintf("Here is the subtitle file in WEBVTT format:\n\n%s\n\nIdentify multiple interesting segments related to the topics \"%s\". Target approximately %d minute(s) per segment, but prioritize natural cut points for complete thoughts. Return only the JSON object with the identified cuts.", subtleContentString, topics, stretchTime)

	requestBody := map[string]interface{}{
		"model": config.GetOpenAIModel(),
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": systemPrompt,
			},
			{
				"role":    "user",
				"content": userPrompt,
			},
		},
		"response_format": map[string]string{
			"type": "json_object",
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("Error creating request JSON: %v", err)
		return nil
	}

	payload := bytes.NewBuffer(jsonData)

	maxRetries := 3
	var apiResponse OpenAIResponse
	var respBody []byte
	var statusCode int

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			backoffDuration := time.Duration(2<<uint(attempt-1)) * time.Second
			time.Sleep(backoffDuration)
		}

		client := &http.Client{
			Timeout: 120 * time.Second,
		}
		req, err := http.NewRequest(method, url, payload)
		if err != nil {
			continue
		}
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer "+config.GetOpenAIKey())

		res, err := client.Do(req)
		if err != nil {
			continue
		}

		statusCode = res.StatusCode

		respBody, err = io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			continue
		}

		if statusCode != http.StatusOK {
			continue
		}

		if err := json.Unmarshal(respBody, &apiResponse); err != nil {
			continue
		}

		break
	}

	if statusCode != http.StatusOK {
		return nil
	}

	if len(apiResponse.Choices) == 0 {
		return nil
	}

	var cutsResponse CutsResponse
	if err := json.Unmarshal([]byte(apiResponse.Choices[0].Message.Content), &cutsResponse); err != nil {
		return nil
	}

	return cutsResponse.Cuts
}

type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type SubtitleEntry struct {
	Index     int
	StartTime time.Duration
	EndTime   time.Duration
	Text      string
}

func parseVTTFile(filePath string) ([]SubtitleEntry, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	var entries []SubtitleEntry
	var currentEntry SubtitleEntry
	var inEntry bool = false
	var index int = 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line == "WEBVTT" || strings.HasPrefix(line, "NOTE") {
			continue
		}

		if strings.Contains(line, "-->") {
			if inEntry {
				entries = append(entries, currentEntry)
			}

			inEntry = true
			index++
			currentEntry = SubtitleEntry{Index: index}

			timestamps := strings.Split(line, "-->")
			if len(timestamps) == 2 {
				currentEntry.StartTime = parseTimestamp(strings.TrimSpace(timestamps[0]))
				currentEntry.EndTime = parseTimestamp(strings.TrimSpace(timestamps[1]))
			}
			currentEntry.Text = ""
		} else if inEntry {
			if currentEntry.Text != "" {
				currentEntry.Text += " "
			}
			currentEntry.Text += line
		}
	}

	if inEntry {
		entries = append(entries, currentEntry)
	}

	return entries, nil
}

func parseTimestamp(timestamp string) time.Duration {
	if idx := strings.Index(timestamp, " "); idx != -1 {
		timestamp = timestamp[:idx]
	}

	timestamp = strings.ReplaceAll(timestamp, ",", ".")

	parts := strings.Split(timestamp, ":")
	if len(parts) < 3 {
		return 0
	}

	hours, _ := time.ParseDuration(parts[0] + "h")
	minutes, _ := time.ParseDuration(parts[1] + "m")

	secondParts := strings.Split(parts[2], ".")
	seconds, _ := time.ParseDuration(secondParts[0] + "s")

	var milliseconds time.Duration
	if len(secondParts) > 1 {
		ms := secondParts[1]
		for len(ms) < 3 {
			ms += "0"
		}
		if len(ms) > 3 {
			ms = ms[:3]
		}
		milliseconds, _ = time.ParseDuration(ms + "ms")
	}

	return hours + minutes + seconds + milliseconds
}

func cleanSubtitleText(text string) string {
	timestampPattern := regexp.MustCompile("<\\d{2}:\\d{2}:\\d{2}\\.\\d{3}>")
	cleanText := timestampPattern.ReplaceAllString(text, "")

	stylePattern := regexp.MustCompile("</?c>")
	cleanText = stylePattern.ReplaceAllString(cleanText, "")

	cleanText = strings.ReplaceAll(cleanText, "  ", " ")

	return strings.TrimSpace(cleanText)
}

func getSubtitlesForTimeRange(subtitleEntries []SubtitleEntry, startSeconds, endSeconds int) string {
	startTime := time.Duration(startSeconds) * time.Second
	endTime := time.Duration(endSeconds) * time.Second

	var subtitleText strings.Builder
	var index int = 1

	for _, entry := range subtitleEntries {
		if (entry.StartTime <= endTime) && (entry.EndTime >= startTime) {
			adjustedStart := int(math.Max(0, float64(entry.StartTime.Seconds()-float64(startSeconds))))
			adjustedEnd := int(math.Min(float64(endSeconds-startSeconds), float64(entry.EndTime.Seconds()-float64(startSeconds))))

			startStr := formatSRTTimestamp(adjustedStart)
			endStr := formatSRTTimestamp(adjustedEnd)

			cleanedText := cleanSubtitleText(entry.Text)

			subtitleText.WriteString(fmt.Sprintf("%d\n%s --> %s\n%s\n\n",
				index, startStr, endStr, cleanedText))
			index++
		}
	}

	return subtitleText.String()
}

// formatSRTTimestamp converts seconds to a properly formatted SRT timestamp string (HH:MM:SS,MMM)
func formatSRTTimestamp(seconds int) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60
	return fmt.Sprintf("%02d:%02d:%02d,000", hours, minutes, secs)
}

// VideoMetadata represents SEO metadata for a video cut
// Contains optimized information for publishing videos across multiple platforms
type VideoMetadata struct {
	Title       string   `json:"title"`       // SEO-optimized title for the video
	Description string   `json:"description"` // Short engaging description (max 250 chars)
	Tags        []string `json:"tags"`        // Relevant search tags without # symbol
	Hashtags    []string `json:"hashtags"`    // Popular hashtags with # symbol included
}

// GenerateMetadata generates optimized SEO metadata using AI based on video content
// Parameters:
//   - videoTitle: The original title of the video clip
//   - subtitleContent: The transcript text from the video
//   - topics: The main topics or themes to focus on
//
// Returns SEO-optimized metadata or an error if generation fails
func GenerateMetadata(videoTitle string, subtitleContent string, topics string) (*VideoMetadata, error) {
	url := "https://api.openai.com/v1/chat/completions"
	method := "POST"

	systemPrompt := fmt.Sprintf(`You are an expert in SEO for YouTube, TikTok, and Instagram videos.
	Your task is to create optimized metadata for a video clip about "%s".
	Generate an attractive title, an engaging description limited to 250 characters, up to 10 relevant tags, and 5 popular hashtags.

	IMPORTANT: Keep the language of your output THE SAME as the language used in the subtitle excerpt.
	DO NOT translate to English - maintain the original language of the subtitles.`, topics)

	userPrompt := fmt.Sprintf(`Based on this subtitle excerpt:
	"%s"

	And with this original title: "%s"

	Create SEO-optimized metadata in JSON format with the following fields:
	1. title: An attractive SEO-optimized title (keep in the SAME LANGUAGE as the subtitle)
	2. description: An engaging description up to 250 characters (keep in the SAME LANGUAGE as the subtitle)
	3. tags: List of up to 10 relevant tags (without the # symbol, keep in the SAME LANGUAGE as the subtitle)
	4. hashtags: List of 5 popular hashtags (including the # symbol, keep in the SAME LANGUAGE as the subtitle)`, subtitleContent, videoTitle)

	requestBody := map[string]interface{}{
		"model": config.GetOpenAIModel(),
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": systemPrompt,
			},
			{
				"role":    "user",
				"content": userPrompt,
			},
		},
		"response_format": map[string]string{
			"type": "json_object",
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request JSON: %v", err)
	}

	payload := bytes.NewBuffer(jsonData)

	maxRetries := 3
	var metadata VideoMetadata
	var respBody []byte
	var statusCode int

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			backoffDuration := time.Duration(2<<uint(attempt-1)) * time.Second
			time.Sleep(backoffDuration)
		}

		client := &http.Client{
			Timeout: 60 * time.Second,
		}
		req, err := http.NewRequest(method, url, payload)
		if err != nil {
			continue
		}
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer "+config.GetOpenAIKey())

		res, err := client.Do(req)
		if err != nil {
			continue
		}

		statusCode = res.StatusCode

		respBody, err = io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			continue
		}

		if statusCode != http.StatusOK {
			continue
		}

		var apiResponse struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}

		if err := json.Unmarshal(respBody, &apiResponse); err != nil {
			continue
		}

		if len(apiResponse.Choices) == 0 {
			continue
		}

		if err := json.Unmarshal([]byte(apiResponse.Choices[0].Message.Content), &metadata); err != nil {
			continue
		}

		break
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status code %d", statusCode)
	}

	return &metadata, nil
}

// UploadToYouTube uploads a video to YouTube using saved credentials
func UploadToYouTube(videoPath, title, description string, tags []string, privacy string) error {
	// Get authentication token
	token, err := auth.GetClient()
	if err != nil {
		return fmt.Errorf("error getting authentication token: %v", err)
	}

	ctx := context.Background()
	tokenSource := oauth2.StaticTokenSource(token)
	client := oauth2.NewClient(ctx, tokenSource)

	// Create YouTube service
	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("error creating YouTube service: %v", err)
	}

	// Open video file
	file, err := os.Open(videoPath)
	if err != nil {
		return fmt.Errorf("error opening video file: %v", err)
	}
	defer file.Close()

	// Set default privacy to unlisted if not specified
	if privacy == "" {
		privacy = "unlisted"
	}

	// Configure video metadata
	upload := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       title,
			Description: description,
			Tags:        tags,
			CategoryId:  "22", // Category "People & Blogs" - can be adjusted as needed
		},
		Status: &youtube.VideoStatus{
			PrivacyStatus: privacy,
		},
	}

	// Execute upload
	call := service.Videos.Insert([]string{"snippet", "status"}, upload)
	call = call.Media(file)
	_, err = call.Do()
	if err != nil {
		return fmt.Errorf("error uploading video: %v", err)
	}

	log.Printf("Video '%s' successfully uploaded to YouTube", title)
	return nil
}
