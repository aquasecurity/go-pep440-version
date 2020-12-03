module github.com/aquasecurity/go-pep440-version

go 1.15

require (
	github.com/aquasecurity/go-version v0.0.0-20201115065329-578079e4ab05
	github.com/stretchr/testify v1.6.1
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
)

replace github.com/aquasecurity/go-version => ../go-version
