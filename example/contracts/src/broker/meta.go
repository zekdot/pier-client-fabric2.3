package main

import (
	"encoding/json"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// getOutMeta
func (broker *Broker) GetOuterMeta(ctx contractapi.TransactionContextInterface)  (map[string]uint64, error)  {
	meta, err := broker.getMap(ctx, outterMeta)
	if err != nil {
		return nil, err
	}
	return meta, nil
}

func (broker *Broker) GetInnerMeta(ctx contractapi.TransactionContextInterface) (map[string]uint64, error) {
	meta, err := broker.getMap(ctx, innerMeta)
	if err != nil {
		return nil, err
	}
	return meta, nil
}

func (broker *Broker) GetCallbackMeta(ctx contractapi.TransactionContextInterface) (map[string]uint64, error) {
	meta, err := broker.getMap(ctx, callbackMeta)
	if err != nil {
		return nil, err
	}
	return meta, nil
}

// TODO modify return type to Event
// getOutMessage to,index
func (broker *Broker) GetOutMessage(ctx contractapi.TransactionContextInterface, destChainID string, sequenceNum string) (string, error) {
	key := outMsgKey(destChainID, sequenceNum)
	v, err := ctx.GetStub().GetState(key)
	if err != nil {
		return "", err
	}
	if v == nil {
		return "", nil
	}
	var res string
	err = json.Unmarshal(v, res)
	if err != nil {
		return "", err
	}
	return res, nil
	//if err != nil {
	//	return errorResponse(err.Error())
	//}
	//return successResponse(v)
}

// getInMessage from,index
func (broker *Broker) GetInMessage(ctx contractapi.TransactionContextInterface, sourceChainID string, sequenceNum string) (string, error) {
	key := inMsgKey(sourceChainID, sequenceNum)
	v, err := ctx.GetStub().GetState(key)
	if err != nil {
		return "", err
	}
	if v == nil {
		return "", nil
	}
	var res string
	err = json.Unmarshal(v, res)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (broker *Broker) markInCounter(ctx contractapi.TransactionContextInterface, from string) error {
	inMeta, err := broker.getMap(ctx, innerMeta)
	if err != nil {
		return err
	}

	inMeta[from]++
	return broker.putMap(ctx, innerMeta, inMeta)
}

func (broker *Broker) markCallbackCounter(ctx contractapi.TransactionContextInterface, from string, index uint64) error {
	meta, err := broker.getMap(ctx, callbackMeta)
	if err != nil {
		return err
	}

	meta[from] = index

	return broker.putMap(ctx, callbackMeta, meta)
}
