# Homepage

The root route now opens a personal reading homepage. The library remains at `/library`, and the sidebar links to both pages.

The homepage shows the current user's 24 most recently read manga chapters, ordered by saved progress time. The most recent chapter is featured above the earlier chapters. Each entry includes its first-page cover, last-read time, page progress, and a link that resumes at the saved page. Chapters at their final page remain in the history.

The homepage fetches the existing authenticated `GET /api/manga` library response. Each manga now includes `updatedAt`, exposing the current user's existing `manga_progress.updated_at` timestamp as RFC 3339, or `null` when no progress has been saved. The homepage filters entries without a valid timestamp and empty archives, sorts by timestamp descending (then manga ID for ties), and displays the first 24. The library's existing ordering is unchanged. History refreshes whenever the homepage mounts.

The database schema and progress-save behavior are unchanged. Normal page progress saves already update the timestamp used for recency. Opening a chapter without changing pages does not create or refresh a progress record. Stored page indices remain zero-based; the homepage displays page numbers starting at one. No separate history endpoint or additional save requests are needed.

Loading, retry, empty-history, and unavailable-cover states follow the existing sharp geometry, surface colors, and orange/pink accents described in [design-principles.md](design-principles.md).
