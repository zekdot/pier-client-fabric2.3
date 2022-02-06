package main

import (
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric/common/util"
)

// get interchain account for transfer contract: setData from,index,tid,name_id,amount
func (broker *Broker) interchainSet(ctx contractapi.TransactionContextInterface, args []string) error {
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

	b := util.ToChaincodeArgs("interchainSet", key, data)
	response := ctx.GetStub().InvokeChaincode(splitedCID[1], b, splitedCID[0])
	if response.Status != shim.OK {
		return fmt.Errorf("invoke chaincode '%s' err: %s", splitedCID[1], response.Message)
	}

	return nil
}

// example for calling get: getData from,index,tid,id
func (broker *Broker) interchainGet(ctx contractapi.TransactionContextInterface, args []string) ([]byte, error) {
	if len(args) < 4 {
		return nil, fmt.Errorf("incorrect number of arguments, expecting 4")
	}
	sourceChainID := args[0]
	sequenceNum := args[1]
	targetCID := args[2]
	key := args[3]

	if err := broker.checkIndex(ctx, sourceChainID, sequenceNum, innerMeta); err != nil {
		return nil, err
	}

	if err := broker.markInCounter(ctx, sourceChainID); err != nil {
		return nil, err
	}

	splitedCID := strings.Split(targetCID, delimiter)
	if len(splitedCID) != 2 {
		return nil, fmt.Errorf("Target chaincode id %s is not valid", targetCID)
	}

	b := util.ToChaincodeArgs("interchainGet", key)
	response := ctx.GetStub().InvokeChaincode(splitedCID[1], b, splitedCID[0])
	if response.Status != shim.OK {
		return nil, fmt.Errorf("invoke chaincode '%s' err: %s", splitedCID[1], response.Message)
	}

	inKey := inMsgKey(sourceChainID, sequenceNum)
	if err := ctx.GetStub().PutState(inKey, response.Payload); err != nil {
		return nil, err
	}

	return response.Payload, nil
}
