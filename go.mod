module my5G-RANTester

go 1.26.2

require (
	github.com/aead/cmac v0.0.0-20160719120800-7af84192f0b1
	github.com/davecgh/go-spew v1.1.1
	github.com/free5gc/go-gtp5gnl v1.6.2
	github.com/free5gc/nas v1.3.0
	github.com/free5gc/ngap v1.2.0
	github.com/free5gc/openapi v1.3.0
	github.com/free5gc/util v1.4.0
	github.com/goccy/go-yaml v1.16.0
	github.com/gopacket/gopacket v1.3.1
	github.com/ishidawataru/sctp v0.0.0-20250303034628-ecf9ed6df987
	github.com/khirono/go-nl v1.0.5
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.10.0
	github.com/tetratelabs/wazero v1.9.0
	github.com/urfave/cli/v2 v2.27.6
	github.com/vishvananda/netlink v1.3.0
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.6 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.2 // indirect
	github.com/khirono/go-genl v1.0.1 // indirect
	github.com/khirono/go-rtnllink v1.1.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/vishvananda/netns v0.0.5 // indirect
	github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1 // indirect
	golang.org/x/net v0.55.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// Native v1.3.0 codec fixes; see third_party/free5gc-nas/README.packetrusher.md.
replace github.com/free5gc/nas => ./third_party/free5gc-nas
