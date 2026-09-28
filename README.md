Here is a comprehensive, professional `README.md` file tailored exactly to your repository's structure and code. It covers the full-stack architecture, the AI agent fallback logic, and step-by-step setup instructions.

You can copy and paste this directly into a `README.md` file in the **root** of your `Rag` repository.

```markdown
# 🧠 GenUI Agentic RAG System

A blazing-fast, production-grade **Agentic Retrieval-Augmented Generation (RAG)** system. This project features a custom Go backend, a local Qdrant vector database, and a React GenUI frontend. 

What sets this system apart is its **Autonomous Web Agent Fallback**: it searches your local, private documents first, but if it cannot find the answer, it autonomously triggers a Tavily web search to scrape live internet context before generating a final answer.

## ✨ Key Features

* **⚡ Lightning Fast AI Generation:** Powered by **Meta Llama-3.1-8B-Instruct** via Nvidia NIM, reducing generation time to just a few seconds.
* **🌐 Autonomous Web Agent:** Uses a strict evaluation prompt. If local documents lack the necessary context, the system safely catches the AI's fallback and automatically triggers the **Tavily API** to scrape live web data.
* **📂 Local Knowledge Base:** Ingests local PDFs and Code files. Uses **Google Gemini API** for high-quality, dense text embeddings, stored locally in a Dockerized **Qdrant** vector database.
* **🎨 Sleek GenUI Frontend:** A beautiful, dark-mode React interface that parses Markdown, highlights syntax for multiple languages on the fly, and displays citation sources (Local Files vs. Web Search).
* **🏗️ Robust Go Architecture:** Built with Go and the Gin framework, utilizing clean architecture patterns (`cmd`, `internal`, `domain`).

---

## 🛠️ Tech Stack

**Backend**
* **Language:** Go (Golang) 1.21+
* **Framework:** Gin Web Framework
* **Vector Database:** Qdrant (via Docker)
* **Embeddings:** Google Gemini (`gemini-1.5-flash`)
* **Generation:** Meta Llama 3.1 8B Instruct (via Nvidia Cloud API)
* **Web Agent:** Tavily Search API

**Frontend**
* **Framework:** React + TypeScript + Vite
* **Styling:** Tailwind CSS
* **Markdown parsing:** `react-markdown`, `remark-gfm`, `react-syntax-highlighter`

---

## 📂 Project Structure

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

*The frontend will run on `http://localhost:5173`.*

---

## 💡 How it Works (The Fallback Loop)

1. The user asks a question via the React UI.
2. The Go backend searches the **Qdrant DB** for relevant chunks and hands them to **Llama 3.1**.
3. **Strict Evaluation:** Llama is prompted to reply *only* with `"I cannot answer"` if the local context does not contain the answer.
4. **Agent Activation:** If the Go server detects this failure string, it triggers the `TavilyAgent`.
5. Tavily scrapes the live internet, injects the web data as a "fake chunk", and forces Llama 3.1 to try again.
6. The UI beautifully formats the response, indicating whether the source was a local file or the live web.

---

## 🛡️ License & Acknowledgements

Built by [Girish070](https://www.google.com/search?q=https://github.com/Girish070).

```

```