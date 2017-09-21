package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRequestLogFields(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	fields := RequestLogFields(r, 200, time.Duration(100))
	assert.Equal(t, 200, fields["code"])

    // Errno
    fields = RequestLogFields(r, 500, time.Duration(100))
    assert.Equal(t, 999, fields["errno"])
    fields = RequestLogFields(r, 400, time.Duration(100))
    assert.Equal(t, 109, fields["errno"])
    fields = RequestLogFields(r, 401, time.Duration(100))
    assert.Equal(t, 104, fields["errno"])
    fields = RequestLogFields(r, 403, time.Duration(100))
    assert.Equal(t, 121, fields["errno"])
}
