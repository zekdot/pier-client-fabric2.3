package main

import (
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const (
	channelID            = "mychannel"
	brokerContractName   = "broker"
	interchainInvokeFunc = "InterchainDataSwapInvoke"
)

type DataSwapper struct{
	contractapi.Contract
}

func (s *DataSwapper) Init(ctx contractapi.TransactionContextInterface) error {
	return nil
}

// get is business function which will invoke the to,tid,id
func (s *DataSwapper) Get(ctx contractapi.TransactionContextInterface) (string, error) {
	_, args := ctx.GetStub().GetFunctionAndParameters()

	switch len(args) {
	case 1:
		// args[0]: key
		dataAsBytes, err := ctx.GetStub().GetState(args[0])
		if err != nil {
			return "", err
		}

		return string(dataAsBytes), nil
	case 3:
		// args[0]: destination appchain id
		// args[1]: destination contract address
		// args[2]: key
		b := toChaincodeArgs(interchainInvokeFunc, args[0], args[1], args[2])
		response := ctx.GetStub().InvokeChaincode(brokerContractName, b, channelID)

		if response.Status != shim.OK {
			return "", fmt.Errorf("invoke broker chaincode %s error: %s", brokerContractName, response.Message)
		}

		return response.Message, nil
	default:
		return "", fmt.Errorf("incorrect number of arguments")
	}
}

// get is business function which will invoke the to,tid,id
func (s *DataSwapper) Set(ctx contractapi.TransactionContextInterface) error {
	_, args := ctx.GetStub().GetFunctionAndParameters()

	if len(args) != 2 {
		return fmt.Errorf("incorrect number of arguments")
	}

	err := ctx.GetStub().PutState(args[0], []byte(args[1]))
	if err != nil {
		return err
	}

	return nil
}

// interchainSet is the callback function getting data by interchain
func (s *DataSwapper) InterchainSet(ctx contractapi.TransactionContextInterface) error {
	return s.Set(ctx)
}

// interchainGet gets data by interchain
func (s *DataSwapper) InterchainGet(ctx contractapi.TransactionContextInterface) (string, error) {
	return s.Get(ctx)
}

func toChaincodeArgs(args ...string) [][]byte {
	bargs := make([][]byte, len(args))
	for i, arg := range args {
		bargs[i] = []byte(arg)
	}
	return bargs
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(DataSwapper))

	if err != nil {
		fmt.Printf("Error create data_swapper chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting data_swapper chaincode: %s", err.Error())
	}
}
