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
	"rag-ingestion/internal/parser"
	"rag-ingestion/internal/parser/implementation"
	"rag-ingestion/internal/storage"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	targetDir := flag.String("dir", ".", "Directory to ingest")
	flag.Parse()

	// 1. Load Environment
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("Warning: No .env file found")
	}

	// 2. Setup Components
	embedder := embedding.NewOllamaEmbedder()
	chunker := chunking.NewStructureAwareChunker(800, 100)
	enricher := enrichment.NewNoOpEnricher()

	// ✅ FIXED: Removed the trailing space inside the quotes!
	vectorStore, err := storage.NewQdrantStore("172.20.128.192", 6334, "Rag_DataBase")
	if err != nil {
		log.Fatalf("Failed to connect Qdrant store: %v\n", err)
	}

	// 3. Initialize Both Parsers (Student Mode Activated 🎓)
	goParser := implementation.NewCodeParser("go")
	pdfParser := implementation.NewPdfParser()

	fmt.Printf("--- 🦖 STARTING INGESTION ON: %s ---\n", *targetDir)

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

		// 4. Smart Parser Selection
		ext := strings.ToLower(filepath.Ext(path))
		var activeParser parser.Parser

		switch ext {
		case ".go":
			activeParser = goParser
		case ".pdf": // 👈 Now supports PDFs!
			activeParser = pdfParser
			fmt.Printf("📄 Found PDF: %s\n", info.Name())
		default:
			return nil
		}

		fmt.Printf("Processing: %s...\n", info.Name())

		// 5. Create Pipeline for this file
		pipeLine := ingestion.NewPipeline(activeParser, chunker, enricher, embedder, vectorStore)
		source := datasource.NewFileSource(path)

		err = pipeLine.Run(context.Background(), source)
		if err != nil {
			fmt.Printf("❌ Failed: %v\n", err)
		} else {
			fmt.Printf("✅ Done\n")
			count++
		}

		// Tiny sleep to be nice to Ollama
		time.Sleep(50 * time.Millisecond)
		return nil
	})

	if err != nil {
		log.Fatalf("Error Walking Path %v", err)
	}
	fmt.Printf("\n--- 🎉 INGESTION COMPLETE. Processed %d files ---\n", count)
}
