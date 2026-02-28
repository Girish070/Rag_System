package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"rag-ingestion/internal/embedding"
	"rag-ingestion/internal/generation"
	"rag-ingestion/internal/retrieval"
	"rag-ingestion/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("Warning: No .env file found")
	}

	embedder := embedding.NewOllamaEmbedder()

	vectorStoer, err := storage.NewQdrantStore("172.19.171.248", 6334, "Rag_DataBase")
	if err != nil {
		log.Fatalf("Failed to connect to Qdrant: %v", err)
	}

	myRetriever := retrieval.NewRetriever(embedder, vectorStoer)
	myBrain := generation.NewGeminiGenerator()

	r := gin.Default()

	r.StaticFile("/", "./Static/index.html")

	r.GET("/search", func(c *gin.Context) {
		queryText := c.Query("q")
		if queryText == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "query paramertr is missing"})
			return
		}
		fmt.Printf("User asked: %s\n", queryText)

		imporvedQuery, err := myBrain.ImproveQuery(queryText)
		if err == nil && imporvedQuery != "" {
			fmt.Printf("Ai rewirtten: %s\n", imporvedQuery)
			queryText = imporvedQuery
		} else {
			fmt.Printf("Reasoning failed: %v\n", err)
		}

		result, err := myRetriever.Query(context.Background(), queryText, 3)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		answer, err := myBrain.GenerateAnswer(queryText, result)
		if err != nil {
			fmt.Printf("Brain Error: %v\n", err)
			answer = "I found the code but the brain timed out"
		}

		c.JSON(http.StatusOK, gin.H{
			"results": result,
			"answer":  answer,
		})
	})

	fmt.Println("Server listening on 8000")
	r.Run(":8000")
}
