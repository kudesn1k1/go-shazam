# 🎵 go-shazam

A music recognition service built with Go and Vue.js that can identify songs from audio samples.

## 🎯 What is this?

go-shazam is a self-hosted music recognition platform that uses audio fingerprinting to identify songs.

### Key Features

- 🎤 **Audio Fingerprinting**: Advanced peak detection and hashing algorithm to create unique audio signatures
- 🔍 **Fast Recognition**: Efficient database queries to match audio samples against a library of songs
- 🎵 **Spotify Integration**: Fetch song metadata from Spotify
- 👤 **User Authentication**: JWT-based auth system
- ⚡ **Background Processing**: Redis-based queue system for async song processing
- 🐳 **Docker Ready**: Everything containerized for easy deployment
- 🎨 **Modern UI**: Clean Vue.js frontend that doesn't hurt your eyes

## 🏗️ Architecture

The project follows a microservices-inspired architecture with clear separation of concerns:

```
┌─────────────┐
│   Client    │  Vue.js frontend
│  (Port 80)  │
└──────┬──────┘
       │
┌──────▼──────┐
│   Server    │  Go + Gin HTTP API
│ (Port 8080) │
└──────┬──────┘
       │
   ┌───┴───┬──────────┐
   │       │          │
┌──▼──┐ ┌─▼───┐ ┌────▼────┐
│ DB  │ │Redis│ │ Worker  │
│Postg│ │Queue│ │(Async)  │
└─────┘ └─────┘ └─────────┘
```

### Components

- **Server**: Main HTTP API server handling requests, authentication, and song recognition
- **Worker**: Background worker processing song uploads, generating fingerprints, and downloading audio
- **Database**: PostgreSQL storing songs, fingerprints, and user data
- **Redis**: Message queue for async job processing
- **Client**: Vue.js SPA for the user interface

## 🚀 Quick Start

### Prerequisites

- Docker and Docker Compose (because life's too short for manual setup)
- A Spotify API account

### Setup Instructions

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourusername/go-shazam.git
   cd go-shazam
   ```

2. **Create environment file**
   
   Create `server/.env` with the variables as in .env.example

3. **Start the services**
   ```bash
   docker-compose up -d
   ```

   This will start:
   - PostgreSQL database (with automatic migrations)
   - Redis queue
   - Go server (API)
   - Worker (background processing)
   - Vue.js client (frontend)

4. **Access the application**
   - Frontend: http://localhost
   - API: http://localhost:5000

## 📚 How It Works

### Audio Fingerprinting

1. **Audio Processing**: Audio is processed into fragments and analyzed
2. **Peak Detection**: Spectral peaks are extracted from the audio (the "fingerprint")
3. **Hashing**: Peaks are combined into hash values that uniquely identify audio segments
4. **Matching**: Hashes are compared against the database to find matching songs
5. **Scoring**: Time-aligned matching determines the best match with confidence scores

## 🔧 Configuration

All configuration is done via environment variables in `server/.env`. Key settings:

- **Database**: Connection details for PostgreSQL
- **Redis**: Queue configuration
- **JWT**: Token secrets and expiration times
- **Spotify**: API credentials for metadata fetching
- **Security**: Cookie and encryption settings

## 🙏 Acknowledgments

- Built with [Go](https://go.dev/) (because it's fast and we like fast things)
- Frontend powered by [Vue.js](https://vuejs.org/) (because React is too mainstream)
- Audio processing inspired by various audio fingerprinting algorithms
- Thanks to all the open-source libraries that make this possible

## ⚠️ Disclaimer

This is a pet project built for learning and fun. It's not meant to compete with commercial music recognition services (though we secretly hope it does).

---

*If you find this project useful, consider giving it a ⭐.*

