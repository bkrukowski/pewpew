package pewpew

import (
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

const tempFilename = "/tmp/testdata"

func TestMain(m *testing.M) {
	//setup
	//create a temp file on disk for use as post body filename
	err := ioutil.WriteFile(tempFilename, []byte(""), 0644)
	if err != nil {
		os.Exit(1)
	}

	retCode := m.Run()

	//teardown
	err = os.Remove(tempFilename)
	if err != nil {
		os.Exit(1)
	}

	os.Exit(retCode)
}

func TestRunStress(t *testing.T) {
	cases := []struct {
		stressConfig StressConfig
		writer       io.Writer
		hasErr       bool
	}{
		{StressConfig{}, ioutil.Discard, true},                                                                                         //invalid config
		{StressConfig{}, nil, true},                                                                                                    //empty writer
		{StressConfig{Targets: []Target{{}}}, ioutil.Discard, true},                                                                    //invalid target
		{StressConfig{Targets: []Target{{URL: "*(", RegexURL: true, Method: "GET", Count: 10, Concurrency: 1}}}, ioutil.Discard, true}, //error building target, invalid regex
		{StressConfig{Targets: []Target{{URL: ":::fail", Method: "GET", Count: 10, Concurrency: 1}}}, ioutil.Discard, true},            //error building target

		//good cases
		{StressConfig{Targets: []Target{{URL: "http://localhost", Method: "GET", Count: 1, Concurrency: 1}, {URL: "http://localhost", Method: "GET", Count: 1, Concurrency: 1}}}, ioutil.Discard, false}, //multiple targets
		{StressConfig{Targets: []Target{{URL: "http://localhost", Method: "GET", Count: 1, Concurrency: 1}}}, ioutil.Discard, false},                                                                     //single target
		{StressConfig{Targets: []Target{{URL: "http://localhost:999999999", Method: "GET", Count: 1, Concurrency: 1}}}, ioutil.Discard, false},                                                           //request that should cause an http err that will get handled
		{StressConfig{Targets: []Target{{URL: "http://localhost", Method: "GET", Count: 1, Concurrency: 1}}, NoHTTP2: true}, ioutil.Discard, false},                                                      //noHTTP
		{StressConfig{Targets: []Target{{URL: "http://localhost", Method: "GET", Count: 1, Concurrency: 1}}, Timeout: "2s"}, ioutil.Discard, false},                                                      //timeout
		{StressConfig{Targets: []Target{{URL: "http://localhost", Method: "GET", Count: 1, Concurrency: 1}}, FollowRedirects: true}, ioutil.Discard, false},                                              //follow redirects
		{StressConfig{Targets: []Target{{URL: "http://localhost", Method: "GET", Count: 1, Concurrency: 1}}, FollowRedirects: false}, ioutil.Discard, false},                                             //don't follow redirects
		{StressConfig{Targets: []Target{{URL: "http://localhost", Method: "GET", Count: 1, Concurrency: 1}}, Verbose: true}, ioutil.Discard, false},                                                      //verbose
		{StressConfig{Targets: []Target{{URL: "http://localhost", Method: "GET", Count: 1, Concurrency: 1}}, Quiet: true}, ioutil.Discard, false},                                                        //quiet
		{StressConfig{Targets: []Target{{URL: "http://localhost", Method: "GET", Count: 1, Concurrency: 1}}, BodyFilename: tempFilename}, ioutil.Discard, false},                                         //body file
		{*NewStressConfig(), ioutil.Discard, false},
	}
	for _, c := range cases {
		_, err := RunStress(c.stressConfig, c.writer)
		if (err != nil) != c.hasErr {
			t.Errorf("RunStress(%+v, %q) err: %t wanted %t", c.stressConfig, c.writer, (err != nil), c.hasErr)
		}
	}
}
