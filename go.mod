module github.com/krewire/kiw

go 1.23

require (
	github.com/krewire/framework v0.3.1
	github.com/krewire/guild v0.1.0
	github.com/krewire/libs v0.3.0
	github.com/krewire/ship v0.0.0
)

require gopkg.in/yaml.v3 v3.0.1

require golang.org/x/net v0.30.0 // indirect

require (
	github.com/krewire/mdbind v0.2.0
	github.com/yuin/goldmark v1.8.5 // indirect
)

replace github.com/krewire/ship => ../ship

replace github.com/krewire/libs => ../libs
