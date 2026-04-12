# roku-stremio-bridge

A high-concurrency Go orchestrator designed to serve as a Thin Client backend for a Roku Stremio app.

## Architecture
- **Backend**: Go (Gin) handles add-on aggregation and Real-Debrid resolution.
- **Frontend**: Roku SceneGraph (BrightScript/XML) acts as a minimalist renderer.

## Quick Start

### 1. Prerequisites
- Go 1.25+
- Real-Debrid API Key

### 2. Setup
Create a `.env` file from the template:
```env
RD_API_KEY=your_key_here
PORT=8080
```

### 3. Run
```bash
go run main.go
```

## API Endpoints
- `GET /catalog/:type/:id`: Aggregates discovery results from Stremio add-ons.
- `GET /streams/:type/:id?magnet=...&s=1&e=1`: Resolves magnet links into playable RD links.

## Roku Deployment
1. Enable Developer Mode on Roku (`Home x3, Up x2, Right, Left, Right, Left, Right`).
2. Zip the `roku/` folder contents.
3. Upload to the Roku Developer Web Interface.