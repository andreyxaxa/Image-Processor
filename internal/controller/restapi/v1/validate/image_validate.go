package validate

const (
	MaxFileSize int64 = 10 * 1024 * 1024

	MinResizeWidth int = 10
	MaxResizeWidth int = 10000

	MinResizeHeight int = 10
	MaxResizeHeight int = 10000

	MinTextLen int = 10
	MaxTextLen int = 64
)

var (
	AllowedContentTypes = map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
	}

	AllowedExtensions = map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
	}
)
