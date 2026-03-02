package namer

// Namer defines the interface for different file naming strategies.
// Each implementation provides a way to generate a new filename.
type Namer interface {
	// Info returns the metadata for the namer.
	Info() Info
	// GenerateName creates a new base filename based on the original and a counter.
	GenerateName(originalBaseName string, counter uint64) (string, error)
}
