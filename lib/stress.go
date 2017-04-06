package pewpew

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
)

//so concurrent workers don't interlace messages
var writeLock sync.Mutex

type workerDone struct{}

type (
	//StressConfig is the top level struct that contains the configuration for a stress test
	StressConfig struct {
		Targets []Target
		Verbose bool
		Quiet   bool

		//global target settings

		Count           int
		Concurrency     int
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
	//Target is location to send the HTTP request.
	Target struct {
		URL string
		//Whether or not to interpret the URL as a regular expression string
		//and generate actual target URLs from that
		RegexURL bool
		//How many total requests to make
		Count int
		//How many requests can be happening simultaneously for this Target
		Concurrency int
		Timeout     string
		//A valid HTTP method: GET, HEAD, POST, etc.
		Method string
		//String that is the content of the HTTP body. Empty string is no body.
		Body string
		//A location on disk to read the HTTP body from. Empty string means it will not be read.
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
)

//Reasonable default values for an new StressConfig
const (
	DefaultURL         = "http://localhost"
	DefaultCount       = 10
	DefaultConcurrency = 1
	DefaultTimeout     = "10s"
	DefaultMethod      = "GET"
	DefaultUserAgent   = "pewpew"
)

//NewStressConfig creates a new StressConfig object
//with package defaults
func NewStressConfig() (s *StressConfig) {
	s = &StressConfig{
		Targets: []Target{
			{
				URL:             DefaultURL,
				Count:           DefaultCount,
				Concurrency:     DefaultConcurrency,
				Timeout:         DefaultTimeout,
				Method:          DefaultMethod,
				UserAgent:       DefaultUserAgent,
				FollowRedirects: true,
			},
		},
	}
	return
}

//RunStress starts the stress tests with the provided StressConfig.
//Throughout the test, data is sent to w, useful for live updates.
func RunStress(s StressConfig, w io.Writer) ([][]RequestStat, error) {
	if w == nil {
		return nil, errors.New("nil writer")
	}
	err := validateTargets(s)
	if err != nil {
		return nil, errors.New("invalid configuration: " + err.Error())
	}
	targetCount := len(s.Targets)

	//setup the queue of requests, one queue per target
	requestQueues := make([](chan http.Request), targetCount)
	for idx, target := range s.Targets {
		requestQueues[idx] = make(chan http.Request, target.Count)
		for i := 0; i < target.Count; i++ {
			req, err := buildRequest(target)
			if err != nil {
				return nil, errors.New("failed to create request with target configuration: " + err.Error())
			}
			requestQueues[idx] <- req
		}
		close(requestQueues[idx])
	}

	if targetCount == 1 {
		fmt.Fprintf(w, "Stress testing %d target:\n", targetCount)
	} else {
		fmt.Fprintf(w, "Stress testing %d targets:\n", targetCount)
	}

	//when a target is finished, send all stats into this
	targetStats := make(chan []RequestStat)
	for idx, target := range s.Targets {
		go func(target Target, requestQueue chan http.Request, targetStats chan []RequestStat) {
			writeLock.Lock()
			fmt.Fprintf(w, "- Running %d tests at %s, %d at a time\n", target.Count, target.URL, target.Concurrency)
			writeLock.Unlock()

			workerDoneChan := make(chan workerDone)   //workers use this to indicate they are done
			requestStatChan := make(chan RequestStat) //workers communicate each requests' info

			client := createClient(target)

			//start up the workers
			for i := 0; i < target.Concurrency; i++ {
				go func() {
					for {
						select {
						case req, ok := <-requestQueue:
							if !ok {
								//queue is empty
								workerDoneChan <- workerDone{}
								return
							}

							response, stat := runRequest(req, client)
							if !s.Quiet {
								writeLock.Lock()
								printStat(stat, w)
								if s.Verbose {
									printVerbose(&req, response, w)
								}
								writeLock.Unlock()
							}

							requestStatChan <- stat
						}
					}
				}()
			}
			requestStats := make([]RequestStat, target.Count)
			requestsCompleteCount := 0
			workersDoneCount := 0
			//wait for all workers to finish
			for {
				select {
				case <-workerDoneChan:
					workersDoneCount++
				case stat := <-requestStatChan:
					requestStats[requestsCompleteCount] = stat
					requestsCompleteCount++
				}
				if workersDoneCount == target.Concurrency {
					//all workers are finished
					break
				}
			}
			targetStats <- requestStats
		}(target, requestQueues[idx], targetStats)
	}
	targetRequestStats := make([][]RequestStat, targetCount)
	targetDoneCount := 0
	for {
		select {
		case reqStats := <-targetStats:
			targetRequestStats[targetDoneCount] = reqStats
			targetDoneCount++
		}
		if targetDoneCount == targetCount {
			//all targets are finished
			break
		}
	}

	return targetRequestStats, nil
}
