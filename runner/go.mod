module github.com/ChristopherHX/worker-actions-job

go 1.24.11

replace github.com/ChristopherHX/github-act-runner => ../github-act-runner

require github.com/ChristopherHX/github-act-runner v0.13.0

require (
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/google/uuid v1.5.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	golang.org/x/sys v0.38.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
