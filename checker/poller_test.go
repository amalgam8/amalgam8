package checker

import (
	"errors"
	"io"
	"time"

	"github.com/amalgam8/sidecar/clients"
	"github.com/amalgam8/sidecar/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type mockNginx struct {
	UpdateFunc func(io.Reader) error
}

func (m *mockNginx) Update(reader io.Reader) error {
	return m.UpdateFunc(reader)
}

var _ = Describe("Tenant Poller", func() {

	var (
		rc *clients.MockController
		n  *mockNginx
		c  *config.Config
		p  *poller

		updateCount int
	)

	BeforeEach(func() {
		updateCount = 0

		rc = &clients.MockController{
			ConfigString: "non-empty-config",
		}
		n = &mockNginx{
			UpdateFunc: func(reader io.Reader) error {
				updateCount++
				return nil
			},
		}
		c = &config.Config{
			Tenant: config.Tenant{
				ID:        "id",
				Token:     "token",
				TTL:       60 * time.Second,
				Heartbeat: 30 * time.Second,
				Port:      8080,
			},
			Registry: config.Registry{
				URL:   "http://regsitry",
				Token: "sd_token",
			},
			Kafka: config.Kafka{
				Brokers: []string{
					"http://broker1",
					"http://broker2",
					"http://broker3",
				},
				Username: "username",
				Password: "password",
			},
			Nginx: config.Nginx{
				Port:    6379,
				Logging: false,
			},
			Controller: config.Controller{
				URL:  "http://controller",
				Poll: 60 * time.Second,
			},
		}

		p = &poller{
			controller: rc,
			nginx:      n,
			config:     c,
		}
	})

	It("polls successfully", func() {
		Expect(p.poll()).ToNot(HaveOccurred())
		Expect(updateCount).To(Equal(1))
	})

	It("reports NGINX update failure", func() {
		n.UpdateFunc = func(reader io.Reader) error {
			return errors.New("Update NGINX failed")
		}

		Expect(p.poll()).To(HaveOccurred())
	})

	It("does not update NGINX if unable to obtain config from Controller", func() {
		rc.ConfigError = errors.New("Get rules failed")

		Expect(p.poll()).To(HaveOccurred())
		Expect(updateCount).To(Equal(0))
	})

})
