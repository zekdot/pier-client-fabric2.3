package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// get is business function which will invoke the to,tid,id
func (broker *Broker) Get(ctx contractapi.TransactionContextInterface) (string, error) {
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
		err := broker.InterchainDataSwapInvoke(ctx, args[0], args[1], args[2])
		return "", err
	default:
		return "", fmt.Errorf("incorrect number of arguments")
	}
}

// get is business function which will invoke the to,tid,id
func (b *Broker) Set(ctx contractapi.TransactionContextInterface) error {
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

// get interchain account for transfer contract: setData from,index,tid,name_id,amount
func (broker *Broker) InterchainSet(ctx contractapi.TransactionContextInterface) error {
	_, args := ctx.GetStub().GetFunctionAndParameters()

	if len(args) < 5 {
		return fmt.Errorf("incorrect number of arguments, expecting 5")
	}

	sourceChainID := args[0]
	sequenceNum := args[1]
	targetCID := args[2]
	key := args[3]
	data := args[4]

	if err := broker.checkIndex(ctx, sourceChainID, sequenceNum, callbackMeta); err != nil {
		return err
	}

	idx, err := strconv.ParseUint(sequenceNum, 10, 64)
	if err != nil {
		return err
	}
	if err := broker.markCallbackCounter(ctx, sourceChainID, idx); err != nil {
		return err
	}

	splitedCID := strings.Split(targetCID, delimiter)
	if len(splitedCID) != 2 {
		return fmt.Errorf("Target chaincode id %s is not valid", targetCID)
	}

	//b := util.ToChaincodeArgs("interchainSet", key, data)
	//response := ctx.GetStub().InvokeChaincode(splitedCID[1], b, splitedCID[0])
	//if response.Status != shim.OK {
	//	return fmt.Errorf("invoke chaincode '%s' err: %s", splitedCID[1], response.Message)
	//}
	err = ctx.GetStub().PutState(key, []byte(data))
	if err != nil {
		return err
	}

	return nil
}

// example for calling get: getData from,index,tid,id
func (broker *Broker) InterchainGet(ctx contractapi.TransactionContextInterface) (string, error) {
	_, args := ctx.GetStub().GetFunctionAndParameters()

	if len(args) < 4 {
		return "", fmt.Errorf("incorrect number of arguments, expecting 4")
	}
	sourceChainID := args[0]
	sequenceNum := args[1]
	targetCID := args[2]
	key := args[3]

	if err := broker.checkIndex(ctx, sourceChainID, sequenceNum, innerMeta); err != nil {
		return "", err
	}

	if err := broker.markInCounter(ctx, sourceChainID); err != nil {
		return "", err
	}

	splitedCID := strings.Split(targetCID, delimiter)
	if len(splitedCID) != 2 {
		return "", fmt.Errorf("Target chaincode id %s is not valid", targetCID)
	}

	//b := util.ToChaincodeArgs("interchainGet", key)
	//response := ctx.GetStub().InvokeChaincode(splitedCID[1], b, splitedCID[0])
	//if response.Status != shim.OK {
	//	return nil, fmt.Errorf("invoke chaincode '%s' err: %s", splitedCID[1], response.Message)
	//}

	// args[0]: key
	dataAsBytes, err := ctx.GetStub().GetState(key)
	if err != nil {
		return "", err
	}

	inKey := inMsgKey(sourceChainID, sequenceNum)
	if err := ctx.GetStub().PutState(inKey, dataAsBytes); err != nil {
		return "", err
	}

	return string(dataAsBytes), nil
}
