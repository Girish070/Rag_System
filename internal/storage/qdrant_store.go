package storage

import (
	"context"
	"fmt"
	"rag-ingestion/internal/domain/document"

	pb "github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type QdrantStore struct {
	client         pb.PointsClient
	collections    pb.CollectionsClient
	collectionName string
}

func NewQdrantStore(host string, port int, collectionName string) (*QdrantStore, error) {
	adder := fmt.Sprintf("%s:%d", host, port)

	conn, err := grpc.NewClient(adder, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("did not connect: %v", err)
	}
	return &QdrantStore{
		client:         pb.NewPointsClient(conn),
		collections:    pb.NewCollectionsClient(conn),
		collectionName: collectionName,
	}, nil
}

func (qs *QdrantStore) Upsert(ctx context.Context, records []document.VectorRecord) error {
	if len(records) == 0 {
		return nil
	}
	hasCollection, _ := qs.collections.CollectionExists(ctx, &pb.CollectionExistsRequest{
		CollectionName: qs.collectionName,
	})

	if hasCollection == nil || !hasCollection.Result.Exists {
		vectorSize := uint64(len(records[0].Vector))

		_, err := qs.collections.Create(ctx, &pb.CreateCollection{
			CollectionName: qs.collectionName,
			VectorsConfig: &pb.VectorsConfig{
				Config: &pb.VectorsConfig_Params{
					Params: &pb.VectorParams{
						Size:     vectorSize,
						Distance: pb.Distance_Cosine,
					},
				},
			},
		})

		if err != nil {
			return fmt.Errorf("failed to create collection: %w", err)
		}
		fmt.Printf("Created Qdrant Collection: %s (Dim: %d)\n", qs.collectionName, vectorSize)
	}
	var points []*pb.PointStruct

	for _, record := range records {

		payload := make(map[string]*pb.Value)
		for k, v := range record.Metadata {
			payload[k] = &pb.Value{Kind: &pb.Value_StringValue{StringValue: v}}
		}
		payload["text"] = &pb.Value{Kind: &pb.Value_StringValue{StringValue: record.Chunk.Text}}

		points = append(points, &pb.PointStruct{
			Id: &pb.PointId{
				PointIdOptions: &pb.PointId_Uuid{Uuid: record.ID},
			},
			Vectors: &pb.Vectors{
				VectorsOptions: &pb.Vectors_Vector{Vector: &pb.Vector{Data: record.Vector}},
			},
			Payload: payload,
		})
	}
	_, err := qs.client.Upsert(ctx, &pb.UpsertPoints{
		CollectionName: qs.collectionName,
		Points:         points,
	})
	if err != nil {
		return fmt.Errorf("Upsert failed: %w", err)
	}
	return nil
}

func (qs *QdrantStore) Search(ctx context.Context, queryVector []float32, limit int) ([]document.Chunk, error) {
	searchResult, err := qs.client.Search(ctx, &pb.SearchPoints{
		CollectionName: qs.collectionName,
		Vector:         queryVector,
		Limit:          uint64(limit),
		//We want payload (text/metadata) back so we can read it
		WithPayload: &pb.WithPayloadSelector{
			SelectorOptions: &pb.WithPayloadSelector_Enable{
				Enable: true,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("Qdrant search failed: %w", err)
	}

	var results []document.Chunk
	for _, hit := range searchResult.Result {
		textVal := hit.Payload["text"].GetStringValue()
		metadata := make(map[string]string)

		for k, v := range hit.Payload {
			if k == "text" {
				continue
			}
			metadata[k] = v.GetStringValue()
		}
		results = append(results, document.Chunk{
			ID:       hit.Id.GetUuid(),
			Text:     textVal,
			Metadata: metadata,
		})
	}
	return results, nil
}
