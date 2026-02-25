package reader

import (
	"archive/zip"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
)

type Page struct {
	Index    int    `json:"index"`
	FileName string `json:"file_name"`
	Size     uint64 `json:"size"`

	source *zip.File
}

type Manga struct {
	FilePath  string  `json:"file_path"`
	Title     string  `json:"title"`
	PageCount int     `json:"page_count"`
	Pages     []*Page `json:"pages"`

	closer *zip.ReadCloser
}

func (c *Manga) GetPageReader(pageIndex int) (io.ReadCloser, error) {
	if pageIndex < 0 || pageIndex >= len(c.Pages) {
		return nil, fmt.Errorf("page index out of bounds")
	}

	return c.Pages[pageIndex].source.Open()
}

func (c *Manga) Close() error {
	return c.closer.Close()
}

func Open(path string) (*Manga, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}

	manga := &Manga{
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
		manga.Pages = append(manga.Pages, &Page{
			Index:    i,
			FileName: f.Name,
			Size:     f.UncompressedSize64,
			source:   f,
		})
	}

	manga.PageCount = len(manga.Pages)
	return manga, nil
}

func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp"
}
