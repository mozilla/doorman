package config

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/mozilla/doorman/doorman"
)

type ReloadResponse struct {
	Success bool
	Message string
}

func TestReloadHandler(t *testing.T) {
	reloadReq, _ := http.NewRequest("POST", "/__reload__", nil)
	var resp ReloadResponse

	tmpfile, _ := ioutil.TempFile("", "")
	defer os.Remove(tmpfile.Name()) // clean up

	tmpfile.Write([]byte(`
service: a
policies:
  -
    id: "1"
    action: update
`))

	d := doorman.NewDefaultLadon()
	handler := reloadHandler([]string{tmpfile.Name()})

	// Reload same file twice.
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(doorman.DoormanContextKey, d)
		c.Request = reloadReq

		handler(c)

		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp.Success)
	}

	// Reload bad file.
	tmpfile.Write([]byte("*some$bad@cont\tent"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(doorman.DoormanContextKey, d)
	c.Request = reloadReq

	handler(c)

	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, w.Code, 500)
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Message, "did not find expected alphabetic or numeric character")

	// Reload bad definition (unknown condition type).
	tmpfile.Write([]byte(`
service: a
policies:
  -
    id: "1"
    action: update
    conditions:
      owner:
        type: fantastic
`))
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Set(doorman.DoormanContextKey, d)
	c.Request = reloadReq

	handler(c)

	assert.Equal(t, w.Code, 500)
}
