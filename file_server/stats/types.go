package stats

// Stats holds the counters for the application's operations.
// The json tags are for serialization to the frontend.
type Stats struct {
	Scanned uint64 `json:"scanned"`
	Renamed uint64 `json:"renamed"`
	Skipped uint64 `json:"skipped"`
	Errors  uint64 `json:"errors"`
}

func (s *Stats) GetScanned() *uint64 { return &s.Scanned }
func (s *Stats) GetRenamed() *uint64 { return &s.Renamed }
func (s *Stats) GetSkipped() *uint64 { return &s.Skipped }
func (s *Stats) GetErrors() *uint64  { return &s.Errors }

// statsProvider is an internal interface to ensure that any object
// passed to EmitStats has the required methods.
type statsProvider interface {
	GetScanned() *uint64
	GetRenamed() *uint64
	GetSkipped() *uint64
	GetErrors() *uint64
}
