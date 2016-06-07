package health_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"encoding/json"

	"github.com/amalgam8/registry/utils/health"
)

//-----------------------------------------------------------------------------
// status creation
func TestDefaultHealthyResult(t *testing.T) {
	h := health.Healthy

	assert.True(t, h.Healthy)
	assert.Empty(t, h.Properties)
}

func TestStatusHealthy(t *testing.T) {
	h := health.StatusHealthy("all's well")
	assert.True(t, h.Healthy)
	assert.NotEmpty(t, h.Properties)
}

func TestStatusUnhealthy(t *testing.T) {
	uh := health.StatusUnhealthy("failed", nil)
	assert.False(t, uh.Healthy)
}

func TestStatusMessage(t *testing.T) {
	message := "ok"
	h := health.StatusHealthy(message)
	assert.Equal(t, message, h.Properties["message"])
}

func TestStatusCause(t *testing.T) {
	e := errors.New("invalid")
	uh := health.StatusUnhealthy("failed", e)
	assert.Equal(t, e.Error(), uh.Properties["cause"])
}

func TestStatusProperties(t *testing.T) {
	props := map[string]interface{}{"message": "ok", "size": 3, "sync": true}
	uh := health.StatusHealthyWithProperties(props)
	assert.Equal(t, props, uh.Properties)
	assert.EqualValues(t, 3, uh.Properties["size"])
	assert.EqualValues(t, true, uh.Properties["sync"])
}

//-----------------------------------------------------------------------------
// JSON marshaling
func TestMarhsalJSONHealthy(t *testing.T) {
	testcases := []health.Status{
		health.Healthy,
		health.StatusHealthy("service is up"),
		health.StatusHealthyWithProperties(nil),
		health.StatusHealthyWithProperties(map[string]interface{}{"message": "ok", "size": 3.0, "sync": true}),
	}

	for _, s := range testcases {
		b, err := json.Marshal(s)
		assert.Nil(t, err)
		decoded := &health.Status{}
		err = json.Unmarshal(b, decoded)
		assert.Nil(t, err)
		assert.EqualValues(t, s, *decoded)
	}
}

func TestMarhsalJSONUnhealthy(t *testing.T) {
	testcases := []health.Status{
		health.StatusUnhealthy("", nil),
		health.StatusUnhealthy("service is down", nil),
		health.StatusUnhealthy("", errors.New("network unreachable")),
		health.StatusUnhealthy("service is down", errors.New("network unreachable")),
		health.StatusUnhealthyWithProperties(map[string]interface{}{"message": "ok", "size": 3.0, "sync": true}),
	}

	for _, s := range testcases {
		b, err := json.Marshal(s)
		assert.Nil(t, err)
		decoded := &health.Status{}
		err = json.Unmarshal(b, decoded)
		assert.Nil(t, err)
		if err != nil {
			println(err.Error())
		}
		assert.EqualValues(t, s, *decoded)
	}
}

func TestUnmarhsalJSON(t *testing.T) {
	testcases := []struct {
		encoded []byte        // encoded JSON
		status  health.Status // expected object
	}{
		{[]byte(`{"healthy":true}`), health.Healthy},
		{[]byte(`{"healthy":true,"properties":{"message":"service is up"}}`), health.StatusHealthy("service is up")},
		{[]byte(`{}`), health.StatusUnhealthy("", nil)},
		{[]byte(`{"healthy":false,"properties":{"message":"service is down"}}`), health.StatusUnhealthy("service is down", nil)},
		{[]byte(`{"healthy":false,"properties":{"cause":"network unreachable"}}`), health.StatusUnhealthy("", errors.New("network unreachable"))},
		{[]byte(`{"healthy":false,"properties":{"message":"service is down","cause":"network unreachable"}}`),
			health.StatusUnhealthy("service is down", errors.New("network unreachable"))},
		{[]byte(`{"healthy":true,"properties":{"message":"ok","size":3.0, "sync":true}}`),
			health.StatusHealthyWithProperties(map[string]interface{}{"message": "ok", "size": 3.0, "sync": true})},
		{[]byte(`{"healthy":false,"properties":{"message":"ok","size":3.0, "sync":true}}`),
			health.StatusUnhealthyWithProperties(map[string]interface{}{"message": "ok", "size": 3.0, "sync": true})},
	}

	for _, s := range testcases {
		decoded := &health.Status{}
		err := json.Unmarshal(s.encoded, decoded)
		assert.Nil(t, err)
		if err != nil {
			println(err.Error())
		}
		assert.EqualValues(t, s.status, *decoded)
	}
}
