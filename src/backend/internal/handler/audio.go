package handler

import (
	"fmt"
	"html"
	"net/http"
	"os"
	"strconv"
	"strings"

	"ohara/src/internal/db"
)

type AudioHandler struct {
	DB *db.DB
}

func formatDuration(seconds int) string {
	m := seconds / 60
	s := seconds % 60
	return fmt.Sprintf("%d:%02d", m, s)
}

func (h *AudioHandler) HandleAudioList(w http.ResponseWriter, r *http.Request) {
	tracks, err := h.DB.ListAudio()
	if err != nil {
		http.Error(w, "Failed to load audio library", http.StatusInternalServerError)
		return
	}

	var cards strings.Builder
	for _, t := range tracks {
		artist := t.Artist
		if artist == "" {
			artist = "Unknown Artist"
		}
		cards.WriteString(fmt.Sprintf(`
		<div class="audio-card" data-id="%d" data-title="%s" data-artist="%s">
			<div class="audio-icon">
				<span>♪</span>
				<span class="duration-badge">%s</span>
			</div>
			<span class="title">%s</span>
			<span class="artist">%s</span>
		</div>`,
			t.ID,
			html.EscapeString(t.Title),
			html.EscapeString(artist),
			formatDuration(t.Duration),
			html.EscapeString(t.Title),
			html.EscapeString(artist),
		))
	}

	page := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Music - Ohara</title>
	<link rel="stylesheet" href="/static/style.css">
	<style>
		body { padding: 20px; padding-bottom: 90px; }
		.grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(160px, 1fr)); gap: 16px; }
		.audio-card { display: flex; flex-direction: column; align-items: center; color: white; background: #1e1e1e; border-radius: 8px; overflow: hidden; cursor: pointer; transition: transform 0.15s; }
		.audio-card:hover { transform: scale(1.04); }
		.audio-card.active { outline: 2px solid #888; background: #2a2a2a; }
		.audio-icon { position: relative; width: 100%%; aspect-ratio: 1/1; background: #2a2a2a; display: flex; align-items: center; justify-content: center; font-size: 3rem; }
		.duration-badge { position: absolute; bottom: 0; left: 0; right: 0; background: rgba(0,0,0,0.65); color: #ccc; font-size: 0.75rem; text-align: center; padding: 3px 0; }
		.title { padding: 6px 8px 2px; font-size: 0.85rem; text-align: center; word-break: break-word; }
		.artist { padding: 0 8px 8px; font-size: 0.75rem; color: #888; text-align: center; word-break: break-word; }
		.player { position: fixed; bottom: 0; left: 0; right: 0; background: #1a1a1a; border-top: 1px solid #333; padding: 12px 20px; display: none; align-items: center; gap: 16px; z-index: 100; }
		.player.visible { display: flex; }
		.player-info { flex: 1; min-width: 0; }
		.player-title { font-size: 0.9rem; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
		.player-artist { font-size: 0.75rem; color: #888; }
		.player audio { flex: 2; min-width: 0; max-width: 60%%; }
		audio { width: 100%%; accent-color: white; }
	</style>
</head>
<body>
	<a class="nav-link" href="/">← Home</a>
	<div class="grid">%s</div>
	<div class="player" id="player">
		<div class="player-info">
			<div class="player-title" id="player-title"></div>
			<div class="player-artist" id="player-artist"></div>
		</div>
		<audio id="audio-el" controls></audio>
	</div>
	<script>
		document.querySelectorAll('.audio-card').forEach(card => {
			card.addEventListener('click', () => {
				document.querySelectorAll('.audio-card').forEach(c => c.classList.remove('active'));
				card.classList.add('active');
				const player = document.getElementById('player');
				player.classList.add('visible');
				document.getElementById('player-title').textContent = card.dataset.title;
				document.getElementById('player-artist').textContent = card.dataset.artist;
				const audio = document.getElementById('audio-el');
				audio.src = '/audio/' + card.dataset.id + '/stream';
				audio.play();
			});
		});
	</script>
</body>
</html>`, cards.String())

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(page))
}

func (h *AudioHandler) HandleAudioStream(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	track, err := h.DB.GetAudioByID(id)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if track == nil {
		http.Error(w, "Track not found", http.StatusNotFound)
		return
	}

	file, err := os.Open(track.Path)
	if err != nil {
		fmt.Printf("[audio] failed to open file %s: %v\n", track.Path, err)
		http.Error(w, "File not found on disk", http.StatusNotFound)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "Could not stat file", http.StatusInternalServerError)
		return
	}

	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
}
