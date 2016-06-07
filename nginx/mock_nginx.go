package nginx

import (
	"io"
)

// MockGenerator implements interface
type MockGenerator struct {
	GenerateString string
	GenerateError  error
}

// Generate mocks method
func (m *MockGenerator) Generate(w io.Writer, id string) error {
	w.Write([]byte(m.GenerateString))
	return m.GenerateError
}
