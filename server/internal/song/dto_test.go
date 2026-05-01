package song

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ctxWithQuery(q string) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/?"+q, nil)
	return c
}

// parseBaseSongFilter

func TestParseBaseSongFilter_Defaults(t *testing.T) {
	f, err := parseBaseSongFilter(ctxWithQuery(""))
	require.NoError(t, err)
	assert.Equal(t, SortCreatedAt, f.Sort)
	assert.Equal(t, SortDesc, f.Order)
	assert.Equal(t, 20, f.Limit)
	assert.Equal(t, 0, f.Offset)
	assert.Nil(t, f.UploadedBy)
}

func TestParseBaseSongFilter_IgnoresUploadedByQuery(t *testing.T) {
	// The base parser does NOT touch UploadedBy — handlers own that scope.
	f, err := parseBaseSongFilter(ctxWithQuery("uploaded_by=" + uuid.New().String()))
	require.NoError(t, err)
	assert.Nil(t, f.UploadedBy)
}

func TestParseBaseSongFilter_InvalidSort(t *testing.T) {
	_, err := parseBaseSongFilter(ctxWithQuery("sort=nope"))
	require.Error(t, err)
	ferr, ok := err.(*FilterError)
	require.True(t, ok)
	_, has := ferr.Fields["sort"]
	assert.True(t, has)
}

func TestParseBaseSongFilter_InvalidOrder(t *testing.T) {
	_, err := parseBaseSongFilter(ctxWithQuery("order=sideways"))
	require.Error(t, err)
}

func TestParseBaseSongFilter_DateRangeMustBeOrdered(t *testing.T) {
	_, err := parseBaseSongFilter(
		ctxWithQuery("created_after=2026-05-01T00:00:00Z&created_before=2026-01-01T00:00:00Z"),
	)
	require.Error(t, err)
	ferr, _ := err.(*FilterError)
	_, has := ferr.Fields["created_before"]
	assert.True(t, has)
}

func TestParseBaseSongFilter_LimitClamped(t *testing.T) {
	f, err := parseBaseSongFilter(ctxWithQuery("limit=9999"))
	require.NoError(t, err)
	assert.Equal(t, 100, f.Limit)
}

func TestParseBaseSongFilter_LongQRejected(t *testing.T) {
	long := make([]byte, 201)
	for i := range long {
		long[i] = 'a'
	}
	_, err := parseBaseSongFilter(ctxWithQuery("q=" + string(long)))
	require.Error(t, err)
}

// parseUploadedByQuery

func TestParseUploadedByQuery_Absent(t *testing.T) {
	id, err := parseUploadedByQuery(ctxWithQuery(""))
	require.NoError(t, err)
	assert.Nil(t, id)
}

func TestParseUploadedByQuery_Valid(t *testing.T) {
	want := uuid.New()
	got, err := parseUploadedByQuery(ctxWithQuery("uploaded_by=" + want.String()))
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, want, *got)
}

func TestParseUploadedByQuery_Invalid(t *testing.T) {
	_, err := parseUploadedByQuery(ctxWithQuery("uploaded_by=not-a-uuid"))
	require.Error(t, err)
	ferr, ok := err.(*FilterError)
	require.True(t, ok)
	_, has := ferr.Fields["uploaded_by"]
	assert.True(t, has)
}
