package payload

type SearchRequest struct {
	RootFolderID string `json:"root_folder_id"`
	Query        string `json:"query"`
	Limit        int    `json:"limit"`
}

type SearchResult struct {
	ChunkID       string  `json:"chunk_id"`
	Score         float64 `json:"score"`
	RootFolderID  string  `json:"root_folder_id"`
	FolderID      string  `json:"folder_id"`
	DocumentID    string  `json:"document_id"`
	DocumentTitle string  `json:"document_title"`
	FolderPath    string  `json:"folder_path"`
	HeadingPath   string  `json:"heading_path"`
	Content       string  `json:"content"`
}

type IndexFolderResponse struct {
	BatchID       string `json:"batch_id"`
	RootFolderID  string `json:"root_folder_id"`
	DocumentCount int    `json:"document_count"`
	StartedAt     string `json:"started_at"`
}

type IndexJobProgress struct {
	ID           string  `json:"id"`
	DocumentID   string  `json:"document_id"`
	RootFolderID *string `json:"root_folder_id"`
	Status       string  `json:"status"`
	ErrorMessage *string `json:"error_message"`
	ChunkCount   int     `json:"chunk_count"`
	CreatedAt    string  `json:"created_at"`
	StartedAt    *string `json:"started_at"`
	FinishedAt   *string `json:"finished_at"`
}
