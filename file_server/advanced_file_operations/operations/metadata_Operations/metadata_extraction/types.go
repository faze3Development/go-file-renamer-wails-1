package metadata_extraction

// Metadata represents the extracted metadata from a file
type Metadata struct {
	// Basic file information
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	ContentType string `json:"contentType"`
	HasMetadata bool   `json:"hasMetadata"`

	// Image-specific metadata
	Image *ImageMetadata `json:"image,omitempty"`

	// PDF-specific metadata
	PDF *PDFMetadata `json:"pdf,omitempty"`

	// Raw metadata for extensibility
	Raw map[string]any `json:"raw,omitempty"`
}

// ImageMetadata contains EXIF and other image metadata
type ImageMetadata struct {
	Width        int               `json:"width,omitempty"`
	Height       int               `json:"height,omitempty"`
	ColorSpace   string            `json:"colorSpace,omitempty"`
	HasExif      bool              `json:"hasExif"`
	ExifData     map[string]any    `json:"exifData,omitempty"`
	DateTime     string            `json:"dateTime,omitempty"`
	CameraMake   string            `json:"cameraMake,omitempty"`
	CameraModel  string            `json:"cameraModel,omitempty"`
	LensModel    string            `json:"lensModel,omitempty"`
	ISO          string            `json:"iso,omitempty"`
	Aperture     string            `json:"aperture,omitempty"`
	ShutterSpeed string            `json:"shutterSpeed,omitempty"`
	FocalLength  string            `json:"focalLength,omitempty"`
	Orientation  string            `json:"orientation,omitempty"`
	Software     string            `json:"software,omitempty"`
	Location     *LocationMetadata `json:"location,omitempty"`
}

// PDFMetadata contains PDF-specific metadata
type PDFMetadata struct {
	Title        string `json:"title,omitempty"`
	Author       string `json:"author,omitempty"`
	Subject      string `json:"subject,omitempty"`
	Creator      string `json:"creator,omitempty"`
	Producer     string `json:"producer,omitempty"`
	Keywords     string `json:"keywords,omitempty"`
	CreationDate string `json:"creationDate,omitempty"`
	ModDate      string `json:"modDate,omitempty"`
	PageCount    int    `json:"pageCount,omitempty"`
	PDFVersion   string `json:"pdfVersion,omitempty"`
}

// LocationMetadata contains GPS location data
type LocationMetadata struct {
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	Altitude  float64 `json:"altitude,omitempty"`
}
