package nginx

import (
	"bytes"
	"errors"
	"io"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type configMock struct {
	UpdateFunc func(string) error
	RevertFunc func() error
}

func (m *configMock) Update(config string) error {
	return m.UpdateFunc(config)
}

func (m *configMock) Revert() error {
	return m.RevertFunc()
}

type serviceMock struct {
	StartFunc   func() error
	ReloadFunc  func() error
	RunningFunc func() (bool, error)
}

func (m *serviceMock) Start() error {
	return m.StartFunc()
}

func (m *serviceMock) Reload() error {
	return m.ReloadFunc()
}

func (m *serviceMock) Running() (bool, error) {
	return m.RunningFunc()
}

var _ = Describe("NGINX", func() {

	var (
		c *configMock
		s *serviceMock
		n Nginx
		r io.Reader
	)

	BeforeEach(func() {
		returnNil := func() error { return nil }

		c = &configMock{
			UpdateFunc: func(config string) error { return nil },
			RevertFunc: returnNil,
		}

		s = &serviceMock{
			StartFunc:   returnNil,
			ReloadFunc:  returnNil,
			RunningFunc: func() (bool, error) { return false, nil },
		}

		n = &nginx{
			config:  c,
			service: s,
		}

		r = bytes.NewBufferString("NGINX config contents")
	})

	Context("NGINX is not running", func() {

		It("Updates NGINX configuration and starts NGINX", func() {
			nginxUpdated := false
			nginxStarted := false

			c.UpdateFunc = func(config string) error {
				nginxUpdated = true
				return nil
			}

			s.StartFunc = func() error {
				nginxStarted = true
				return nil
			}

			Expect(n.Update(r)).ToNot(HaveOccurred())
			Expect(nginxUpdated).To(BeTrue())
			Expect(nginxStarted).To(BeTrue())
		})

		Context("NGINX fails to start", func() {

			var (
				revertCalled bool
				startCount   int
			)

			BeforeEach(func() {
				revertCalled = false
				startCount = 0

				s.StartFunc = func() error {
					startCount++
					if !revertCalled {
						return errors.New("Service could not start")
					}
					return nil
				}

				c.RevertFunc = func() error {
					revertCalled = true
					return nil
				}
			})

			It("Reverts to the backup NGINX configuration and starts NGINX", func() {
				Expect(n.Update(r)).To(HaveOccurred())
				Expect(revertCalled).To(BeTrue())
				Expect(startCount).To(Equal(2))
			})

			Context("Revert fails", func() {

			})

		})

	})

	Context("NGINX is running", func() {

		var (
			reloadCount int
			updateCount int
			revertCount int
		)

		BeforeEach(func() {
			reloadCount = 0
			updateCount = 0
			revertCount = 0

			s.RunningFunc = func() (bool, error) {
				return true, nil
			}

			s.ReloadFunc = func() error {
				reloadCount++
				return nil
			}

			c.UpdateFunc = func(config string) error {
				updateCount++
				return nil
			}

			c.RevertFunc = func() error {
				revertCount++
				return nil
			}
		})

		It("Updates NGINX configuration and reloads NGINX", func() {
			Expect(n.Update(r)).ToNot(HaveOccurred())
			Expect(reloadCount).To(Equal(1))
			Expect(updateCount).To(Equal(1))
			Expect(revertCount).To(Equal(0))
		})

		Context("NGINX fails to reload", func() {

			BeforeEach(func() {
				s.ReloadFunc = func() error {
					reloadCount++
					return errors.New("NGINX reload failed")
				}
			})

			It("Reverts to the backup NGINX configuration", func() {
				Expect(n.Update(r)).To(HaveOccurred())
				Expect(reloadCount).To(Equal(1))
				Expect(updateCount).To(Equal(1))
				Expect(revertCount).To(Equal(1))
			})

			Context("Revert fails", func() {

			})

		})

	})

})
