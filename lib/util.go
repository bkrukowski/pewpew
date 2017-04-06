package pewpew

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	reggen "github.com/lucasjones/reggen"
)

func validateTargets(s StressConfig) error {
	if len(s.Targets) == 0 {
		return errors.New("zero targets")
	}
	for _, target := range s.Targets {
		//checks
		if target.URL == "" {
			return errors.New("empty URL")
		}
		if target.Count <= 0 {
			return errors.New("request count must be greater than zero")
		}
		if target.Concurrency <= 0 {
			return errors.New("concurrency must be greater than zero")
		}
		if target.Method == "" {
			return errors.New("method cannot be empty string")
		}
		if target.Timeout != "" {
			//TODO should save this parsed duration so don't have to inefficiently reparse later
			timeout, err := time.ParseDuration(target.Timeout)
			if err != nil {
				return errors.New("failed to parse timeout: " + target.Timeout)
			}
			if timeout <= time.Millisecond {
				return errors.New("timeout must be greater than one millisecond")
			}
		}
		if target.Concurrency > target.Count {
			return errors.New("concurrency must be higher than request count")
		}
	}
	return nil
}

//splits on delim into parts and trims whitespace
//delim1 splits the pairs, delim2 splits amongst the pairs
//like parseKeyValString("key1: val2, key3 : val4,key5:val6 ", ",", ":") becomes
//["key1"]->"val2"
//["key3"]->"val4"
//["key5"]->"val6"
func parseKeyValString(keyValStr, delim1, delim2 string) (map[string]string, error) {
	m := make(map[string]string)
	if delim1 == delim2 {
		return m, errors.New("delimiters can't be equal")
	}
	pairs := strings.SplitN(keyValStr, delim1, -1)
	for _, pair := range pairs {
		parts := strings.SplitN(pair, delim2, 2)
		if len(parts) != 2 {
			return m, errors.New("failed to parse into two parts")
		}
		key, val := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		if key == "" || val == "" {
			return m, errors.New("key or value is empty")
		}
		m[key] = val
	}
	return m, nil
}

//build the http request out of the target's config
func buildRequest(t Target) (http.Request, error) {
	if t.URL == "" {
		return http.Request{}, errors.New("empty URL")
	}
	if len(t.URL) < 8 {
		return http.Request{}, errors.New("URL too short")
	}
	//prepend "http://" if scheme not provided
	//maybe a cleaner way to do this via net.url?
	if t.URL[:7] != "http://" && t.URL[:8] != "https://" {
		t.URL = "http://" + t.URL
	}
	var urlStr string
	var err error
	//when regex set, generate urls
	if t.RegexURL {
		urlStr, err = reggen.Generate(t.URL, 10)
		if err != nil {
			return http.Request{}, errors.New("failed to parse regex: " + err.Error())
		}
	} else {
		urlStr = t.URL
	}
	URL, err := url.Parse(urlStr)
	if err != nil {
		return http.Request{}, errors.New("failed to parse URL " + urlStr + " : " + err.Error())
	}
	if URL.Host == "" {
		return http.Request{}, errors.New("empty hostname")
	}

	//setup the request
	var req *http.Request
	if t.BodyFilename != "" {
		fileContents, err := ioutil.ReadFile(t.BodyFilename)
		if err != nil {
			return http.Request{}, errors.New("failed to read contents of file " + t.BodyFilename + ": " + err.Error())
		}
		req, err = http.NewRequest(t.Method, URL.String(), bytes.NewBuffer(fileContents))
	} else if t.Body != "" {
		req, err = http.NewRequest(t.Method, URL.String(), bytes.NewBuffer([]byte(t.Body)))
	} else {
		req, err = http.NewRequest(t.Method, URL.String(), nil)
	}
	if err != nil {
		return http.Request{}, errors.New("failed to create request: " + err.Error())
	}
	//add headers
	if t.Headers != "" {
		headerMap, err := parseKeyValString(t.Headers, ",", ":")
		if err != nil {
			return http.Request{}, errors.New("could not parse headers: " + err.Error())
		}
		for key, val := range headerMap {
			req.Header.Add(key, val)
		}
	}

	req.Header.Set("User-Agent", t.UserAgent)

	//add cookies
	if t.Cookies != "" {
		cookieMap, err := parseKeyValString(t.Cookies, ";", "=")
		if err != nil {
			return http.Request{}, errors.New("could not parse cookies: " + err.Error())
		}
		for key, val := range cookieMap {
			req.AddCookie(&http.Cookie{Name: key, Value: val})
		}
	}

	if t.BasicAuth != "" {
		authMap, err := parseKeyValString(t.BasicAuth, ",", ":")
		if err != nil {
			return http.Request{}, errors.New("could not parse basic auth: " + err.Error())
		}
		for key, val := range authMap {
			req.SetBasicAuth(key, val)
			break
		}
	}
	return *req, nil
}
