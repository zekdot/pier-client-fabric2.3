package main

import (
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"path/filepath"
)


type Service struct {
	contract *gateway.Contract
}

func (s *Service) init() error {
	log.Println("============ application-golang starts ============")
	//reqConfig, err := UnmarshalConfig(".")
	//if err != nil {
	//	return err
	//}
	//log.Println(reqConfig)
	reqConfig := &Fabric{
		Addr:        "localhost:7053",
		OrganizationsPath:        "/home/zekdot/bitxhub/fabric-samples/test-network/organizations",
		Username:    "appuser",
		CCID:        "broker",
		ChannelId:   "mychannel",
	}

	err := os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	if err != nil {
		return err
	}
	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}

	if !wallet.Exists(reqConfig.Username) {
		err = populateWallet(wallet, reqConfig.OrganizationsPath, reqConfig.Username)
		if err != nil {
			log.Fatalf("Failed to populate wallet contents: %v", err)
		}
	}

	ccpPath := filepath.Join(
		reqConfig.OrganizationsPath,
		"peerOrganizations",
		"org1.example.com",
		"connection-org1.yaml",
	)
	//log.Println(ccpPath)

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, reqConfig.Username),
	)
	if err != nil {
		//log.Fatalf("Failed to connect to gateway: %v", err)
		return err
	}
	defer gw.Close()

	network, err := gw.GetNetwork(reqConfig.ChannelId)
	if err != nil {
		//log.Fatalf("Failed to get network: %v", err)
		return err
	}

	contract := network.GetContract(reqConfig.CCID)
	s.contract = contract
	return nil
}

type ReqArgs struct {
	FuncName string
	Args []string
}

func (s *Service) SubmitTransaction(req *ReqArgs, reply *string) error{
	var result []byte
	var err error
	//if len(req.Args) == 0 {
	//	result, err = s.contract.SubmitTransaction(req.FuncName)
	//} else {
	result, err = s.contract.SubmitTransaction(req.FuncName, req.Args...)
	//}

	if err != nil {
		return err
	}
	*reply = string(result)
	return nil
}

func (s *Service) EvaluateTransaction(req *ReqArgs, reply *string) error{
	var result []byte
	var err error
	//if len(req.Args) == 0 {
	//	result, err = s.contract.EvaluateTransaction(req.FuncName)
	//} else {
	result, err = s.contract.EvaluateTransaction(req.FuncName, req.Args...)
	//}
	if err != nil {
		return err
	}
	*reply = string(result)
	return nil
}

func populateWallet(wallet *gateway.Wallet, organizationsPath string, username string) error {
	credPath := filepath.Join(
		organizationsPath,
		"peerOrganizations",
		"org1.example.com",
		"users",
		"User1@org1.example.com",
		"msp",
	)

	certPath := filepath.Join(credPath, "signcerts", "User1@org1.example.com-cert.pem")
	// read the certificate pem
	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	keyDir := filepath.Join(credPath, "keystore")
	// there's a single file in this dir containing the private key
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return err
	}
	if len(files) != 1 {
		return fmt.Errorf("keystore folder should have contain one file")
	}
	keyPath := filepath.Join(keyDir, files[0].Name())
	key, err := ioutil.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}

	identity := gateway.NewX509Identity("Org1MSP", string(cert), string(key))

	return wallet.Put(username, identity)
}


func main() {
	service := new(Service)
	err := service.init()
	if err != nil {
		fmt.Errorf(err.Error())
		//return
	}
	rpc.Register(service)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":1212")
	if e != nil {
		log.Fatal("listen error: ", e)
	}
	http.Serve(l, nil)
}