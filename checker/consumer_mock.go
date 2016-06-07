package checker

// MockConsumer mocks interface
type MockConsumer struct {
	CloseError        error
	ReceiveEventError error
	ReceiveEventKey   string
	ReceiveEventValue string
}

// ReceiveEvent mocks method
func (c *MockConsumer) ReceiveEvent() (string, string, error) {
	return c.ReceiveEventKey, c.ReceiveEventValue, c.ReceiveEventError
}

// Close mocks method
func (c *MockConsumer) Close() error {
	return c.CloseError
}
