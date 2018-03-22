package sumologic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

//https://help.sumologic.com/APIs/Search-Job-API/About-the-Search-Job-API#Creating_a_search_job
// TLDR;
// Rate Limit 240 rpm
// use ISO 8601 for time ranges
// Process Flow
// 1. Request a Search Job - Client.StartSearch(*Search) - query and time range.
// 2. Response - a search job ID or error
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

// SearchJob represents a search job in Sumologic
type SearchJob struct {
	Status  int    `json:"status"`
	ID      string `json:"id,omitempty"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// StartSearch calls the Sumologic API Search Endpoint.
// POST search/jobs
func (c *Client) StartSearch(ssr StartSearchRequest) (*SearchJob, error) {
	body, _ := json.Marshal(ssr)

	relativeURL, _ := url.Parse("search/jobs")
	url := c.EndpointURL.ResolveReference(relativeURL)

	req, err := http.NewRequest("POST", url.String(), bytes.NewBuffer(body))
	req.Header.Add("Authorization", "Basic "+c.AuthToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, _ := ioutil.ReadAll(resp.Body)

	switch resp.StatusCode {
	case http.StatusAccepted:
		var sj = new(SearchJob)
		err = json.Unmarshal(responseBody, &sj)
		if err != nil {
			return nil, err
		}
		return sj, nil
	case http.StatusUnauthorized:
		return nil, ErrClientAuthenticationError
	default:
		return nil, fmt.Errorf("unexecpted http status code %v", resp.StatusCode)
	}
}
