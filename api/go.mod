module arteria-s.net/api

go 1.25.1

replace arteria-s.net/googleapi => ./googleapi

replace arteria-s.net/entraapi => ./entraapi

replace arteria-s.net/postgres => ./postgres

require (
	arteria-s.net/entraapi v0.0.0-00010101000000-000000000000
	arteria-s.net/googleapi v0.0.0-00010101000000-000000000000
	arteria-s.net/postgres v0.0.0-00010101000000-000000000000
	github.com/GoogleCloudPlatform/functions-framework-go v1.9.2
	golang.org/x/oauth2 v0.32.0
)

require (
	cloud.google.com/go/compute/metadata v0.9.0 // indirect
	github.com/cloudevents/sdk-go/v2 v2.16.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
)
