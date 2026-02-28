package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"rag-ingestion/internal/chunking"
	"rag-ingestion/internal/datasource"
	"rag-ingestion/internal/embedding"
	"rag-ingestion/internal/enrichment"
	"rag-ingestion/internal/ingestion"
	"rag-ingestion/internal/parser/implementation"
	"rag-ingestion/internal/storage"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	targetDir := flag.String("dir", ".", "Directory to ingest")
	flag.Parse()

	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("Warning: No .env file found")
	}
	embedder := embedding.NewOllamaEmbedder()
	codeParser := implementation.NewCodeParser("go")
	chunker := chunking.NewStructureAwareChunker(800, 100)
	enricher := enrichment.NewNoOpEnricher()

	vectoreStore, err := storage.NewQdrantStore("172.19.171.248", 6334, "Rag_DataBase")
	if err != nil {
		log.Fatalf("Failed to connect Qdrant store: %v\n", err)
	}

	pipeLine := ingestion.NewPipeline(codeParser, chunker, enricher, embedder, vectoreStore)

	fmt.Printf("Starting Ingestion on: %s", *targetDir)

	count := 0
	err = filepath.Walk(*targetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".go" || ext == ".md" {
			fmt.Println("Processing: %s...", path)

			source := datasource.NewFileSource(path)

			err := pipeLine.Run(context.Background(), source)
			if err != nil {
				fmt.Printf("Failed: %v\n", err)
			} else {
				fmt.Printf("Done \n")
				count++
			}
			time.Sleep(50 * time.Millisecond)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error Walking Path %v", err)
	}
	fmt.Printf("Ingestion Complete, Processed %d files \n", count)
}
