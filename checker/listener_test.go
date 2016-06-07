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

var _ = Describe("Tenant listener", func() {

	var (
		consumer *MockConsumer
		rc       *clients.MockController
		n        *mockNginx
		c        *config.Config
		l        *listener

		updateCount int
	)

	BeforeEach(func() {
		updateCount = 0

		consumer = &MockConsumer{
			ReceiveEventKey: "id",
		}
		rc = &clients.MockController{}
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
				URL:   "http://registry",
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

		l = &listener{
			consumer:   consumer,
			controller: rc,
			nginx:      n,
			config:     c,
		}

	})

	It("listens for an update event successfully", func() {
		Expect(l.listenForUpdate()).ToNot(HaveOccurred())
		Expect(updateCount).To(Equal(1))
	})

	It("reports NGINX update failure", func() {
		n.UpdateFunc = func(reader io.Reader) error {
			return errors.New("Update NGINX failed")
		}

		Expect(l.listenForUpdate()).To(HaveOccurred())
	})

	It("does not update NGINX if unable to obtain config from Controller", func() {
		rc.ConfigError = errors.New("Get rules failed")

		Expect(l.listenForUpdate()).To(HaveOccurred())
		Expect(updateCount).To(Equal(0))
	})

})
