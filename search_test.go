package sumologic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

//https://help.sumologic.com/APIs/Search-Job-API/About-the-Search-Job-API#Creating_a_search_job
func TestStartSearch(t *testing.T) {
	testStartSearch := StartSearchRequest{
		Query:    "_sourceCategory=test/sumo",
		From:     fmt.Sprintf(time.Now().UTC().Format(time.RFC3339)),
		To:       fmt.Sprintf(time.Now().UTC().Format(time.RFC3339)),
		TimeZone: "PST",
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		if r.Method != "POST" {
			t.Errorf("Expected ‘POST’ request, got ‘%s’", r.Method)
		}
		expectedURL := fmt.Sprintf("/search/jobs")
		if r.URL.EscapedPath() != expectedURL {
			t.Errorf("Expected request to ‘%s’, got ‘%s’", expectedURL, r.URL.EscapedPath())
		}
		body, _ := json.Marshal(SearchJob{
			Status:  202,
			ID:      "TestStartSearchJob",
			Code:    "searchjob.valid",
			Message: "Search Running",
		})
		w.Write(body)
	}))
	defer ts.Close()

	c, err := NewClient("accessToken", ts.URL)
	if err != nil {
		t.Errorf("NewClient() returned an error: %s", err)
		return
	}

	startSearchResponse, err := c.StartSearch(testStartSearch)
	if err != nil {
		t.Errorf("StartSearch() returned an error: %s", err)
		return
	}

	if startSearchResponse.Message != "Search Running" {
		t.Errorf("StartSearch() expected message 'Search Running', got `%v`", startSearchResponse.Message)
		return
	}
}

func TestGetSearchStatus(t *testing.T) {
	// req.Header.Set("Cookie", "name=xxxx; count=x")
	testSearchJob := SearchJob{
		ID: "testsearchjob",
	}
	testSearchJobStatusRequest := SearchJobStatusRequest{
		ID:     testSearchJob.ID,
		Offset: 0,
		Limit:  100,
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if r.Method != "GET" {
			t.Errorf("Expected ‘GET’ request, got ‘%s’", r.Method)
		}
		expectedURL := fmt.Sprintf("/search/jobs/%s", testSearchJob.ID)
		if r.URL.EscapedPath() != expectedURL {
			t.Errorf("Expected request to ‘%s’, got ‘%s’", expectedURL, r.URL.EscapedPath())
		}
		body, _ := json.Marshal(SearchJobStatusResponse{
			State: "GATHERING RESULTS",
		})
		w.Write(body)
	}))
	defer ts.Close()

	c, err := NewClient("accessToken", ts.URL)
	if err != nil {
		t.Errorf("NewClient() returned an error: %s", err)
		return
	}

	_, err = c.GetSearchJobStatus(testSearchJobStatusRequest)
	if err != nil {
		t.Errorf("GetSearchJobStatus() returned an error: %s", err)
		return
	}

}
