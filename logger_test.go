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
}
