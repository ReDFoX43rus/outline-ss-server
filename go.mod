module github.com/Jigsaw-Code/outline-ss-server

require (
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/oschwald/geoip2-golang v1.4.0
	github.com/prometheus/client_golang v1.7.1
	github.com/shadowsocks/go-shadowsocks2 v0.1.5
	github.com/stretchr/testify v1.6.1
	golang.org/x/crypto v0.0.0-20220321153916-2c7772ba3064
	gopkg.in/yaml.v2 v2.3.0

	features v1.0.0
)

replace features v1.0.0 => ./features

go 1.14
