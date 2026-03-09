package testutils

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func AssertJSONResponse(t *testing.T, expectedCode int, expectedJSON string, resp *http.Response) {
	t.Helper()
	assert.Equal(t, expectedCode, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	assert.JSONEq(t, expectedJSON, string(body))
}

func StringJSON(data any) string {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return string(jsonBytes)
}
