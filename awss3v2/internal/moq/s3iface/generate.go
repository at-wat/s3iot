package mock_s3iface

//go:generate go run github.com/matryer/moq -pkg mock_s3iface -out generated.go ./iface HTTPClient:MockHTTPClient
