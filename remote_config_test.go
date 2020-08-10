package gobrake

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NewRemoteSettings", func() {
	var rc *remoteConfig

	BeforeEach(func() {
		rc = newRemoteConfig(&NotifierOptions{
			ProjectId:  1,
			ProjectKey: "key",
		})
	})

	Describe("Interval", func() {
		Context("when JSON PollSec is zero", func() {
			It("returns the default interval", func() {
				interval := 600 * time.Second
				Expect(rc.Interval()).To(Equal(interval))
			})
		})

		Context("when JSON PollSec is above zero", func() {
			BeforeEach(func() {
				rc.JSON.PollSec = 123
			})

			It("returns the interval from JSON", func() {
				interval := 123 * time.Second
				Expect(rc.Interval()).To(Equal(interval))
			})
		})
	})

	Describe("ConfigRoute", func() {
		Context("when JSON ConfigRoute is empty", func() {
			It("returns the default config route", func() {
				url := "https://v1-staging-notifier-configs" +
					".s3.amazonaws.com/2020-06-18/config/" +
					"1/config.json"
				Expect(rc.ConfigRoute()).To(Equal(url))
			})
		})

		Context("when JSON ConfigRoute is non-empty", func() {
			BeforeEach(func() {
				rc.JSON.ConfigRoute = "http://example.com"
			})

			It("returns the config route from JSON", func() {
				Expect(rc.ConfigRoute()).To(
					Equal("http://example.com"),
				)
			})
		})
	})

	Describe("EnabledErrorNotifications", func() {
		Context("when JSON settings has the 'apm' setting", func() {
			BeforeEach(func() {
				rc.JSON.RemoteSettings = append(
					rc.JSON.RemoteSettings,
					&RemoteSettings{
						Name: "errors",
					},
				)
			})

			Context("and when it is enabled", func() {
				BeforeEach(func() {
					rc.JSON.RemoteSettings[0].Enabled = true
				})

				It("returns true", func() {
					Expect(rc.EnabledErrorNotifications()).To(
						BeTrue(),
					)
				})
			})

			Context("and when it is disabled", func() {
				BeforeEach(func() {
					rc.JSON.RemoteSettings[0].Enabled = false
				})

				It("returns false", func() {
					Expect(rc.EnabledErrorNotifications()).To(
						BeFalse(),
					)
				})
			})
		})

		Context("when JSON settings has no 'apm' setting", func() {
			It("returns true", func() {
				Expect(rc.EnabledErrorNotifications()).To(
					BeTrue(),
				)
			})
		})
	})

	Describe("EnabledAPM", func() {
		Context("when JSON settings has the 'apm' setting", func() {
			BeforeEach(func() {
				rc.JSON.RemoteSettings = append(
					rc.JSON.RemoteSettings,
					&RemoteSettings{
						Name: "apm",
					},
				)
			})

			Context("and when it is enabled", func() {
				BeforeEach(func() {
					rc.JSON.RemoteSettings[0].Enabled = true
				})

				It("returns true", func() {
					Expect(rc.EnabledAPM()).To(BeTrue())
				})
			})

			Context("and when it is disabled", func() {
				BeforeEach(func() {
					rc.JSON.RemoteSettings[0].Enabled = false
				})

				It("returns false", func() {
					Expect(rc.EnabledAPM()).To(BeFalse())
				})
			})
		})

		Context("when JSON settings has no 'apm' setting", func() {
			It("returns true", func() {
				Expect(rc.EnabledAPM()).To(BeTrue())
			})
		})
	})

	Describe("ErrorHost", func() {
		Context("when JSON settings has the 'errors' setting", func() {
			BeforeEach(func() {
				rc.JSON.RemoteSettings = append(
					rc.JSON.RemoteSettings,
					&RemoteSettings{
						Name: "errors",
					},
				)
			})

			Context("and when an endpoint is specified", func() {
				BeforeEach(func() {
					setting := rc.JSON.RemoteSettings[0]
					setting.Endpoint = "http://api.example.com"
				})

				It("returns the endpoint", func() {
					Expect(rc.ErrorHost()).To(
						Equal("http://api.example.com"),
					)
				})
			})

			Context("and when an endpoint is NOT specified", func() {
				It("returns an empty string", func() {
					Expect(rc.ErrorHost()).To(Equal(""))
				})
			})
		})

		Context("when JSON settings has no 'errors' setting", func() {
			It("returns an empty string", func() {
				Expect(rc.ErrorHost()).To(Equal(""))
			})
		})
	})

	Describe("ApmHost", func() {
		Context("when JSON settings has the 'apm' setting", func() {
			BeforeEach(func() {
				rc.JSON.RemoteSettings = append(
					rc.JSON.RemoteSettings,
					&RemoteSettings{
						Name: "apm",
					},
				)
			})

			Context("and when an endpoint is specified", func() {
				BeforeEach(func() {
					setting := rc.JSON.RemoteSettings[0]
					setting.Endpoint = "http://api.example.com"
				})

				It("returns the endpoint", func() {
					Expect(rc.ApmHost()).To(
						Equal("http://api.example.com"),
					)
				})
			})

			Context("and when an endpoint is NOT specified", func() {
				It("returns an empty string", func() {
					Expect(rc.ApmHost()).To(Equal(""))
				})
			})
		})

		Context("when JSON settings has no 'apm' setting", func() {
			It("returns an empty string", func() {
				Expect(rc.ApmHost()).To(Equal(""))
			})
		})
	})
})
