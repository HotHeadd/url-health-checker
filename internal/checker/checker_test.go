package checker

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckOne(t *testing.T) {
	tests := []struct {
		name           string
		handlerStatus  int
		wantStatusCode int
		wantErr        bool
		url            string
	}{
		{"success", 200, 200, false, ""},
		{"server error", 500, 500, false, ""},
		{"not found", 404, 404, false, ""},
		{"bad url", 0, 0, true, "broken_url"},
	}
	c := NewChecker()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := tt.url
			if url == "" {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(tt.handlerStatus)
				}))
				t.Cleanup(server.Close)
				url = server.URL
			}
			res := c.CheckOne(t.Context(), url)

			assert.Equal(t, tt.wantStatusCode, res.StatusCode)
			if tt.wantErr {
				assert.NotEmpty(t, res.Err)
			} else {
				assert.Empty(t, res.Err)
			}
		})
	}
}

func TestCheckAll(t *testing.T) {
	type TestRes struct {
		url        string
		statusCode int
		hasError   bool
	}
	type SingleTest struct {
		name         string
		urls         []string
		expected     []TestRes
		serverAnswer int
	}
	tests := []SingleTest{
		{
			"success twice",
			[]string{"", ""},
			[]TestRes{{url: "", statusCode: 200, hasError: false}, {url: "", statusCode: 200, hasError: false}},
			http.StatusOK,
		},
		{
			"success and broken",
			[]string{"", "broken_url"},
			[]TestRes{{url: "", statusCode: 200, hasError: false}, {url: "broken_url", statusCode: 0, hasError: true}},
			http.StatusOK,
		},
	}
	c := NewChecker()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.serverAnswer)
			}))
			t.Cleanup(server.Close)

			urls := tt.urls
			for i := range urls {
				if urls[i] == "" {
					urls[i] = server.URL
				}
			}

			results := c.CheckAll(t.Context(), urls)

			for i := range results {
				hasError := results[i].Err != ""
				if tt.expected[i].hasError {
					assert.True(t, hasError)
				} else {
					assert.False(t, hasError)
				}
				assert.Equal(t, tt.expected[i].statusCode, results[i].StatusCode)
				if tt.expected[i].url != "" {
					assert.Equal(t, tt.expected[i].url, results[i].URL)
				}
			}
		})
	}
}
