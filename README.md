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
├── cmd/
│   ├── ingest/       # Script to parse, chunk, embed, and upload local files to Qdrant
│   └── server/       # Main Go/Gin API server running the Agentic RAG logic
├── internal/
│   ├── agent/        # Tavily web-scraping agent implementation
│   ├── chunking/     # Structure-aware document chunking logic
│   ├── embedding/    # Gemini API integration for vector embeddings
│   ├── generation/   # Nvidia Llama 3.1 generator and strict-prompt logic
│   ├── ingestion/    # Orchestrates parsing -> chunking -> embedding -> storage
│   ├── parser/       # Code (.go) and Document (.pdf) parsing implementations
│   └── storage/      # Qdrant vector store connection and search logic
├── qdrant_storage/   # Persistent local volume for your Qdrant database
└── rag-frontend/     # React Vite frontend application

```

---

## ⚙️ Prerequisites

Before running this project, ensure you have the following installed:

* [Go](https://go.dev/doc/install) (1.21 or higher)
* [Node.js](https://nodejs.org/en) (v18 or higher)
* [Docker Desktop](https://www.docker.com/products/docker-desktop/)

---

## 🚀 Installation & Setup

### 1. Environment Variables

Create a `.env` file in the root directory. **Ensure this file is added to your `.gitignore` to protect your keys.**

```env
GEMINI_API_KEY="your-google-gemini-key"
NVIDIA_API_KEY="Bearer your-nvidia-nim-key"
TAVILY_API_KEY="your-tavily-api-key"

```

### 2. Start the Vector Database

Open Docker Desktop, then run the Qdrant container to store your local embeddings:

```bash
docker run -p 6333:6333 -p 6334:6334 -v $(pwd)/qdrant_storage:/qdrant/storage:z qdrant/qdrant

```

### 3. Ingest Local Documents

To load your PDFs or code into the database, ensure they are placed in the target directory (configured in `cmd/ingest/main.go`) and run the ingestion pipeline:

```bash
go run cmd/ingest/main.go

```

### 4. Start the Go Backend Server

Open a terminal in the root directory and start the Gin API server. This handles both local RAG searches and Web Agent fallbacks.

```bash
go run cmd/server/main.go

```

*The API will run on `http://localhost:8000`.*

### 5. Start the React Frontend

Open a new terminal window, navigate to the frontend folder, install dependencies, and start the Vite dev server:

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
