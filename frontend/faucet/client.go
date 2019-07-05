package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/bgentry/speakeasy"
)

// HTTPClient is used to communicate with the Rivine-based daemon,
// using the exposed (local) REST API over HTTP.
type (
	HTTPClient struct {
		RootURL   string
		Password  string
		UserAgent string
	}

	// HTTPError is return for HTTP Errors by the HTTPClient
	HTTPError struct {
		internalError error
		statusCode    int
	}
)

var (
	// ErrStatusNotFound is returned when status wasn't found.
	ErrStatusNotFound = errors.New("expecting a response, but API returned status code 204 No Content")
)

// PostResp makes a POST API call and decodes the response. An error is
// returned if the response status is not 2xx.
func (c *HTTPClient) PostResp(call, data string, reply interface{}) error {
	resp, err := c.apiPost(call, data)
	if err != nil {
		if httpErr, ok := err.(*HTTPError); ok {
			if httpErr.statusCode == http.StatusForbidden {
				return errUnauthorized
			}
		}
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return errors.New("expecting a response, but API returned status code 204 No Content")
	}

	// auth coin tx - render proper error
	if resp.StatusCode == http.StatusForbidden {
		return errUnauthorized
	}

	err = json.NewDecoder(resp.Body).Decode(&reply)
	if err != nil {
		return err
	}
	return nil
}

// GetAPI makes a GET API call and decodes the response. An error is returned
// if the response status is not 2xx.
func (c *HTTPClient) GetAPI(call string, obj interface{}) error {
	resp, err := c.apiGet(call, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return ErrStatusNotFound
	}

	err = json.NewDecoder(resp.Body).Decode(obj)
	if err != nil {
		return err
	}
	return nil
}

// ApiGet wraps a GET request with a status code check, such that if the GET does
// not return 2xx, the error will be read and returned. When no error is returned,
// the response's body isn't closed, otherwise it is.
func (c *HTTPClient) apiGet(call, data string) (*http.Response, error) {
	resp, err := HTTPGet(c.RootURL+call, data, c.UserAgent)
	if err != nil {
		return nil, errors.New("no response from daemon")
	}
	// check error code
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		// try again using an authenticated HTTP Post call
		password, err := c.apiPassword()
		if err != nil {
			return nil, err
		}
		resp, err = HTTPGETAuthenticated(c.RootURL+call, data, c.UserAgent, password)
		if err != nil {
			return nil, errors.New("no response from daemon - authentication failed")
		}
	}
	if Non2xx(resp.StatusCode) {
		if resp.StatusCode == http.StatusNotFound {
			resp.Body.Close()
			return nil, &HTTPError{
				internalError: errors.New("API call not recognized: " + call),
				statusCode:    resp.StatusCode,
			}
		}
		err := DecodeError(resp)
		resp.Body.Close()
		return nil, &HTTPError{
			internalError: err,
			statusCode:    resp.StatusCode,
		}
	}
	return resp, nil
}

// ApiPost wraps a POST request with a status code check, such that if the POST
// does not return 2xx, the error will be read and returned. When no error is returned,
// the response's body isn't closed, otherwise it is.
func (c *HTTPClient) apiPost(call, data string) (*http.Response, error) {
	resp, err := HTTPPost(c.RootURL+call, data, c.UserAgent)
	if err != nil {
		return nil, errors.New("no response from daemon")
	}
	// check error code
	if resp.StatusCode == http.StatusUnauthorized {
		b, rErr := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if rErr != nil || !c.responseIsAPIPasswordError(b) {
			apiErr, ok := c.responseAsError(b)
			if ok {
				return nil, fmt.Errorf("Unauthorized (401): %s", apiErr.Message)
			}
			return nil, errors.New("API Call failed with the (401) unauthorized status")
		}
		// try again using an authenticated HTTP Post call
		password, err := c.apiPassword()
		if err != nil {
			return nil, err
		}
		resp, err = HTTPPostAuthenticated(c.RootURL+call, data, c.UserAgent, password)
		if err != nil {
			return nil, errors.New("no response from daemon - authentication failed")
		}
	}
	if Non2xx(resp.StatusCode) {
		if resp.StatusCode == http.StatusNotFound {
			resp.Body.Close()
			return nil, &HTTPError{
				internalError: errors.New("API call not recognized: " + call),
				statusCode:    resp.StatusCode,
			}
		}
		err := DecodeError(resp)
		resp.Body.Close()
		return nil, &HTTPError{
			internalError: err,
			statusCode:    resp.StatusCode,
		}
	}
	return resp, nil
}

func (c *HTTPClient) responseIsAPIPasswordError(resp []byte) bool {
	err, ok := c.responseAsError(resp)
	return ok && err.Message == "API Basic authentication failed."
}

func (c *HTTPClient) responseAsError(resp []byte) (Error, bool) {
	var apiError Error
	err := json.Unmarshal(resp, &apiError)
	return apiError, err == nil
}

func (c *HTTPClient) apiPassword() (string, error) {
	if c.Password != "" {
		return c.Password, nil
	}
	var err error
	c.Password, err = speakeasy.Ask("API password: ")
	if err != nil {
		return "", err
	}
	return c.Password, nil
}

// HTTPGet is a utility function for making http get requests to sia with a
// whitelisted user-agent. A non-2xx response does not return an error.
func HTTPGet(url, data, userAgent string) (resp *http.Response, err error) {
	var req *http.Request
	if data != "" {
		req, err = http.NewRequest("GET", url, strings.NewReader(data))
	} else {
		req, err = http.NewRequest("GET", url, nil)
	}
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	return http.DefaultClient.Do(req)
}

// HTTPGETAuthenticated is a utility function for making authenticated http get
// requests to sia with a whitelisted user-agent and the supplied password. A
// non-2xx response does not return an error.
func HTTPGETAuthenticated(url, data, userAgent, password string) (resp *http.Response, err error) {
	var req *http.Request
	if data != "" {
		req, err = http.NewRequest("GET", url, strings.NewReader(data))
	} else {
		req, err = http.NewRequest("GET", url, nil)
	}
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.SetBasicAuth("", password)
	return http.DefaultClient.Do(req)
}

// HTTPPost is a utility function for making post requests to sia with a
// whitelisted user-agent. A non-2xx response does not return an error.
func HTTPPost(url, data, userAgent string) (resp *http.Response, err error) {
	var req *http.Request
	if data != "" {
		req, err = http.NewRequest("POST", url, strings.NewReader(data))
	} else {
		req, err = http.NewRequest("POST", url, nil)
	}
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return http.DefaultClient.Do(req)
}

// HTTPPostAuthenticated is a utility function for making authenticated http
// post requests to sia with a whitelisted user-agent and the supplied
// password. A non-2xx response does not return an error.
func HTTPPostAuthenticated(url, data, userAgent, password string) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("", password)
	return http.DefaultClient.Do(req)
}

// Error implements error.Error
func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d error: %v", e.statusCode, e.internalError)
}

// HTTPStatusCode returns the internal status code,
// returned by the HTTP client in case of an error.
func (e *HTTPError) HTTPStatusCode() int {
	return e.statusCode
}

// Non2xx returns true for non-success HTTP status codes.
func Non2xx(code int) bool {
	return code < 200 || code > 299
}

// DecodeError returns the Error from a API response. This method should
// only be called if the response's status code is non-2xx. The error returned
// may not be of type Error in the event of an error unmarshalling the
// JSON.
func DecodeError(resp *http.Response) error {
	var apiErr Error
	err := json.NewDecoder(resp.Body).Decode(&apiErr)
	if err != nil {
		return err
	}
	return apiErr
}

// Error is a type that is encoded as JSON and returned in an API response in
// the event of an error. Only the Message field is required. More fields may
// be added to this struct in the future for better error reporting.
type Error struct {
	// Message describes the error in English. Typically it is set to
	// `err.Error()`. This field is required.
	Message string `json:"message"`

	// TODO: add a Param field with the (omitempty option in the json tag)
	// to indicate that the error was caused by an invalid, missing, or
	// incorrect parameter. This is not trivial as the API does not
	// currently do parameter validation itself. For example, the
	// /gateway/connect endpoint relies on the gateway.Connect method to
	// validate the netaddress. However, this prevents the API from knowing
	// whether an error returned by gateway.Connect is because of a
	// connection error or an invalid netaddress parameter. Validating
	// parameters in the API is not sufficient, as a parameter's value may
	// be valid or invalid depending on the current state of a module.
}

// Error implements the error interface for the Error type. It returns only the
// Message field.
func (err Error) Error() string {
	return err.Message
}
