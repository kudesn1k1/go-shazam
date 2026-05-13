package pagination

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func paramsFromURL(rawURL string) Params {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	var captured Params
	r.GET("/probe", func(c *gin.Context) {
		captured = ParseParams(c)
		c.Status(200)
	})
	req := httptest.NewRequest(http.MethodGet, rawURL, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return captured
}

func TestParseParams(t *testing.T) {
	cases := []struct {
		name string
		url  string
		want Params
	}{
		{"defaults",          "/probe",                   Params{Page: DefaultPage, Limit: DefaultLimit}},
		{"valid",             "/probe?page=3&limit=50",   Params{Page: 3, Limit: 50}},
		{"negative page",     "/probe?page=-1",           Params{Page: DefaultPage, Limit: DefaultLimit}},
		{"zero page",         "/probe?page=0",            Params{Page: DefaultPage, Limit: DefaultLimit}},
		{"zero limit",        "/probe?limit=0",           Params{Page: DefaultPage, Limit: DefaultLimit}},
		{"negative limit",    "/probe?limit=-5",          Params{Page: DefaultPage, Limit: DefaultLimit}},
		{"over-max clamped",  "/probe?limit=500",         Params{Page: DefaultPage, Limit: MaxLimit}},
		{"max-boundary",      "/probe?limit=100",         Params{Page: DefaultPage, Limit: MaxLimit}},
		{"non-numeric page",  "/probe?page=abc",          Params{Page: DefaultPage, Limit: DefaultLimit}},
		{"non-numeric limit", "/probe?limit=xyz",         Params{Page: DefaultPage, Limit: DefaultLimit}},
		{"both invalid",      "/probe?page=abc&limit=xyz", Params{Page: DefaultPage, Limit: DefaultLimit}},
		{"empty params",      "/probe?page=&limit=",      Params{Page: DefaultPage, Limit: DefaultLimit}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := paramsFromURL(tc.url)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestNewResponse_NilDataBecomesEmptySlice(t *testing.T) {
	resp := NewResponse[int](nil, 0, 1, 20)
	assert.NotNil(t, resp.Data)
	assert.Len(t, resp.Data, 0)
	assert.Equal(t, 0, resp.Total)
}

func TestNewResponse_PassthroughFields(t *testing.T) {
	resp := NewResponse([]string{"a", "b"}, 100, 5, 10)
	assert.Equal(t, []string{"a", "b"}, resp.Data)
	assert.Equal(t, 100, resp.Total)
	assert.Equal(t, 5, resp.Page)
	assert.Equal(t, 10, resp.Limit)
}
