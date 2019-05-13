module github.com/lyraproj/servicesdk

require (
	github.com/golang/protobuf v1.3.0
	github.com/hashicorp/go-hclog v0.8.0
	github.com/hashicorp/go-plugin v0.0.0-20190220160451-3f118e8ee104
	github.com/lyraproj/data-protobuf v0.0.0-20190329160005-a909d9e1f93b
	github.com/lyraproj/issue v0.0.0-20190329160035-8bc10230f995
	github.com/lyraproj/pcore v0.0.0-20190502085713-c95bdae56d68
	github.com/lyraproj/semver v0.0.0-20181213164306-02ecea2cd6a2
	github.com/stretchr/testify v1.3.0
	golang.org/x/net v0.0.0-20190311183353-d8887717615a
	google.golang.org/grpc v1.19.0
)

// remove once lyraproj/issue#7 is merged
replace github.com/lyraproj/issue => github.com/thallgren/issue v0.0.0-20190512160618-668e97752cb0
