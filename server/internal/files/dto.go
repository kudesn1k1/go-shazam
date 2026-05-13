package files

type UploadResponse struct {
	Hash        string `json:"hash"`
	ContentType string `json:"content_type"`
	SizeBytes   int    `json:"size_bytes"`
}
