package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"rag-ingestion/internal/agent"
	"rag-ingestion/internal/domain/document"
	"rag-ingestion/internal/embedding"
	"rag-ingestion/internal/generation"
	"rag-ingestion/internal/retrieval"
	"rag-ingestion/internal/storage"
	"strings"
	"time"

	"github.com/gin-contrib/cors" // <--- NEW IMPORT
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: No .env file found")
	}

	embedder := embedding.NewOllamaEmbedder()
	// Make sure this IP matches your WSL IP!
	vectorStore, err := storage.NewQdrantStore("172.20.128.192", 6334, "Rag_DataBase")
	if err != nil {
		log.Fatalf("Failed to connect to Qdrant: %v", err)
	}

	myRetriever := retrieval.NewRetriever(embedder, vectorStore)
	myBrain := generation.NewNvidiaGenerator()
	webAgent := agent.NewTavilyAgent()

	r := gin.Default()

	// 🔓 ENABLE CORS (Allow React to talk to Go)
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // Vite's default port
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// (Keep your routes the same)
	r.GET("/search", func(c *gin.Context) {
		queryText := c.Query("q")
		if queryText == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter is missing"})
			return
		}
		fmt.Printf("👤 User asked: '%s'\n", queryText)

		improvedQuery, err := myBrain.ImproveQuery(queryText)
		if err == nil && improvedQuery != "" {
			fmt.Printf("🧠 AI rewritten: '%s'\n", improvedQuery)
			queryText = improvedQuery
		}

		results, err := myRetriever.Query(context.Background(), queryText, 3)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		answer, err := myBrain.GenerateAnswer(queryText, results)

		if err != nil || strings.Contains(answer, "I cannot answer") || answer == "" {
			fmt.Println("⚠️ Local DB lacked context. Triggering Web Agent for:", queryText)

			webContent, webErr := webAgent.SearchWeb(queryText)
			if webErr == nil && webContent != "" {

				fakeWebChunk := document.Chunk{
					Metadata: map[string]string{"filename": "Tavily Web Search"},
					Text:     webContent,
				}
				answer, err = myBrain.GenerateAnswer(queryText, []document.Chunk{fakeWebChunk})
				if err != nil {
					answer = "Brain freeze! The local DB failed, and the Web Agent crashed."
				} else {
					answer = "> 🌐 **Web Agent Activated:** I couldn't find this in your local files, so I searched the internet.\n\n" + answer
					results = []document.Chunk{fakeWebChunk}
				}
			} else {
				fmt.Println("❌ TAVILY ERROR:", webErr)
				answer = "Brain freeze! I couldn't generate an answer locally, and web search failed."
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"results": results,
			"answer":  answer,
		})
	})

	fmt.Println("Server listening on 8000")
	r.Run(":8000")
}
