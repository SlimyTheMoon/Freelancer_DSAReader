# Freelancer DSA Reader v1.1

A cyberpunk-styled log reader for Freelancer with real-time streaming and export capabilities.

## Features

- 🔴 **Real-time Log Streaming** - Watch log files update live via Server-Sent Events
- ⏸ **Pause/Resume** - Buffer incoming logs while paused
- ⬇ **Time-based Export** - Export log sections by timestamp range
- 🎨 **Cyberpunk UI** - Sci-fi themed interface with CRT effects
- 🔒 **Self-signed HTTPS** - Automatic certificate generation

## Quick Start

### Windows

Run `dsa_reader.exe` and navigate to https://localhost:8443 in your browser.

### Linux

```bash
chmod +x dsa_reader_linux
./dsa_reader_linux
```

---

## Compiling from Source

### Prerequisites

Install Go from https://go.dev/ for your operating system.

### Windows

```powershell
# Navigate to the project directory
cd path\to\Freelancer_DSAReader

# Build Windows executable
go build -o dsa_reader.exe .
```

### Linux (from Windows - Cross-compilation)

```powershell
# Navigate to the project directory
cd path\to\Freelancer_DSAReader

# Set environment variables for Linux
$env:GOOS = "linux"
$env:GOARCH = "amd64"

# Build Linux executable
go build -o dsa_reader_linux .

# Or as a single command:
go build -o dsa_reader_linux -ldflags "-s -w" -buildvcs=false
```

### Linux (native compilation)

```bash
# Navigate to the project directory
cd /path/to/Freelancer_DSAReader

# Build Linux executable
GOOS=linux GOARCH=amd64 go build -o dsa_reader_linux .
```

---

## Usage

1. Launch the application
2. Click **CONFIG** to set your Freelancer log file path
3. Click **SAVE & REBOOT** to apply
4. Logs will stream in real-time
5. Use **PAUSE SYSTEM** to pause/resume streaming
6. Use **EXPORT** to download log sections by time range

---

## Notes

- First run generates a self-signed SSL certificate
- Config is saved to `config.json`
- Logs are not modified - read-only access