package main

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// --- CONFIGURATION ---
const (
	defaultLogPath = ``
	port           = ":443"
	certFile       = "server.crt"
	keyFile        = "server.key"
	configFile     = "config.json"
)

// --- STRUCTS ---

type AppConfig struct {
	LogPath string `json:"log_path"`
}

// Global config with mutex for thread safety
var (
	currentConfig AppConfig
	configMutex   sync.RWMutex
)

// --- HTML FRONTEND ---
var htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>Freelancer Log Reader</title>
	<style>
		:root {
			--bg-color: #121212;
			--text-color: #e0e0e0;
			--panel-bg: #1e1e1e;
			--highlight: #00bcd4;
			--highlight-hover: #00acc1;
			--error: #cf6679;
			--success: #03dac6;
		}
		body { 
			background-color: var(--bg-color); 
			color: var(--text-color); 
			font-family: 'Consolas', 'Monaco', monospace; 
			margin: 0; padding: 0; font-size: 14px;
		}
		/* Navbar */
		#controls { 
			position: fixed; top: 0; left: 0; right: 0; height: 50px;
			background: var(--panel-bg); border-bottom: 1px solid #333; 
			display: flex; align-items: center; padding: 0 20px; gap: 20px; 
			z-index: 1000; box-shadow: 0 2px 5px rgba(0,0,0,0.5);
		}
		/* Main Content */
		#log-container { 
			margin-top: 60px; padding: 20px; 
			white-space: pre-wrap; word-break: break-word; line-height: 1.4;
		}
		.log-entry { margin-bottom: 2px; }
		
		/* Buttons */
		button { 
			padding: 8px 15px; cursor: pointer; font-size: 13px; font-weight: bold; 
			border: none; border-radius: 4px; text-transform: uppercase; transition: all 0.2s;
		}
		#btn-action { background-color: #d32f2f; color: #fff; width: 120px; }
		#btn-action.resume { background-color: #388e3c; }
		
		.nav-btn { background-color: #333; color: #fff; }
		.nav-btn:hover { background-color: #555; }
		
		#btn-export { background-color: var(--highlight); color: #000; }
		#btn-export:hover { background-color: var(--highlight-hover); }

		.right-aligned { margin-left: auto; display: flex; gap: 10px; }

		/* Status */
		#status { font-weight: bold; color: #888; }
		.paused-mode { background-color: #000; }
		.paused-cue { border-left: 5px solid #d32f2f; padding-left: 15px; }

		/* Modal */
		.modal-overlay {
			display: none; position: fixed; top: 0; left: 0; width: 100%; height: 100%;
			background: rgba(0,0,0,0.8); z-index: 2000; align-items: center; justify-content: center;
		}
		.modal-box {
			background: var(--panel-bg); padding: 30px; border-radius: 8px; 
			width: 500px; max-width: 90%; box-shadow: 0 4px 15px rgba(0,0,0,0.7);
		}
		h2 { margin-top: 0; color: var(--highlight); }
		label { display: block; margin-top: 15px; color: #aaa; font-size: 12px; }
		input[type="text"] {
			width: 100%; padding: 10px; margin-top: 5px; background: #333; 
			border: 1px solid #555; color: white; font-family: monospace; box-sizing: border-box;
		}
		.modal-buttons { display: flex; justify-content: flex-end; gap: 10px; margin-top: 20px; }
		.hint { font-size: 11px; color: #666; margin-top: 2px; }
	</style>
</head>
<body>
	<div id="controls">
		<button id="btn-action" onclick="togglePause()">PAUSE</button>
		<div id="status">STATUS: LIVE</div>
		<div id="file-display" style="font-size: 12px; color: #666; margin-left: 20px; overflow: hidden; white-space: nowrap; text-overflow: ellipsis; max-width: 300px;">
			{{if .LogPath}}Reading: {{.LogPath}}{{else}}NO FILE SELECTED{{end}}
		</div>
		
		<div class="right-aligned">
			<button id="btn-export" onclick="openExport()">&#128190; Export</button>
			<button class="nav-btn" onclick="openSettings()">&#9881; Settings</button>
		</div>
	</div>

	<div id="log-container">
		{{if not .LogPath}}
			<div style="color: var(--highlight); padding: 20px; border: 1px dashed #555;">
				Welcome! Please click the <b>Settings</b> button in the top right to select your Freelancer log file.
			</div>
		{{end}}
	</div>

	<div id="settings-modal" class="modal-overlay">
		<div class="modal-box">
			<h2>Configuration</h2>
			<label>LOG FILE PATH</label>
			<input type="text" id="path-input" value="{{.LogPath}}" placeholder="C:\Users\Name\...\DSAce.log">
			<div id="save-msg" style="height:20px; font-size:12px; margin-top:5px;"></div>
			<div class="modal-buttons">
				<button onclick="closeModals()" style="background:transparent; color:#888;">Cancel</button>
				<button onclick="saveSettings()" style="background:var(--highlight); color:#000;">Save & Reload</button>
			</div>
		</div>
	</div>

	<div id="export-modal" class="modal-overlay">
		<div class="modal-box">
			<h2>Export Log Section</h2>
			<p style="color:#ccc; font-size:13px;">Leave fields empty to include everything.</p>

			<label>START TIME (Format: 30.12.2025 03:42:02)</label>
			<input type="text" id="export-start" placeholder="30.12.2025 03:42:02">

			<label>END TIME (Format: 30.12.2025 04:00:00)</label>
			<input type="text" id="export-end" placeholder="31.12.2025 23:59:59">
			
			<div class="modal-buttons">
				<button onclick="closeModals()" style="background:transparent; color:#888;">Cancel</button>
				<button onclick="performExport()" style="background:var(--success); color:#000;">Download</button>
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
				statusLabel.textContent = "STATUS: PAUSED (" + messageBuffer.length + " lines)";
			} else {
				appendLine(event.data);
			}
		};

		evtSource.onerror = function() {
			statusLabel.textContent = "DISCONNECTED";
			statusLabel.style.color = "red";
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
				statusLabel.textContent = "STATUS: PAUSED";
				statusLabel.style.color = "#d32f2f";
				document.body.classList.add("paused-mode");
				container.classList.add("paused-cue");
			} else {
				autoScroll = true;
				while (messageBuffer.length > 0) appendLine(messageBuffer.shift());
				btn.textContent = "PAUSE";
				btn.className = "";
				statusLabel.textContent = "STATUS: LIVE";
				statusLabel.style.color = "#888";
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

func main() {
	loadConfig()
	ensureCert()

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/stream", handleStream)
	http.HandleFunc("/config", handleConfig)
	http.HandleFunc("/download", handleDownload)

	fmt.Println("----------------------------------------------------------------")
	fmt.Printf("Freelancer Log Reader Active\n")
	fmt.Printf("Web Interface: https://localhost%s\n", port)
	if currentConfig.LogPath != "" {
		fmt.Printf("Current Log: %s\n", currentConfig.LogPath)
	}
	fmt.Println("----------------------------------------------------------------")

	err := http.ListenAndServeTLS(port, certFile, keyFile, nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// --- HANDLERS ---

func handleIndex(w http.ResponseWriter, r *http.Request) {
	configMutex.RLock()
	data := currentConfig
	configMutex.RUnlock()
	t, _ := template.New("index").Parse(htmlTemplate)
	t.Execute(w, data)
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
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
	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		http.Error(w, "File does not exist on server", 400)
		return
	}
	configMutex.Lock()
	currentConfig.LogPath = newPath
	saveConfig()
	configMutex.Unlock()
	w.WriteHeader(http.StatusOK)
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
	configMutex.RLock()
	path := currentConfig.LogPath
	configMutex.RUnlock()

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

func handleStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	configMutex.RLock()
	path := currentConfig.LogPath
	configMutex.RUnlock()

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

// --- HELPERS ---

func loadConfig() {
	file, err := os.Open(configFile)
	if err != nil {
		currentConfig.LogPath = defaultLogPath
		return
	}
	defer file.Close()
	json.NewDecoder(file).Decode(&currentConfig)
	if currentConfig.LogPath == "" {
		currentConfig.LogPath = defaultLogPath
	}
}

func saveConfig() {
	file, err := os.Create(configFile)
	if err != nil {
		log.Printf("Error saving config: %v", err)
		return
	}
	defer file.Close()
	json.NewEncoder(file).Encode(currentConfig)
}

func ensureCert() {
	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		fmt.Println("Generating self-signed certificate...")
		priv, _ := rsa.GenerateKey(rand.Reader, 2048)
		template := x509.Certificate{
			SerialNumber: big.NewInt(1),
			Subject:      pkix.Name{Organization: []string{"Freelancer Log Tool"}},
			NotBefore:    time.Now(),
			NotAfter:     time.Now().Add(365 * 24 * time.Hour),
			KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
			ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		}
		derBytes, _ := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
		fCert, _ := os.Create(certFile)
		pem.Encode(fCert, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
		fCert.Close()
		fKey, _ := os.Create(keyFile)
		pem.Encode(fKey, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
		fKey.Close()
	}
}
