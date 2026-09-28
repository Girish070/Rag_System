# Rag

A local-first retrieval-augmented generation project built with Go and a React frontend. It ingests files, embeds them with Ollama, stores vectors in Qdrant, and answers questions through a search API with web fallback support.

## Overview

This project is designed to:

- parse supported files such as Go source and PDF documents
- split large content into chunks
- generate embeddings locally with Ollama
- store those vectors in Qdrant for similarity search
- retrieve the most relevant chunks for a user question
- use a model to answer from those chunks
- trigger a Tavily web search when local context is missing

## Tech stack

- Go
- Gin
- Qdrant
- Ollama
- NVIDIA API
- Tavily API
- React + TypeScript + Vite
- Tailwind CSS

## Project structure

```text
.
├── cmd/
│   ├── ingest/
│   │   └── main.go
│   └── server/
│       └── main.go
├── internal/
│   ├── agent/
│   ├── chunking/
│   ├── datasource/
│   ├── domain/
│   ├── embedding/
│   ├── enrichment/
│   ├── generation/
│   ├── ingestion/
│   ├── parser/
│   ├── retrieval/
│   └── storage/
├── rag-frontend/
├── go.mod
├── README.md
└── scripts/
```

## Prerequisites

Before running the project, make sure you have:

- Go installed
- Node.js installed
- Docker installed
- Ollama installed and running
- Access to NVIDIA and Tavily API keys

## Environment setup

Create a `.env` file in the project root:

```env
NVIDIA_API_KEY=your_nvidia_key
TAVILY_API_KEY=your_tavily_key
```

## Run the app

### 1. Start Ollama

```bash
ollama serve
ollama pull nomic-embed-text
```

### 2. Start Qdrant

```bash
docker run -p 6333:6333 -p 6334:6334 -v qdrant_data:/qdrant/storage qdrant/qdrant
```

### 3. Ingest content

```bash
go run ./cmd/ingest --dir .
```

### 4. Start the Go backend

```bash
go run ./cmd/server
```

The backend listens on:

```text
http://localhost:8000
```

### 5. Start the frontend

```bash
cd rag-frontend
npm install
npm run dev
```

The frontend runs at:

```text
http://localhost:5173
```

## API usage

The main search endpoint is:

```http
GET /search?q=your+question
```

## Notes

- The project currently hardcodes the Qdrant host in the Go startup files, so you may need to update it for your local environment.
- Ingestion must be run before querying the vector database.
- The fallback flow uses Tavily when the local context is insufficient.

## License

No explicit license is currently defined in the repository.
