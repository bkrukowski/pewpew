package pewpew

import (
	"errors"
	"io"
)

type (
	//BenchConfig is the top level struct that contains the configuration for a bench test
	BenchConfig struct {
		BenchTargets []BenchTarget
		Verbose      bool
		Quiet        bool

		//global target settings

		BeginRPS int
		EndRPS   int
		Interval int
		Duration string
		Cooldown string

		Timeout         string
		Method          string
		Body            string
		BodyFilename    string
		Headers         string
		Cookies         string
		UserAgent       string
		BasicAuth       string
		Compress        bool
		KeepAlive       bool
		FollowRedirects bool
		NoHTTP2         bool
		EnforceSSL      bool
	}
	//BenchTarget combines bench related configuration with a Target configuration
	BenchTarget struct {
		Target Target
	}
)

//NewBenchConfig creates a new BenchConfig
//with package defaults
func NewBenchConfig() (s *BenchConfig) {
	s = &BenchConfig{
		BeginRPS: DefaultBeginRPS,
		EndRPS:   DefaultEndRPS,
		Interval: DefaultInterval,
		Duration: DefaultDuration,
		Cooldown: DefaultCooldown,
		BenchTargets: []BenchTarget{
			{
				Target: Target{
					URL:             DefaultURL,
					Timeout:         DefaultTimeout,
					Method:          DefaultMethod,
					UserAgent:       DefaultUserAgent,
					FollowRedirects: true,
				},
			},
		},
	}
	return
}

//RunBench starts the bench tests with the provided BenchConfig.
//Throughout the test, data is sent to w, useful for live updates.
func RunBench(b BenchConfig, w io.Writer) ([][]RequestStat, error) {
	if w == nil {
		return nil, errors.New("nil writer")
	}
	return nil, nil
}
