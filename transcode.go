package main

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
)

// GenerateThumbnail generates a thumbnail from a video file using ffmpeg
func GenerateThumbnail(videoPath string, outputPath string, timeOffset string) error {
	if timeOffset == "" {
		timeOffset = "00:00:01"
	}

	cmd := exec.Command("ffmpeg",
		"-i", videoPath,
		"-ss", timeOffset,
		"-vframes", "1",
		"-vf", "scale=320:-1",
		outputPath,
		"-y", // Overwrite output file
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg error: %v, output: %s", err, string(output))
	}

	log.Printf("Thumbnail generated: %s", outputPath)
	return nil
}

// TranscodeVideo transcodes video to a web-friendly format
func TranscodeVideo(inputPath string, outputPath string) error {
	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-c:v", "libx264", // H.264 codec
		"-preset", "medium", // Encoding speed
		"-crf", "23", // Quality (lower = better)
		"-c:a", "aac", // AAC audio
		"-b:a", "128k", // Audio bitrate
		"-movflags", "+faststart", // Enable progressive download
		outputPath,
		"-y",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg error: %v, output: %s", err, string(output))
	}

	log.Printf("Video transcoded: %s", outputPath)
	return nil
}

// GetVideoDuration gets the duration of a video file in seconds
func GetVideoDuration(videoPath string) (float64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		videoPath,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe error: %v", err)
	}

	var duration float64
	_, err = fmt.Sscanf(string(output), "%f", &duration)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %v", err)
	}

	return duration, nil
}

// ProcessUploadedVideo processes a newly uploaded video
func ProcessUploadedVideo(videoFilename string) error {
	videoPath := filepath.Join("videos", videoFilename)

	// Generate thumbnail
	thumbnailFilename := videoFilename[:len(videoFilename)-len(filepath.Ext(videoFilename))] + ".jpg"
	thumbnailPath := filepath.Join("thumbnails", thumbnailFilename)

	if err := GenerateThumbnail(videoPath, thumbnailPath, "00:00:05"); err != nil {
		log.Printf("Warning: Could not generate thumbnail: %v", err)
	}

	// Get duration
	duration, err := GetVideoDuration(videoPath)
	if err != nil {
		log.Printf("Warning: Could not get video duration: %v", err)
	} else {
		log.Printf("Video duration: %.2f seconds", duration)
	}

	return nil
}
