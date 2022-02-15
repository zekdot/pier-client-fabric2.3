module github.com/meshplus/pier-client-fabric

go 1.13

require (
	github.com/Rican7/retry v0.1.0
	github.com/VividCortex/gohistogram v1.0.0 // indirect
	github.com/cloudflare/cfssl v0.0.0-20180223231731-4e2dcbde5004
	github.com/golang/protobuf v1.5.2
	github.com/golangci/golangci-lint v1.23.0 // indirect
	github.com/google/certificate-transparency-go v1.1.0 // indirect
	github.com/hyperledger/fabric v2.0.1+incompatible
	github.com/hyperledger/fabric-chaincode-go v0.0.0-20210718160520-38d29fabecb9
	github.com/hyperledger/fabric-contract-api-go v1.1.1
	github.com/hyperledger/fabric-lib-go v1.0.0 // indirect
	github.com/hyperledger/fabric-protos-go v0.0.0-20200424173316-dd554ba3746e
	github.com/hyperledger/fabric-sdk-go v1.0.0-alpha5
	github.com/meshplus/bitxhub-kit v1.0.1-0.20200525112026-df2160653e23
	github.com/meshplus/bitxhub-model v1.0.0-rc4.0.20200608065824-2fbc63639e92
	github.com/meshplus/pier v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.5.0
	github.com/spf13/viper v1.6.1
	github.com/hyperledger/fabric-sdk-go v1.0.0-rc1
    	golang.org/x/net v0.0.0-20201021035429-f5854403a974 // indirect
    	golang.org/x/sys v0.0.0-20210119212857-b64e53b001e4 // indirect
    	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
)

replace github.com/golang/protobuf => github.com/golang/protobuf v1.3.2

replace google.golang.org/grpc => google.golang.org/grpc v1.27.1

replace gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.7

replace golang.org/x/net => golang.org/x/net v0.0.0-20200202094626-16171245cfb2

replace github.com/meshplus/pier => ../pier

replace golang.org/x/text => golang.org/x/text v0.3.0

replace golang.org/x/sys => golang.org/x/sys v0.0.0-20200509044756-6aff5f38e54f

replace github.com/spf13/afero => github.com/spf13/afero v1.1.2

replace github.com/spf13/pflag => github.com/spf13/pflag v1.0.5

replace github.com/pelletier/go-toml => github.com/pelletier/go-toml v1.2.0

replace github.com/spf13/jwalterweatherman => github.com/spf13/jwalterweatherman v1.0.0

replace github.com/mholt/archiver => github.com/mholt/archiver v0.0.0-20180417220235-e4ef56d48eb0

//replace github.com/btcsuite/btcd => github.com/btcsuite/btcd v0.20.1-beta
//
//replace google.golang.org/genproto => google.golang.org/genproto v0.0.0-20200218151345-dad8c97a84f5
//
//replace github.com/spf13/cast => github.com/spf13/cast v1.3.0
//
//replace github.com/mitchellh/mapstructure => github.com/mitchellh/mapstructure v1.1.2
