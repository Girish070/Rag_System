# Rag

A local-first retrieval-augmented generation app built in Go with a React frontend. It ingests Go source files and PDF documents, embeds them with Ollama, stores them in Qdrant, and answers questions through a search API with a web fallback when local context is insufficient.

## What this project does

This repository combines:

- a Go backend for indexing and search
- a Qdrant vector database for similarity search
- Ollama embeddings for semantic matching
- NVIDIA-hosted Llama generation for answer synthesis
- Tavily web search as a fallback when no local answer is available
- a Vite + React frontend for querying the system

The flow is:

1. Parse supported files
2. Split them into chunks
3. Embed each chunk with Ollama
4. Store vectors in Qdrant
5. Search the closest chunks for a question
6. Ask the model to answer using those chunks
7. If the model says it cannot answer, fetch web context and retry

## Tech stack

Backend:
- Go
- Gin
- Qdrant gRPC client
- Ollama
- NVIDIA API for chat completion
- Tavily API for web fallback

Frontend:
- React
- TypeScript
- Vite
- Tailwind CSS
- Markdown rendering in the browser

## Repository structure

```text
.
├── cmd/
│   ├── ingest/
│   │   └── main.go        # indexes files from a directory into Qdrant
│   └── server/
│       └── main.go        # Gin API server for search and answer generation
├── internal/
│   ├── agent/
│   │   └── tavily.go      # Tavily web search wrapper
│   ├── chunking/
│   │   └── ...            # chunking logic
│   ├── datasource/
│   │   └── ...            # file input abstraction
│   ├── domain/
│   │   └── ...            # domain models
│   ├── embedding/
│   │   └── ollama_embedder.go
│   ├── enrichment/
│   │   └── ...
│   ├── generation/
│   │   └── nvidia.go      # answer generation and query rewriting
│   ├── ingestion/
│   │   └── pipeline.go    # parse -> chunk -> embed -> store pipeline
│   ├── parser/
│   │   └── ...            # code/pdf parsing implementations
│   ├── retrieval/
│   │   └── retriever.go   # vector retrieval logic
│   └── storage/
│       └── qdrant_store.go # Qdrant access layer
├── rag-frontend/
│   └── ...                # React app
├── go.mod
├── .env.example           # not present in repo; create your own .env
├── README.md
└── scripts/
```

## Prerequisites

Install these first:

- Go 1.25 or newer
- Node.js 18+
- Docker
- Ollama

## Required services

This app expects the following to be running:

- Qdrant at port 6334 (the code currently connects to 172.20.128.192:6334)
- Ollama at http://localhost:11434
- NVIDIA API access via `NVIDIA_API_KEY`
- Tavily API access via `TAVILY_API_KEY`

> The code currently hardcodes the Qdrant host in the Go startup files. If you are not on the same network setup as this project, update the host in both `cmd/ingest/main.go` and `cmd/server/main.go` before running the app.

## Environment variables

Create a `.env` file in the project root with values similar to:

```env
NVIDIA_API_KEY=your_nvidia_api_key
TAVILY_API_KEY=your_tavily_api_key
```

The app also expects Ollama to be installed and available locally.

## Setup

### 1. Start Ollama

If you do not already have Ollama installed:

```bash
ollama serve
ollama pull nomic-embed-text
```

### 2. Start Qdrant

```bash
docker run -p 6333:6333 -p 6334:6334 -v qdrant_data:/qdrant/storage qdrant/qdrant
```

### 3. Ingest documents

The ingestion tool walks a directory and processes supported file types. In the current implementation it handles:

- `.go`
- `.pdf`

Example:

```bash
go run ./cmd/ingest --dir ./your-data-folder
```

If you want to ingest the repository itself:

```bash
go run ./cmd/ingest --dir .
```

### 4. Start the backend

```bash
go run ./cmd/server
```

The server listens on:

```text
http://localhost:8000/search?q=your+question
```

### 5. Start the frontend

```bash
cd rag-frontend
npm install
npm run dev
```

The frontend typically runs at:

```text
http://localhost:5173
```

## API usage

The backend exposes a GET endpoint:

```http
GET /search?q=what+does+this+project+do
```

Example response:

```json
{
  "results": [
    {
      "text": "Relevant chunk from indexed document",
      "metadata": {
        "filename": "example.go"
      }
    }
  ],
  "answer": "The project indexes local documents and answers questions from them."
}
```

## Fallback behavior

The app tries to answer from the local vector search results first. If the model responds with an inability to answer or returns an empty result, the server triggers the Tavily API and reruns generation with scraped web content.

This fallback is handled in:

- `cmd/server/main.go`
- `internal/agent/tavily.go`
- `internal/generation/nvidia.go`

## Notes

- The vector search is local and document-based, so ingestion must be run before searching.
- The app is optimized for code and PDF content rather than broad general-purpose web search.
- Qdrant addresses are currently configured in code and may require adjustment for your environment.

## License

This project does not currently declare a license in the repository metadata. If you plan to distribute or reuse it, add an appropriate license file before publishing.

## Next steps

A few common improvements would be:

- add a `.env.example` file
- move Qdrant host/port to environment variables
- add configurable file extensions and directories
- add a proper health endpoint
- add JSON schema validation and stronger error handling
- make the frontend show source metadata and fallback provenance more clearly
