package clients

import "github.com/amalgam8/controller/resources"

type MockNginx struct {
	UpdateHttpError error
}

func (m *MockNginx) UpdateHttpUpstreams(conf resources.NGINXJson) error {
	return m.UpdateHttpError
}
