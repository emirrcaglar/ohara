package manga

import (
	"path/filepath"
	"strings"

	"github.com/emirrcaglar/chaptr-go/chaptr"
)

// ParserVersion identifies the pinned parser used when importing a filename.
const ParserVersion = "chaptr-go/v0.0.0-20261027111849-91442a122b80"

// Number preserves chaptr's integer decimal component, including explicit .0.
type Number struct {
	Whole   uint32  `json:"whole"`
	Decimal *uint16 `json:"decimal"`
}

type NumberRange struct {
	Start Number  `json:"start"`
	End   *Number `json:"end"`
}

type Metadata struct {
	Volume    *NumberRange `json:"volume"`
	Chapter   *NumberRange `json:"chapter"`
	Group     string       `json:"group"`
	Source    string       `json:"source"`
	Language  string       `json:"language"`
	Revision  *uint8       `json:"revision"`
	IsOneshot bool         `json:"isOneshot"`
	Edition   string       `json:"edition"`
}

func ParseFilename(path string) (string, Metadata) {
	filename := filepath.Base(path)
	parsed := chaptr.ParseManga(filename)
	title := strings.TrimSpace(parsed.Title)
	if title == "" {
		title = strings.TrimSuffix(filename, filepath.Ext(filename))
	}
	return title, Metadata{
		Volume: numberRange(parsed.Volume), Chapter: numberRange(parsed.Chapter),
		Group: parsed.Group, Source: string(parsed.Source), Language: string(parsed.Language),
		Revision: parsed.Revision, IsOneshot: parsed.IsOneshot, Edition: parsed.Edition,
	}
}

func numberRange(value *chaptr.NumberRange) *NumberRange {
	if value == nil {
		return nil
	}
	result := &NumberRange{Start: Number{Whole: value.Start.Whole, Decimal: value.Start.Decimal}}
	if value.End != nil {
		result.End = &Number{Whole: value.End.Whole, Decimal: value.End.Decimal}
	}
	return result
}
