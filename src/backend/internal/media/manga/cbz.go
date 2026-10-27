// Package manga reads manga archives and parses manga filename metadata.
package manga

import (
	"archive/zip"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
)

// Archive owns an open CBZ file. Its caller must close it after all page readers.
type Archive struct {
	FilePath  string  `json:"file_path"`
	Title     string  `json:"title"`
	PageCount int     `json:"page_count"`
	Pages     []*Page `json:"pages"`

	closer *zip.ReadCloser
}

type Page struct {
	Index    int    `json:"index"`
	FileName string `json:"file_name"`
	Size     uint64 `json:"size"`

	source *zip.File
}

// Open opens an independent CBZ archive whose lifetime belongs to the caller.
func Open(path string) (*Archive, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}

	archive := &Archive{
		FilePath: path,
		Title:    strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
		closer:   r,
		Pages:    make([]*Page, 0),
	}
	var validFiles []*zip.File
	for _, f := range r.File {
		if isImageFile(f.Name) {
			validFiles = append(validFiles, f)
		}
	}
	sort.Slice(validFiles, func(i, j int) bool {
		return validFiles[i].Name < validFiles[j].Name
	})
	for i, f := range validFiles {
		archive.Pages = append(archive.Pages, &Page{
			Index:    i,
			FileName: f.Name,
			Size:     f.UncompressedSize64,
			source:   f,
		})
	}

	archive.PageCount = len(archive.Pages)
	return archive, nil
}

// OpenPage returns a page reader that the caller must close before the archive.
func (a *Archive) OpenPage(pageIndex int) (io.ReadCloser, error) {
	if pageIndex < 0 || pageIndex >= len(a.Pages) {
		return nil, fmt.Errorf("page index out of bounds")
	}

	return a.Pages[pageIndex].source.Open()
}

func (a *Archive) Close() error {
	if a == nil || a.closer == nil {
		return nil
	}
	return a.closer.Close()
}

func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp"
}
