package v2handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Tests require sample data in db

func TestTagsByPostHandler(t *testing.T) {
	// TODO request example slug from db
	req, err := http.NewRequest("GET", "/v2.0/post_tags/succinct_summation_of_weeks_events_81817/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(TagsByPostHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v instead of %v", status, http.StatusOK)
	}

	expected := `
	[{"title":"claims","slug":"claims","post_count":0},{"title":"initial","slug":"initial","post_count":0},{"title":"Jobless","slug":"jobless","post_count":0},{"title":"june","slug":"june","post_count":0},{"title":"remain","slug":"remain","post_count":0},{"title":"retail","slug":"retail","post_count":0},{"title":"revised","slug":"revised","post_count":0},{"title":"rose","slug":"rose","post_count":0},{"title":"sales","slug":"sales","post_count":0},{"title":"week","slug":"week","post_count":0}]`
	if rr.Body.String() != expected {
		t.Errorf("TagsByPostHandler handler returned unexpected body: got %v instead of %v", rr.Body.String(), expected)
	}

	if ctype := rr.Header().Get("Content-Type"); ctype != "application/json" {
		t.Errorf("TagsByPostHandler content type header does not match: got %v instead of %v", ctype, "application/json")
	}
}
