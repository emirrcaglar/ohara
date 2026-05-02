package cbz

import (
	"archive/zip"
	"fmt"
	"io"
	"ohara/src/internal/db"
	"path/filepath"
	"sort"
	"strings"
)

type CBZService struct {
	db  *db.DB
	CBZ *CBZ
}

type ICBZService interface {
	SaveCBZ(cbz *CBZ) error
	GetPageReader(pageIndex int) (io.ReadCloser, error)
	Open(path string) (*CBZ, error)
	Close() error
}

func NewCBZService(db *db.DB) *CBZService {
	return &CBZService{db: db}
}

type CBZ struct {
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

func (s *CBZService) GetPageReader(pageIndex int) (io.ReadCloser, error) {
	return s.CBZ.GetPageReader(pageIndex)
}

func (s *CBZService) Close() error {
	if s.CBZ == nil {
		return nil
	}
	err := s.CBZ.Close()
	s.CBZ = nil
	return err
}

func (s *CBZService) Open(path string) (*CBZ, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}

	cbz := &CBZ{
		FilePath: path,
		Title:    strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
		closer:   r,
		Pages:    make([]*Page, 0),
	}
	s.CBZ = cbz
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
		cbz.Pages = append(cbz.Pages, &Page{
			Index:    i,
			FileName: f.Name,
			Size:     f.UncompressedSize64,
			source:   f,
		})
	}

	cbz.PageCount = len(cbz.Pages)
	return cbz, nil
}

func (s *CBZService) SaveCBZ(cbz *CBZ) error {
	return nil
}

func (cbz *CBZ) GetPageReader(pageIndex int) (io.ReadCloser, error) {
	if pageIndex < 0 || pageIndex >= len(cbz.Pages) {
		return nil, fmt.Errorf("page index out of bounds")
	}

	return cbz.Pages[pageIndex].source.Open()
}

func (cbz *CBZ) Close() error {
	if cbz == nil || cbz.closer == nil {
		return nil
	}
	return cbz.closer.Close()
}

func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp"
}
