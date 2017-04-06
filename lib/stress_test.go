package pewpew

import (
	"io"
	"io/ioutil"
	"os"
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
		{StressConfig{}, ioutil.Discard, true},                                                                                                                     //invalid config
		{StressConfig{}, nil, true},                                                                                                                                //empty writer
		{StressConfig{StressTargets: []StressTarget{{}}}, ioutil.Discard, true},                                                                                    //invalid target
		{StressConfig{StressTargets: []StressTarget{{Count: 10, Concurrency: 1, Target: Target{URL: "*(", RegexURL: true, Method: "GET"}}}}, ioutil.Discard, true}, //error building target, invalid regex
		{StressConfig{StressTargets: []StressTarget{{Count: 10, Concurrency: 1, Target: Target{URL: ":::fail", Method: "GET"}}}}, ioutil.Discard, true},            //error building target

		//good cases
		{StressConfig{StressTargets: []StressTarget{{Count: 1, Concurrency: 1, Target: Target{URL: "http://localhost", Method: "GET"}}, {Count: 1, Concurrency: 1, Target: Target{URL: "http://localhost", Method: "GET"}}}}, ioutil.Discard, false}, //multiple targets
		{StressConfig{StressTargets: []StressTarget{{Count: 1, Concurrency: 1, Target: Target{URL: "http://localhost", Method: "GET"}}}}, ioutil.Discard, false},                                                                                     //single target
		{StressConfig{StressTargets: []StressTarget{{Count: 1, Concurrency: 1, Target: Target{URL: "http://localhost:999999999", Method: "GET"}}}}, ioutil.Discard, false},                                                                           //request that should cause an http err that will get handled
		{StressConfig{StressTargets: []StressTarget{{Count: 1, Concurrency: 1, Target: Target{URL: "http://localhost", Method: "GET"}}}, NoHTTP2: true}, ioutil.Discard, false},                                                                      //noHTTP
		{StressConfig{StressTargets: []StressTarget{{Count: 1, Concurrency: 1, Target: Target{URL: "http://localhost", Method: "GET"}}}, Timeout: "2s"}, ioutil.Discard, false},                                                                      //timeout
		{StressConfig{StressTargets: []StressTarget{{Count: 1, Concurrency: 1, Target: Target{URL: "http://localhost", Method: "GET"}}}, FollowRedirects: true}, ioutil.Discard, false},                                                              //follow redirects
		{StressConfig{StressTargets: []StressTarget{{Count: 1, Concurrency: 1, Target: Target{URL: "http://localhost", Method: "GET"}}}, FollowRedirects: false}, ioutil.Discard, false},                                                             //don't follow redirects
		{StressConfig{StressTargets: []StressTarget{{Count: 1, Concurrency: 1, Target: Target{URL: "http://localhost", Method: "GET"}}}, Verbose: true}, ioutil.Discard, false},                                                                      //verbose
		{StressConfig{StressTargets: []StressTarget{{Count: 1, Concurrency: 1, Target: Target{URL: "http://localhost", Method: "GET"}}}, Quiet: true}, ioutil.Discard, false},                                                                        //quiet
		{StressConfig{StressTargets: []StressTarget{{Count: 1, Concurrency: 1, Target: Target{URL: "http://localhost", Method: "GET"}}}, BodyFilename: tempFilename}, ioutil.Discard, false},                                                         //body file
		{*NewStressConfig(), ioutil.Discard, false},
	}
	for _, c := range cases {
		_, err := RunStress(c.stressConfig, c.writer)
		if (err != nil) != c.hasErr {
			t.Errorf("RunStress(%+v, %q) err: %t wanted %t", c.stressConfig, c.writer, (err != nil), c.hasErr)
		}
	}
}
