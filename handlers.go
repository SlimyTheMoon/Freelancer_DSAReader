package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// --- HTML TEMPLATE ---
var htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>Freelancer Log Reader // SYSTEM</title>
	<style>
		:root {
			--bg-primary: #0a0e17;
			--bg-secondary: #0f1520;
			--bg-panel: #141c2e;
			--cyan: #00f0ff;
			--cyan-dim: #00a0b0;
			--magenta: #ff00ff;
			--green: #00ff88;
			--red: #ff3366;
			--yellow: #ffcc00;
			--text-primary: #e0f0ff;
			--text-dim: #6a8faa;
			--border-glow: 0 0 10px rgba(0, 240, 255, 0.3);
		}
		
		* { box-sizing: border-box; }
		
		body {
			background-color: var(--bg-primary);
			background-image: 
				linear-gradient(rgba(0, 240, 255, 0.03) 1px, transparent 1px),
				linear-gradient(90deg, rgba(0, 240, 255, 0.03) 1px, transparent 1px);
			background-size: 50px 50px;
			color: var(--text-primary);
			font-family: 'Courier New', Consolas, monospace;
			margin: 0; padding: 0; font-size: 13px;
			min-height: 100vh;
		}

		/* Scanline effect */
		body::after {
			content: "";
			position: fixed; top: 0; left: 0; width: 100%; height: 100%;
			background: repeating-linear-gradient(
				0deg, rgba(0,0,0,0.1) 0px, rgba(0,0,0,0.1) 1px, transparent 1px, transparent 2px);
			pointer-events: none; z-index: 9999;
		}

		/* Navbar */
		#controls {
			position: fixed; top: 0; left: 0; right: 0; height: 55px;
			background: linear-gradient(180deg, var(--bg-panel) 0%, rgba(20, 28, 46, 0.95) 100%);
			border-bottom: 2px solid var(--cyan);
			display: flex; align-items: center; padding: 0 25px; gap: 20px;
			z-index: 1000;
			box-shadow: 0 0 20px rgba(0, 240, 255, 0.2), inset 0 1px 0 rgba(0, 240, 255, 0.1);
		}
		
		#controls::before {
			content: "●"; color: var(--cyan); font-size: 10px; animation: pulse 2s infinite;
		}

		@keyframes pulse {
			0%, 100% { opacity: 1; } 50% { opacity: 0.3; }
		}

		/* Main Content */
		#log-container {
			margin-top: 65px; padding: 25px;
			white-space: pre-wrap; word-break: break-word; line-height: 1.5;
			min-height: calc(100vh - 65px);
		}
		.log-entry { margin-bottom: 1px; font-size: 12px; letter-spacing: 0.3px; }

		/* Buttons */
		button {
			padding: 10px 18px; cursor: pointer; font-size: 11px; font-weight: bold;
			border: 1px solid var(--cyan); background: transparent; color: var(--cyan);
			font-family: 'Courier New', monospace; text-transform: uppercase; 
			transition: all 0.2s; letter-spacing: 1px;
			box-shadow: var(--border-glow);
		}
		button:hover {
			background: var(--cyan); color: var(--bg-primary);
			box-shadow: 0 0 20px rgba(0, 240, 255, 0.5);
		}
		
		#btn-action { 
			min-width: 130px; 
			border-color: var(--red); color: var(--red);
		}
		#btn-action:hover {
			background: var(--red); color: var(--bg-primary);
			box-shadow: 0 0 20px rgba(255, 51, 102, 0.5);
		}
		#btn-action.resume { 
			border-color: var(--green); color: var(--green);
		}
		#btn-action.resume:hover {
			background: var(--green); color: var(--bg-primary);
			box-shadow: 0 0 20px rgba(0, 255, 136, 0.5);
		}

		.nav-btn { border-color: var(--text-dim); color: var(--text-dim); }
		.nav-btn:hover { border-color: var(--cyan); color: var(--cyan); background: rgba(0, 240, 255, 0.1); }

		#btn-export { 
			border-color: var(--magenta); color: var(--magenta);
		}
		#btn-export:hover {
			background: var(--magenta); color: var(--bg-primary);
			box-shadow: 0 0 20px rgba(255, 0, 255, 0.5);
		}

		.right-aligned { margin-left: auto; display: flex; gap: 12px; }

		/* Status */
		#status { 
			font-weight: bold; color: var(--green); 
			text-shadow: 0 0 5px rgba(0, 255, 136, 0.5);
			letter-spacing: 2px;
		}
		
		.paused-mode { background-color: rgba(255, 51, 102, 0.05); }
		.paused-cue { border-left: 3px solid var(--red); padding-left: 15px; }

		/* File Display */
		#file-display { 
			font-size: 11px; color: var(--text-dim); margin-left: 20px; 
			overflow: hidden; white-space: nowrap; text-overflow: ellipsis; 
			max-width: 350px; font-family: 'Courier New', monospace;
		}
		#file-display::before { content: "📂 "; opacity: 0.5; }

		/* Modal */
		.modal-overlay {
			display: none; position: fixed; top: 0; left: 0; width: 100%; height: 100%;
			background: rgba(10, 14, 23, 0.95); z-index: 2000; 
			align-items: center; justify-content: center;
			backdrop-filter: blur(5px);
		}
		.modal-box {
			background: var(--bg-panel); padding: 35px; 
			border: 2px solid var(--cyan);
			width: 550px; max-width: 90%;
			box-shadow: 0 0 30px rgba(0, 240, 255, 0.2), inset 0 0 50px rgba(0, 240, 255, 0.05);
			position: relative;
		}
		.modal-box::before {
			content: ""; position: absolute; top: -2px; left: -2px; right: -2px; bottom: -2px;
			background: linear-gradient(45deg, var(--cyan), transparent, var(--cyan));
			z-index: -1; opacity: 0.3;
		}
		h2 { 
			margin-top: 0; color: var(--cyan); 
			text-transform: uppercase; letter-spacing: 3px;
			text-shadow: 0 0 10px rgba(0, 240, 255, 0.5);
			border-bottom: 1px solid var(--cyan-dim); padding-bottom: 15px;
		}
		h2::before { content: "► "; opacity: 0.5; }
		
		label { 
			display: block; margin-top: 20px; color: var(--text-dim); 
			font-size: 11px; text-transform: uppercase; letter-spacing: 1px;
		}
		input[type="text"] {
			width: 100%; padding: 12px 15px; margin-top: 8px; 
			background: rgba(0, 0, 0, 0.3);
			border: 1px solid var(--cyan-dim); color: var(--cyan);
			font-family: 'Courier New', monospace; font-size: 13px;
			box-sizing: border-box; letter-spacing: 1px;
		}
		input[type="text"]:focus {
			outline: none; border-color: var(--cyan);
			box-shadow: 0 0 15px rgba(0, 240, 255, 0.3);
		}
		input[type="text"]::placeholder { color: var(--text-dim); opacity: 0.5; }
		
		.modal-buttons { 
			display: flex; justify-content: flex-end; gap: 12px; margin-top: 25px;
			padding-top: 20px; border-top: 1px solid rgba(0, 240, 255, 0.2);
		}
		
		/* Log entries colors */
		.log-entry { color: var(--text-dim); }
		.log-entry:nth-child(odd) { color: var(--text-primary); }
		
		/* Welcome message */
		.welcome-box {
			border: 1px dashed var(--cyan-dim); padding: 25px;
			background: rgba(0, 240, 255, 0.03);
			margin-top: 20px;
		}
		.welcome-box b { color: var(--cyan); }
	</style>
</head>
<body>
	<div id="controls">
		<button id="btn-action" onclick="togglePause()">PAUSE SYSTEM</button>
		<div id="status">● LIVE</div>
		<div id="file-display">
			{{if .LogPath}}{{.LogPath}}{{else}}NO DATA STREAM{{end}}
		</div>

		<div class="right-aligned">
			<button id="btn-export" onclick="openExport()">⬇ EXPORT</button>
			<button class="nav-btn" onclick="openSettings()">⚙ CONFIG</button>
		</div>
	</div>

	<div id="log-container">
		{{if not .LogPath}}
			<div class="welcome-box">
				<span style="color: var(--cyan);">⚠ SYSTEM INITIALIZATION REQUIRED</span><br><br>
				No data stream configured. Click <b>CONFIG</b> to establish connection.
			</div>
		{{end}}
	</div>

	<div id="settings-modal" class="modal-overlay">
		<div class="modal-box">
			<h2>System Configuration</h2>
			<label>Data Source Path</label>
			<input type="text" id="path-input" value="{{.LogPath}}" placeholder="C:\...\DSAce.log">
			<div id="save-msg" style="height:20px; font-size:12px; margin-top:8px;"></div>
			<div class="modal-buttons">
				<button onclick="closeModals()" style="border-color: var(--text-dim); color: var(--text-dim);">ABORT</button>
				<button onclick="saveSettings()" style="border-color: var(--cyan); color: var(--cyan);">SAVE & REBOOT</button>
			</div>
		</div>
	</div>

	<div id="export-modal" class="modal-overlay">
		<div class="modal-box">
			<h2>Data Export</h2>
			<p style="color: var(--text-dim); font-size: 12px;">Specify temporal parameters for extraction.</p>

			<label>Timestamp Start</label>
			<input type="text" id="export-start" placeholder="DD.MM.YYYY HH:MM:SS">

			<label>Timestamp End</label>
			<input type="text" id="export-end" placeholder="DD.MM.YYYY HH:MM:SS">
			
			<div class="modal-buttons">
				<button onclick="closeModals()" style="border-color: var(--text-dim); color: var(--text-dim);">ABORT</button>
				<button onclick="performExport()" style="border-color: var(--magenta); color: var(--magenta);">INITIATE DOWNLOAD</button>
			</div>
		</div>
	</div>

	<script>
		const container = document.getElementById('log-container');
		const btn = document.getElementById('btn-action');
		const statusLabel = document.getElementById('status');
		const settingsModal = document.getElementById('settings-modal');
		const exportModal = document.getElementById('export-modal');
		const pathInput = document.getElementById('path-input');
		const saveMsg = document.getElementById('save-msg');

		let isPaused = false;
		let messageBuffer = [];
		let autoScroll = true;

		// --- MODAL LOGIC ---
		function openSettings() { settingsModal.style.display = 'flex'; pathInput.focus(); }
		function openExport() { exportModal.style.display = 'flex'; document.getElementById('export-start').focus(); }
		function closeModals() {
			settingsModal.style.display = 'none';
			exportModal.style.display = 'none';
			saveMsg.textContent = '';
		}

		function saveSettings() {
			const newPath = pathInput.value.trim();
			if (!newPath) return;
			saveMsg.textContent = "Verifying path...";
			saveMsg.style.color = "#aaa";
			fetch('/config', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ log_path: newPath })
			})
			.then(response => {
				if (response.ok) {
					saveMsg.textContent = "Saved! Reloading...";
					saveMsg.style.color = "var(--success)";
					setTimeout(() => location.reload(), 500);
				} else {
					return response.text().then(text => { throw new Error(text) });
				}
			})
			.catch(err => {
				saveMsg.textContent = "Error: " + err.message;
				saveMsg.style.color = "var(--error)";
			});
		}

		function performExport() {
			const start = document.getElementById('export-start').value.trim();
			const end = document.getElementById('export-end').value.trim();

			// Build URL with query params
			const params = new URLSearchParams();
			if (start) params.append('start', start);
			if (end) params.append('end', end);

			// Trigger download
			window.location.href = '/download?' + params.toString();
			closeModals();
		}

		// --- SSE LOGIC ---
		const evtSource = new EventSource("/stream");

		evtSource.onmessage = function(event) {
			if (event.data === "HEARTBEAT") return;
			if (isPaused) {
				messageBuffer.push(event.data);
				statusLabel.textContent = "● PAUSED [" + messageBuffer.length + " queued]";
				statusLabel.style.color = "var(--yellow)";
			} else {
				appendLine(event.data);
			}
		};

		evtSource.onerror = function() {
			statusLabel.textContent = "● CONNECTION LOST";
			statusLabel.style.color = "var(--red)";
		};

		function appendLine(text) {
			const div = document.createElement('div');
			div.className = 'log-entry';
			div.textContent = text;
			container.appendChild(div);
			if (autoScroll) window.scrollTo(0, document.body.scrollHeight);
		}

		function togglePause() {
			isPaused = !isPaused;
			if (isPaused) {
				autoScroll = false;
				btn.textContent = "RESUME";
				btn.className = "resume";
				statusLabel.textContent = "● PAUSED";
				statusLabel.style.color = "var(--yellow)";
				document.body.classList.add("paused-mode");
				container.classList.add("paused-cue");
			} else {
				autoScroll = true;
				while (messageBuffer.length > 0) appendLine(messageBuffer.shift());
				btn.textContent = "PAUSE SYSTEM";
				btn.className = "";
				statusLabel.textContent = "● LIVE";
				statusLabel.style.color = "var(--green)";
				document.body.classList.remove("paused-mode");
				container.classList.remove("paused-cue");
				window.scrollTo(0, document.body.scrollHeight);
			}
		}

		window.onscroll = function() {
			if (!isPaused) {
				const nearBottom = (window.innerHeight + window.scrollY) >= document.body.offsetHeight - 50;
				autoScroll = nearBottom;
			}
		};
	</script>
</body>
</html>
`

// --- HANDLERS ---

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	data := AppConfig{
		LogPath: GetLogPath(),
	}
	t, _ := template.New("index").Parse(htmlTemplate)
	t.Execute(w, data)
}

func HandleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}
	var newConfig AppConfig
	if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}
	newPath := strings.Trim(newConfig.LogPath, " \"'")
	if !ValidatePath(newPath) {
		http.Error(w, "File does not exist on server", 400)
		return
	}
	SetLogPath(newPath)
	SaveConfig()
	w.WriteHeader(http.StatusOK)
}

func HandleDownload(w http.ResponseWriter, r *http.Request) {
	path := GetLogPath()

	if path == "" {
		http.Error(w, "No log file configured", 400)
		return
	}

	// 1. Parse Query Parameters (start, end)
	queryStart := r.URL.Query().Get("start")
	queryEnd := r.URL.Query().Get("end")

	var startTime, endTime time.Time
	var err error
	inputLayout := "02.01.2006 15:04:05"

	if queryStart != "" {
		startTime, err = time.Parse(inputLayout, queryStart)
		if err != nil {
			http.Error(w, "Invalid Start Time format. Use DD.MM.YYYY HH:MM:SS", 400)
			return
		}
	}
	if queryEnd != "" {
		endTime, err = time.Parse(inputLayout, queryEnd)
		if err != nil {
			http.Error(w, "Invalid End Time format. Use DD.MM.YYYY HH:MM:SS", 400)
			return
		}
	}

	file, err := os.Open(path)
	if err != nil {
		http.Error(w, "Could not open log file", 500)
		return
	}
	defer file.Close()

	// 2. Set Headers for Download (Markdown)
	w.Header().Set("Content-Disposition", "attachment; filename=\"log_export.md\"")
	w.Header().Set("Content-Type", "text/markdown")

	// 3. Start Markdown Code Block
	fmt.Fprintln(w, "```Exported Log Section:")

	// 4. Scan and Filter
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if queryStart == "" && queryEnd == "" {
			fmt.Fprintln(w, line)
			continue
		}

		lineTime, hasTime := extractTimeFromLine(line)

		if hasTime {
			if queryStart != "" && lineTime.Before(startTime) {
				continue
			}
			if queryEnd != "" && lineTime.After(endTime) {
				continue
			}
			fmt.Fprintln(w, line)
		} else {
			if queryStart == "" {
				fmt.Fprintln(w, line)
			}
		}
	}

	// 5. End Markdown Code Block
	fmt.Fprintln(w, "```")
}

func extractTimeFromLine(line string) (time.Time, bool) {
	if len(line) < 21 {
		return time.Time{}, false
	}
	if line[0] != '[' || line[20] != ']' {
		return time.Time{}, false
	}
	timeStr := line[1:20]
	t, err := time.Parse("02.01.2006 15:04:05", timeStr)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

func HandleStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	path := GetLogPath()

	if path == "" {
		fmt.Fprintf(w, "data: [SYSTEM] No log file configured. Go to Settings.\n\n")
		return
	}

	file, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(w, "data: [SYSTEM] Error opening file: %v\n\n", err)
		return
	}
	defer file.Close()

	file.Seek(0, 2)
	reader := bufio.NewReader(file)

	for {
		select {
		case <-r.Context().Done():
			return
		default:
		}
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				time.Sleep(200 * time.Millisecond)
				continue
			}
			break
		}
		fmt.Fprintf(w, "data: %s\n\n", line)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
}
