package vectordb

type VectorDB interface {
	Insert(id string, vector []float32, metadata map[string]interface{}) error
	Search(vector []float32, topK int) ([]SearchResult, error)
	Delete(id string) error
}

type SearchResult struct {
	ID       string
	Score    float32
	Metadata map[string]interface{}
}
