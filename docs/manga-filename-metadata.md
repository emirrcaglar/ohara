# Manga filename metadata

New CBZ imports use the pinned `github.com/emirrcaglar/chaptr-go/chaptr` parser in the shared scanner import path. Uploads and directory scans therefore record the same filename metadata. Archive reading and page caching do not run the parser.

The manga table retains its original filename-derived `title` and adds `series_title`, `metadata_json`, and `parser_version`. The list and detail APIs expose these as `seriesTitle`, `metadata`, and `parserVersion`, alongside `fileName`. Metadata includes volume/chapter ranges, group, source, language, revision, edition, and oneshot status. An empty parsed title falls back to the filename without its extension.

Library cards and reading history show the series title with release information underneath. Cards expose the original filename, group, and language in expandable details. Manga search matches the original filename and title as well as parsed release information. Catalog placement remains independent of the parsed series title.

Library sorting uses the series title, volume range, chapter range, and original filename. Number comparisons preserve chaptr's whole and optional integer decimal components: a missing decimal sorts before explicit `.0`, and `.5` sorts before `.10`. Ranges retain both endpoints. No floating point conversion is used.

Schema migration adds empty metadata defaults without parsing existing records. Normal scans continue skipping indexed paths, and importing an existing path preserves its metadata, catalog, ID, and reading progress. There is no backfill or refresh operation.

When updating the parser dependency, update `ParserVersion` in `internal/media/manga/metadata.go` to match the version pinned in `go.mod`.
