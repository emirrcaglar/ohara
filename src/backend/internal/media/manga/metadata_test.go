package manga

import (
	"path/filepath"
	"testing"
)

func TestParseFilename(t *testing.T) {
	title, metadata := ParseFilename(filepath.Join("library", "[Group] Title v03 c042.5 (Digital).cbz"))
	if title != "Title" || metadata.Group != "Group" || metadata.Source != "Digital" {
		t.Fatalf("unexpected parsed release: %q %+v", title, metadata)
	}
	if metadata.Volume == nil || metadata.Volume.Start.Whole != 3 || metadata.Chapter == nil ||
		metadata.Chapter.Start.Whole != 42 || metadata.Chapter.Start.Decimal == nil || *metadata.Chapter.Start.Decimal != 5 {
		t.Fatalf("lost chapter/volume components: %+v", metadata)
	}
	_, batch := ParseFilename("Title c001-050.cbz")
	if batch.Chapter == nil || batch.Chapter.End == nil || batch.Chapter.End.Whole != 50 {
		t.Fatalf("lost chapter range: %+v", batch.Chapter)
	}
	_, zero := ParseFilename("Title c000.0.cbz")
	if zero.Chapter == nil || zero.Chapter.Start.Whole != 0 || zero.Chapter.Start.Decimal == nil || *zero.Chapter.Start.Decimal != 0 {
		t.Fatalf("lost explicit zero: %+v", zero.Chapter)
	}
	title, _ = ParseFilename("[Group].cbz")
	if title != "[Group]" {
		t.Fatalf("expected filename fallback, got %q", title)
	}
}
