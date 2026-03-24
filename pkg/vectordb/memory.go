package vectordb

import (
	"math"
	"sort"
	"sync"
)

type MemoryVectorDB struct {
	mu      sync.RWMutex
	vectors map[string][]float32
	metadata map[string]map[string]interface{}
}

func NewMemoryVectorDB() *MemoryVectorDB {
	return &MemoryVectorDB{
		vectors:  make(map[string][]float32),
		metadata: make(map[string]map[string]interface{}),
	}
}

func (db *MemoryVectorDB) Insert(id string, vector []float32, metadata map[string]interface{}) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.vectors[id] = vector
	db.metadata[id] = metadata
	return nil
}

func (db *MemoryVectorDB) Search(vector []float32, topK int) ([]SearchResult, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	type scoredResult struct {
		id    string
		score float32
	}

	results := make([]scoredResult, 0, len(db.vectors))
	for id, v := range db.vectors {
		score := cosineSimilarity(vector, v)
		results = append(results, scoredResult{id: id, score: score})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	if topK > len(results) {
		topK = len(results)
	}

	searchResults := make([]SearchResult, topK)
	for i := 0; i < topK; i++ {
		searchResults[i] = SearchResult{
			ID:       results[i].id,
			Score:    results[i].score,
			Metadata: db.metadata[results[i].id],
		}
	}

	return searchResults, nil
}

func (db *MemoryVectorDB) Delete(id string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	delete(db.vectors, id)
	delete(db.metadata, id)
	return nil
}

func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return float32(dotProduct / (math.Sqrt(normA) * math.Sqrt(normB)))
}
