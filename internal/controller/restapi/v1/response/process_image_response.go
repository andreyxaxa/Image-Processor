package response

type ProcessImage struct {
	ImageID      string `json:"image_id"`
	OriginalName string `json:"original_name"`
	Size         int    `json:"size"`
	ContentType  string `json:"content_type"`
	Status       string `json:"status"`
	Operation    string `json:"operation"`
	CreatedAt    string `json:"created_at"`
}
