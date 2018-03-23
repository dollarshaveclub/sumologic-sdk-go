package sumologic

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// https://help.sumologic.com/APIs/Search-Job-API/About-the-Search-Job-API#Creating_a_search_job
// TLDR;
// Rate Limit 240 rpm
// use ISO 8601 for time ranges
// Process Flow
// 1. Request a Search Job - Client.StartSearch(*Search) - query and time range.
// 2. Response - a search job ID or error SearchJob
// 3. Request search status - Client.GetSearchStatus(id int) must be done every 5 in at least
// 4. Response
//      - a job status 'gathering results', 'done executing', etc
//      - message and record counts
// 5. Request - request the results, job does not have to be complete
// 6. Response - JSON search results

// StartSearchRequest is the data needed to start a search
type StartSearchRequest struct {
	Query    string `json:"query"`
	From     string `json:"from"`
	To       string `json:"to"`
	TimeZone string `json:"timeZone"`
}

// SearchJob represents a search job in Sumologic, returned after starting a search.
type SearchJob struct {
	Status  int    `json:"status"`
	ID      string `json:"id,omitempty"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// SearchJobStates are the different states a search job can be in.
var SearchJobStates = map[string]string{
	"NOT STARTED":            "Search job has not been started yet.",
	"GATHERING RESULTS":      "Search job is still gathering more results, however results might already be available.",
	"FORCE PAUSED":           "Query that is paused by the system. It is true only for non-aggregate queries that are paused at the limit of 100k. This limit is dynamic and may vary from customer to customer.",
	"DONE GATHERING RESULTS": "Search job is done gathering results; the entire specified time range has been covered.",
	"CANCELED":               "The search job has been canceled.",
}

// StartSearch calls the Sumologic API Search Endpoint.
// POST search/jobs
func (c *Client) StartSearch(ssr StartSearchRequest) (*SearchJob, []*http.Cookie, error) {
	body, _ := json.Marshal(ssr)

	relativeURL, _ := url.Parse("search/jobs")
	url := c.EndpointURL.ResolveReference(relativeURL)

	req, err := http.NewRequest("POST", url.String(), bytes.NewBuffer(body))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic "+c.AuthToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	responseBody, _ := ioutil.ReadAll(resp.Body)

	switch resp.StatusCode {
	case http.StatusAccepted:
		var sj = new(SearchJob)

		err = json.Unmarshal(responseBody, &sj)
		if err != nil {
			return nil, nil, err
		}
		cookies := resp.Cookies()
		return sj, cookies, nil
	case http.StatusUnauthorized:
		return nil, nil, ErrClientAuthenticationError
	case http.StatusBadRequest:
		var sj = new(SearchJob)
		err = json.Unmarshal(responseBody, &sj)
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, fmt.Errorf("Start SearchJob BadRequest, %v, %v", sj.Code, sj.Message)
	default:
		return nil, nil, fmt.Errorf("unexepected http status code %v", resp.StatusCode)
	}
}

// HistogramBucket corresponds to the histogram display in the Sumo Logic interactive analytics API.
type HistogramBucket struct {
	Length         int `json:"length"`
	Count          int `json:"count"`
	StartTimeStamp int `json:"startTimeStamp"`
}

// SearchJobStatusResponse stores the response from getting a search status.
type SearchJobStatusResponse struct {
	State           string             `json:"state"`
	MessageCount    int                `json:"messageCount"`
	HistgramBuckets []*HistogramBucket `json:"histogramBuckets"`
	RecordCount     int                `json:"recordCount"`
	PendingWarnings []string           `json:"pendingWarnings"`
	PendingErrors   []string           `json:"pendingErrors"`
}

// GetSearchJobStatus retrieves the status of a running job.
func (c *Client) GetSearchJobStatus(searchJobID string, cookies []*http.Cookie) (*SearchJobStatusResponse, error) {

	relativeURL, _ := url.Parse(fmt.Sprintf("search/jobs/%s", searchJobID))
	url := c.EndpointURL.ResolveReference(relativeURL)
	req, err := http.NewRequest("GET", url.String(), nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic "+c.AuthToken)
	for _, v := range cookies {
		req.AddCookie(v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, _ := ioutil.ReadAll(resp.Body)

	switch resp.StatusCode {
	case http.StatusOK:
		var jobStatus = new(SearchJobStatusResponse)
		err = json.Unmarshal(responseBody, &jobStatus)
		if err != nil {
			return nil, err
		}
		return jobStatus, nil
	default:
		return nil, fmt.Errorf("Status not OK : %v", resp.StatusCode)
	}
}

// SearchJobResultsRequest is a wrapper for the search job messages params.
type SearchJobResultsRequest struct {
	ID     string `json:"searchJobId"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
}

// SearchJobResultField is one field from a search result.
type SearchJobResultField struct {
	Name      string `json:"name"`
	FieldType string `json:"fieldType"`
	KeyField  bool   `json:"keyField"`
}

// SearchJobResult represents a search job result
type SearchJobResult struct {
}

// GetSearchResults will retrieve the messages from a finished search job.
func (c *Client) GetSearchResults(sjmr SearchJobResultsRequest) (*SearchJobResult, error) {
	// q := req.URL.Query()
	// q.Add("offset", strconv.Itoa(sjsr.Offset))
	// q.Add("limit", strconv.Itoa(sjsr.Limit))
	// req.URL.RawQuery = q.Encode()
	return nil, errors.New("Not Implemented")
}
