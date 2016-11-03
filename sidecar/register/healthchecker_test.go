package register

import (
	"errors"
	"testing"
	"time"

	"github.com/amalgam8/amalgam8/sidecar/register/healthcheck"
	"github.com/stretchr/testify/mock"
)

type MockLifecycle struct {
	mock.Mock
}

func (m *MockLifecycle) Start() {
	m.Called()
}

func (m *MockLifecycle) Stop() {
	m.Called()
}

type MockHealthCheckAgent struct {
	mock.Mock

	C chan error
}

func (m *MockHealthCheckAgent) Start(c chan error) {
	m.Called(c)
	m.C = c
}

func (m *MockHealthCheckAgent) Stop() {
	m.Called()
}

func TestHealthChecker(t *testing.T) {
	delay := 10 * time.Millisecond // Delay for changes to take effect in other goroutines.

	// Setup the checker.
	hcAgents := []healthcheck.Agent{
		&MockHealthCheckAgent{},
		&MockHealthCheckAgent{},
		&MockHealthCheckAgent{},
	}

	for i := range hcAgents {
		hcAgents[i].(*MockHealthCheckAgent).On("Start", mock.AnythingOfType("chan error")).Return()
		hcAgents[i].(*MockHealthCheckAgent).On("Stop").Return()
	}

	registration := &MockLifecycle{}
	registration.On("Start").Return()
	registration.On("Stop").Return()

	regStartCount := 0
	regStopCount := 0

	checker := NewHealthChecker(registration, hcAgents)

	// Start checking.
	checker.Start()
	time.Sleep(delay)

	// Assert that start was called on health check agents.
	for _, hcAgent := range hcAgents {
		hcAgent.(*MockHealthCheckAgent).AssertNumberOfCalls(t, "Start", 1)
	}

	registration.AssertNumberOfCalls(t, "Start", regStartCount) // Assert doesn't start registration

	// Make N - 1 healthy.
	for i := 0; i < len(hcAgents)-1; i++ {
		hcAgents[i].(*MockHealthCheckAgent).C <- nil
	}
	time.Sleep(delay)

	registration.AssertNumberOfCalls(t, "Start", regStartCount) // Doesn't start registration

	// Make N healthy.
	for i := 0; i < len(hcAgents); i++ {
		hcAgents[i].(*MockHealthCheckAgent).C <- nil
	}

	time.Sleep(delay)

	regStartCount++
	registration.AssertNumberOfCalls(t, "Start", regStartCount) // Starts registration

	// Make 1 unhealthy.
	hcAgents[0].(*MockHealthCheckAgent).C <- errors.New("mock healthcheck agent error")
	time.Sleep(delay)

	registration.AssertNumberOfCalls(t, "Start", regStartCount) // Doesn't start registration
	regStopCount++
	registration.AssertNumberOfCalls(t, "Stop", regStopCount) // Stops registration

	// Make N unhealthy.
	for i := 0; i < len(hcAgents); i++ {
		hcAgents[i].(*MockHealthCheckAgent).C <- errors.New("mock healthcheck agent error")
	}
	time.Sleep(delay)

	registration.AssertNumberOfCalls(t, "Start", regStartCount) // Doesn't start registration
	registration.AssertNumberOfCalls(t, "Stop", regStopCount)   // Doesn't stop registration

	// Make N healthy.
	for i := 0; i < len(hcAgents); i++ {
		hcAgents[i].(*MockHealthCheckAgent).C <- nil
	}
	time.Sleep(delay)

	regStartCount++
	registration.AssertNumberOfCalls(t, "Start", regStartCount) // Starts registration

	// Stop the checker.
	checker.Stop()
	time.Sleep(delay)

	regStopCount++
	registration.AssertNumberOfCalls(t, "Stop", regStopCount) // Stops registration

	// Assert that stop is called on health check agents.
	for _, hcAgent := range hcAgents {
		hcAgent.(*MockHealthCheckAgent).AssertNumberOfCalls(t, "Stop", 1)
	}
}
