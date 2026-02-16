package entity

type Status string

const (
	Pending    Status = "pending"
	Processing Status = "processing"
	Processed  Status = "processed"
	Failed     Status = "failed"
)
