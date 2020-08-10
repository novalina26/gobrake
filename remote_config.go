package gobrake

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// How frequently we should poll the config API.
const defaultInterval = 5 * time.Second // 600 * time.Second

// API version of the S3 API to poll.
const apiVer = "2020-06-18"

// What URL to poll.
const configRoutePattern = "https://v1-%s.s3.amazonaws.com/%s/config/%d/config.json"

// Default AWS S3 bucket with notifier configs.
const defaultBucket = "staging-notifier-configs"

// Setting names in JSON returned by the API.
const (
	apmSetting   = "apm"
	errorSetting = "errors"
)

type remoteConfig struct {
	opt    *NotifierOptions
	poller *poller

	JSON *RemoteConfigJSON
}

type RemoteConfigJSON struct {
	ProjectId   int64  `json:"project_id"`
	UpdatedAt   int64  `json:"updated_at"`
	PollSec     int64  `json:"poll_sec"`
	ConfigRoute string `json:"config_route"`

	RemoteSettings []*RemoteSettings `json:"settings"`
}

type RemoteSettings struct {
	Name     string `json:"name"`
	Enabled  bool   `json:"enabled"`
	Endpoint string `json:"endpoint"`
}

func newRemoteConfig(opt *NotifierOptions) *remoteConfig {
	return &remoteConfig{
		opt:  opt,
		JSON: &RemoteConfigJSON{},
	}
}

type poller struct {
	ticker *time.Ticker
	closer chan bool
}

func newPoller(interval time.Duration) *poller {
	return &poller{
		ticker: time.NewTicker(interval),
		closer: make(chan bool),
	}
}

func (p *poller) Stop() {
	p.ticker.Stop()
	close(p.closer)
}

type configCallback func(*remoteConfig)

func (rc *remoteConfig) Poll(cb configCallback) {
	rc.poller = newPoller(rc.Interval())
	go rc.poll(cb)
}

func (rc *remoteConfig) poll(cb configCallback) {
	for {
		select {
		case <-rc.poller.closer:
			return
		case <-rc.poller.ticker.C:
			cfg, err := rc.fetchConfig()
			if err != nil {
				logger.Printf(fmt.Sprintf("fetchConfig failed: %s", err))
				continue
			}

			if cfg == nil {
				return
			}
			rc.poller.ticker.Stop()
			rc.JSON = cfg
			rc.poller.ticker = time.NewTicker(rc.Interval())

			cb(rc)
		}
	}
}

func (rc *remoteConfig) StopPolling() {
	rc.poller.Stop()
}

func (rc *remoteConfig) fetchConfig() (*RemoteConfigJSON, error) {
	resp, err := http.Get(rc.ConfigRoute())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Printf(fmt.Sprintf("fetchConfig failed: %s", err))
	}

	// AWS S3 API returns XML when request is not valid. In this case we
	// just print the returned body and exit.
	if strings.HasPrefix(string(body), "<?xml ") {
		return nil, errors.New(string(body))
	}

	var j *RemoteConfigJSON
	err = json.Unmarshal(body, &j)
	if err != nil {
		return nil, err
	}

	return j, nil
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
		defaultBucket, apiVer, rc.opt.ProjectId)
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
		if s.Name == errorSetting && s.Endpoint != "" {
			return s.Endpoint
		} else {
			return rc.opt.Host
		}
	}

	return rc.opt.Host
}

func (rc *remoteConfig) APMHost() string {
	for _, s := range rc.JSON.RemoteSettings {
		if s.Name == apmSetting && s.Endpoint != "" {
			return s.Endpoint
		} else {
			return rc.opt.Host
		}
	}

	return rc.opt.APMHost
}
