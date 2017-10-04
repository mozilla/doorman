package warden

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain defined in warden_test.go
// func TestMain(m *testing.M) {}

func TestVerifyJWT(t *testing.T) {
	var err error

	validator := newJWTValidator("https://minimal-demo-iam.auth0.com/")

	r, _ := http.NewRequest("GET", "/", nil)

	_, err = verifyJWT(validator, r)
	require.NotNil(t, err)
	assert.Equal(t, "Token not found", err.Error())

	r.Header.Set("Authorization", "Basic abc")
	_, err = verifyJWT(validator, r)
	require.NotNil(t, err)
	assert.Equal(t, "Token not found", err.Error())

	r.Header.Set("Authorization", "Bearer abc zy")
	_, err = verifyJWT(validator, r)
	require.NotNil(t, err)
	assert.Equal(t, "square/go-jose: compact JWS format must have three parts", err.Error())

	r.Header.Set("Authorization", "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6Ik5EZzFOemczTlRFeVEwVTFNMEZCTnpCQlFqa3hOVVk1UTBVMU9USXpOalEzUXpVek5UWkRNQSJ9.eyJpc3MiOiJodHRwczovL21pbmltYWwtZGVtby1pYW0uYXV0aDAuY29tLyIsInN1YiI6Imdvb2dsZS1vYXV0aDJ8MTA0MTAyMzA2MTExMzUwNTc2NjI4IiwiYXVkIjpbImh0dHA6Ly9taW5pbWFsLWRlbW8taWFtLmxvY2FsaG9zdDo4MDAwIiwiaHR0cHM6Ly9taW5pbWFsLWRlbW8taWFtLmF1dGgwLmNvbS91c2VyaW5mbyJdLCJpYXQiOjE1MDY2MDQzMTMsImV4cCI6MTUwNjYxMTUxMywiYXpwIjoiV1lSWXBKeVM1RG5EeXhMVFJWR0NRR0NXR28yS05RTE4iLCJzY29wZSI6Im9wZW5pZCBwcm9maWxlIn0.JmfQajLJ6UMU8sGwv-4FyN0hAPjlLnixoVXAJwn9-985Y4jnMNiG22RWAk5qsdhxVKjIsyQFGA2oHuKELfcrI-LEHX3dxePxx9jSGUdC1wzk3p2q3YCRwIV3DUFEtBVeml8gdB9V7tVBE6XDivfq7RphiC8c5zz28_vlB2iPPaAwfucJLc1d5t83xlBaSYU9-hWDet3HbgjQg4zvFat6C2-CuKkCuQEG92tsOdoD8RIJtlWmLiMVUhCFgr3pGa7_ZNiKmMFkgZiDsX2qqD107CfOLG3IutcLGCqlpHxOuVltGZNp3QCXwtjIoZSV-5IXssXKLYuz-75GpfEAmUB5fg")
	_, err = verifyJWT(validator, r)
	require.NotNil(t, err)
	assert.Equal(t, "square/go-jose/jwt: validation failed, token is expired (exp)", err.Error())
}
