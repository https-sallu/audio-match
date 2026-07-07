# AudioMatch 🎵 
**A Custom Audio Recognition DSP Engine**

AudioMatch is a full-stack, entirely offline audio recognition engine inspired by Shazam. Instead of relying on third-party APIs, this project implements raw Digital Signal Processing (DSP) mathematics in Go to extract acoustic fingerprints from audio files, store them in SQLite, and match microphone recordings against the database in real-time.

---

## ✨ Core Features
* **Custom DSP Engine:** Implements Fast Fourier Transforms (FFT), spectrogram generation, constellation mapping, and combinatorial hashing from scratch.
* **Smart Two-Stage Recognition:** The React frontend employs a dynamic state machine that performs a "Quick Scan" at 2.5 seconds for instant gratification, and seamlessly falls back to a 10-second "Deep Scan" with adaptive UI prompts (e.g., "Move closer") if the initial audio is too quiet.
* **Dataset Management:** A sleek, animated Admin dashboard to ingest new audio tracks, view the learned database, and safely purge acoustic signatures.
* **Fully Containerized:** The entire infrastructure (Frontend, Backend, and Database) is orchestrated via Docker Compose for zero-configuration deployments.
* **Persistent Storage:** SQLite database and generated audio hashes persist locally via Docker volumes.
* **High-Speed DSP Ingestion:** Utilizes SQLite Write-Ahead Logging (WAL) and bulk transaction wrapping to process and save thousands of combinatorial audio fingerprints in milliseconds.
* **Auto-Provisioning DB:** The backend architecture ensures the database schema is automatically rebuilt if the database file is ever missing or deleted, preventing fatal crash loops.

## 🛠️ Tech Stack
* **Backend:** Go (1.21), `go-chi` (Routing), `go-dsp` (FFT Math), `go-sqlite3` (Database).
* **Frontend:** React, TypeScript, Vite, Tailwind CSS, Lucide React (UI Icons).
* **Infrastructure:** Docker, Docker Compose, Alpine Linux.

---

## 🧠 How the Math Works (The Engine)
When a song is ingested or recorded, the Go backend performs the following steps:
1. **Audio Decoding:** Reads raw 16-bit PCM Mono audio data.
2. **Fast Fourier Transform (FFT):** Slices the audio into overlapping chunks and applies a Hamming Window to convert time-domain audio into frequency-domain data (a Spectrogram).
3. **Constellation Map Extraction:** Scans the spectrogram to find the local maxima (the loudest, most distinct frequency peaks) while ignoring background noise.
4. **Combinatorial Hashing:** Pairs frequencies together based on a target time-zone delay (`FreqA|FreqB|TimeDelta`) and runs them through a SHA-1 hasher to create a unique fingerprint.
5. **Histogram Matching:** When searching for a match, it compares the time offsets of the recorded hashes against the database hashes. The song with the most consistent offset alignment wins.
6. **Smart Pipeline:** To optimize server load and user experience, the system attempts an early exit match at 2.5 seconds. If the confidence score is too low, it automatically extends the recording buffer to capture a denser spectrogram.

---

## ⚠️ Audio Format Requirements
Because the math relies on strict matrix alignments, ingested training files must be strictly formatted. Ensure your dataset files are:
* **Format:** `.wav`
* **Channels:** 1 (Mono)
* **Sample Rate:** 44100 Hz (44.1 kHz)
* **Bit Depth:** 16-bit PCM

*(Tip: You can use FFmpeg or Audacity to batch convert standard MP3s into this exact format).*

---

## 🚀 Getting Started

This project features a zero-configuration, self-healing SQLite database. 

### Prerequisites
* Docker
* Docker Compose

### Installation & Usage
1. Clone the repository:
   ```bash
   git clone https://github.com/https-sallu/audio-match.git
   cd audio-match
Spin up the containers:

Bash
docker-compose up --build -d
Open your browser and navigate to http://localhost:3000.

Train the Engine: Navigate to the Dataset tab, upload a formatted .wav file, and wait for the backend to extract the fingerprints.

Test it out: Navigate to the Recognize tab, play the uploaded song out loud, and click the microphone!


📁 Folder Structure
Plaintext
audio-match/
├── backend/                  # Go DSP Engine & API
│   ├── cmd/server/           # Main application entrypoint
│   ├── db/migrations/        # SQLite schema
│   ├── internal/             
│   │   ├── api/              # HTTP Handlers & Router
│   │   ├── database/         # SQLite Driver Config
│   │   ├── dsp/              # Core Math (FFT, Hashes, Matcher)
│   │   ├── models/           # Go Structs
│   │   └── repository/       # Database queries
│   └── go.mod
├── frontend/                 # React UI
│   ├── src/                  
│   │   ├── hooks/            # useAudioRecorder hook
│   │   ├── pages/            # View components
│   │   └── services/         # API fetch logic
│   ├── package.json
│   └── vite.config.ts
├── docker/                   # Dockerfiles
│   ├── backend.Dockerfile
│   └── frontend.Dockerfile
└── docker-compose.yml        # Orchestration


🛠️ Troubleshooting
Error: database disk image is malformed (Backend 500 Crash)

Cause: The SQLite binary file was corrupted, usually due to a container stopping mid-transaction.

Solution:

Stop the containers: docker-compose down

Delete the corrupted database file: rm backend/data/audiomatch.db

Restart the server: docker-compose up -d

(The backend will automatically generate a fresh, uncorrupted database and rebuild the tables).

👨‍💻 Author
Salman Abbas Full Stack Developer Intern @ CapregSoft

📍 Wah Cantonment, Pakistan