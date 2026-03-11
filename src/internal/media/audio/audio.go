package audio

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dhowden/tag"
)

type Track struct {
	FilePath string
	Title    string
	Artist   string
	Album    string
	Duration int
}

func Open(path string) (*Track, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	m, err := tag.ReadFrom(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read tags for %s: %w", path, err)
	}

	title := m.Title()
	if title == "" {
		base := path[strings.LastIndexAny(path, "/\\")+1:]
		title = strings.TrimSuffix(base, filepath.Ext(base))
	}

	duration := 0
	if d, err := getDuration(path); err == nil {
		duration = int(d)
	}

	return &Track{
		FilePath: path,
		Title:    title,
		Artist:   m.Artist(),
		Album:    m.Album(),
		Duration: duration,
	}, nil
}

func getDuration(path string) (float64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	)

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return 0, err
	}

	cleanStr := strings.TrimSpace(out.String())
	return strconv.ParseFloat(cleanStr, 64)
}
