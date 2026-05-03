package logic

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	ffprobeTimeout = 25 * time.Second
	ffmpegTimeout  = 60 * time.Second
)

func probeDuration(parent context.Context, audioURL string) (float64, error) {
	ctx, cancel := context.WithTimeout(parent, ffprobeTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		audioURL,
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if ctx.Err() == context.DeadlineExceeded {
		return 0, errors.New("ffprobe timed out")
	}
	if err != nil {
		return 0, fmt.Errorf("%v: %s", err, strings.TrimSpace(stderr.String()))
	}

	duration, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil || duration <= 0 {
		return 0, fmt.Errorf("invalid ffprobe duration %q", strings.TrimSpace(string(out)))
	}
	return duration, nil
}

func cutAudio(parent context.Context, audioURL string, start, length float64) ([]byte, error) {
	ctx, cancel := context.WithTimeout(parent, ffmpegTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-hide_banner",
		"-loglevel", "error",
		"-ss", formatSeconds(start),
		"-i", audioURL,
		"-t", formatSeconds(length),
		"-vn",
		"-f", "mp3",
		"-codec:a", "libmp3lame",
		"-b:a", "192k",
		"pipe:1",
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if ctx.Err() == context.DeadlineExceeded {
		return nil, errors.New("ffmpeg timed out")
	}
	if err != nil {
		return nil, fmt.Errorf("%v: %s", err, strings.TrimSpace(stderr.String()))
	}
	if len(out) == 0 {
		return nil, errors.New("ffmpeg produced empty audio")
	}
	return out, nil
}
