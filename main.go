package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv" // <--- NEW IMPORT

	"rag-ingestion/internal/chunking"
	"rag-ingestion/internal/datasource"
	"rag-ingestion/internal/embedding"
	"rag-ingestion/internal/enrichment"
	"rag-ingestion/internal/generation" // <--- NEW IMPORT
	"rag-ingestion/internal/parser/implementation"
	"rag-ingestion/internal/retrieval"
	"rag-ingestion/internal/storage"
)

func main() {
	// 1. Load API Keys from .env
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found")
	}

	// 2. Setup Dependencies
	embedder := embedding.NewOllamaEmbedder()
	codeParser := implementation.NewCodeParser("go")
	chunker := chunking.NewStructureAwareChunker(500, 50)
	enricher := enrichment.NewNoOpEnricher()

	// Connect to DB (Use your WSL IP here!)
	vectorStore, err := storage.NewQdrantStore("172.31.97.208", 6334, "rag_knowledge_base")
	if err != nil {
		log.Fatalf("Failed to connect to Qdrant: %v", err)
	}

	// 3. INGESTION (Sniper Mode)
	fmt.Println("--- 🚀 STARTUP: INGESTING DATA ---")
	ingestFiles(vectorStore, embedder, codeParser, chunker, enricher)
	fmt.Println("--- ✅ DATA READY ---")

	// 4. SETUP API SERVER
	r := gin.Default()
	r.StaticFile("/", "./static/index.html")

	myRetriever := retrieval.NewRetriever(embedder, vectorStore)
	myBrain := generation.NewGeminiGenerator() // <--- NEW BRAIN

	// ENDPOINT: /search?q=database
	r.GET("/search", func(c *gin.Context) {
		queryText := c.Query("q")
		if queryText == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
			return
		}

		fmt.Printf("🔍 API received query: %s\n", queryText)

		// A. RETRIEVE (Local)
		results, err := myRetriever.Query(context.Background(), queryText, 3)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// B. GENERATE (Hybrid)
		fmt.Println("🧠 Sending to AI for reasoning...")
		answer, err := myBrain.GenerateAnswer(queryText, results)
		if err != nil {
			fmt.Printf("AI Error: %v\n", err)
			answer = "I found the code, but I couldn't generate an explanation (API Error)."
		}

		// C. RESPOND
		c.JSON(http.StatusOK, gin.H{
			"count":   len(results),
			"query":   queryText,
			"answer":  answer,
			"results": results,
		})
	})

	fmt.Println("\n--- 🤖 API SERVER LISTENING ON :8080 ---")
	r.Run(":8080")
}

// (ingestFiles function stays the same below...)
func ingestFiles(store *storage.QdrantStore, embedder embedding.Embedder, parser *implementation.CodeParser, chunker *chunking.StructureAwareChunker, enricher *enrichment.NoOpEnricher) {
	ctx := context.Background()
	files := []string{"main.go", filepath.Join("internal", "storage", "qdrant_store.go")}
	for _, file := range files {
		source := datasource.NewFileSource(file)
		raw, err := source.Read()
		if err != nil {
			continue
		}
		doc, err := parser.Parse(raw)
		if err != nil {
			continue
		}
		chunks, err := chunker.Split(doc)
		if err != nil {
			continue
		}
		for i := range chunks {
			chunks[i].Metadata["filename"] = file
		}
		vectors, err := embedder.Embed(ctx, chunks)
		if err != nil {
			continue
		}
		store.Upsert(ctx, vectors)
		fmt.Printf("   -> Ingested: %s\n", file)
		time.Sleep(100 * time.Millisecond)
	}
}
