package gobrake

import (
	"fmt"
	"time"
)

// How frequently we should poll the config API.
const defaultInterval = 600 * time.Second

// API version of the S3 API to poll.
const apiVer = "2020-06-18"

// What URL to poll.
const configRoutePattern = "https://v1-%s.s3.amazonaws.com/%s/config/%d/config.json"

// Default AWS S3 bucket with notifier configs.
const defaultBucket = "staging-notifier-configs"

const (
	apmSetting   = "apm"
	errorSetting = "errors"
)

type remoteConfig struct {
	JSON *RemoteConfigJSON
}

type RemoteConfigJSON struct {
	ProjectId   int64  `json:"project_id"`
	UpdatedAt   int64  `json:"updated_at"`
	PollSec     int64  `json:"pol_sec"`
	ConfigRoute string `json:"config_route"`

	RemoteSettings []*RemoteSettings `json:"settings"`
}

type RemoteSettings struct {
	Name     string `json:"name"`
	Enabled  bool   `json:"enabled"`
	Endpoint string `json:"endpoint"`
}

func newRemoteConfig() *remoteConfig {
	return &remoteConfig{
		JSON: &RemoteConfigJSON{},
	}
}

func (rc *remoteConfig) Interval() time.Duration {
	if rc.JSON.PollSec > 0 {
		return time.Duration(rc.JSON.PollSec) * time.Second
	}

	return defaultInterval
}

func (rc *remoteConfig) ConfigRoute() string {
	if rc.JSON.ConfigRoute != "" {
		return rc.JSON.ConfigRoute
	}

	return fmt.Sprintf(configRoutePattern,
		defaultBucket, apiVer, rc.JSON.ProjectId)
}

func (rc *remoteConfig) EnabledErrorNotifications() bool {
	for _, s := range rc.JSON.RemoteSettings {
		if s.Name == errorSetting {
			return s.Enabled
		}
	}

	return true
}

func (rc *remoteConfig) EnabledAPM() bool {
	for _, s := range rc.JSON.RemoteSettings {
		if s.Name == apmSetting {
			return s.Enabled
		}
	}

	return true
}

func (rc *remoteConfig) ErrorHost() string {
	for _, s := range rc.JSON.RemoteSettings {
		if s.Name == errorSetting {
			return s.Endpoint
		}
	}

	return ""
}

func (rc *remoteConfig) ApmHost() string {
	for _, s := range rc.JSON.RemoteSettings {
		if s.Name == apmSetting {
			return s.Endpoint
		}
	}

	return ""
}
