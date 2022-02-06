package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	//pb "github.com/hyperledger/fabric/protos/peer"
)

const (
	interchainEventName = "interchain-event-name"
	innerMeta           = "inner-meta"
	outterMeta          = "outter-meta"
	callbackMeta        = "callback-meta"
	whiteList           = "white-list"
	adminList           = "admin-list"
	//passed              = "1"
	//rejected            = "2"
	delimiter           = "&"
)

type Broker struct{
	contractapi.Contract
}

type Event struct {
	Index         uint64 `json:"index"`
	DstChainID    string `json:"dst_chain_id"`
	SrcContractID string `json:"src_contract_id"`
	DstContractID string `json:"dst_contract_id"`
	Func          string `json:"func"`
	Args          string `json:"args"`
	Callback      string `json:"callback"`
}

func (broker *Broker) Init(ctx contractapi.TransactionContextInterface) error {
	clientID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return err // fmt.Error(fmt.Sprintf("get client id: %s", err.Error()))
	}

	m := make(map[string]uint64)
	m[clientID] = 1
	err = broker.putMap(ctx, adminList, m)
	if err != nil {
		return err
	}

	return broker.initialize(ctx)
}

func (broker *Broker) initialize(ctx contractapi.TransactionContextInterface) error {
	inCounter := make(map[string]uint64)
	outCounter := make(map[string]uint64)
	callbackCounter := make(map[string]uint64)

	if err := broker.putMap(ctx, innerMeta, inCounter); err != nil {
		return err
	}

	if err := broker.putMap(ctx, outterMeta, outCounter); err != nil {
		return err
	}

	if err := broker.putMap(ctx, callbackMeta, callbackCounter); err != nil {
		return err
	}

	return nil
}

func (broker *Broker) InterchainTransferInvoke(ctx contractapi.TransactionContextInterface) error {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) < 3 {
		//return errorResponse("incorrect number of arguments, expecting 5"), nil
		return fmt.Errorf("incorrect number of arguments, expecting 3")
	}
	cid, err := getChaincodeID(ctx)
	if err != nil {
		return err
	}

	newArgs := make([]string, 0)
	newArgs = append(newArgs, args[0], cid, args[1], "interchainCharge", strings.Join(args[2:], ","), "interchainConfirm")

	return broker.interchainInvoke(ctx, newArgs)
}

func (broker *Broker) InterchainDataSwapInvoke(ctx contractapi.TransactionContextInterface,
	toId string, contractId string, key string) error {
	//_, args := ctx.GetStub().GetFunctionAndParameters()
	//if len(args) < 3 {
	//	//return errorResponse("incorrect number of arguments, expecting 5"), nil
	//	return fmt.Errorf("incorrect number of arguments, expecting 3")
	//}
	cid, err := getChaincodeID(ctx)
	if err != nil {
		return err
	}

	newArgs := make([]string, 0)
	// to fromid toid func args callback
	newArgs = append(newArgs, toId, cid, contractId, "interchainGet", key, "interchainSet")

	return broker.interchainInvoke(ctx, newArgs)
}

// InterchainInvoke
// address to,
// address fid,
// address tid,
// string func,
// string args,
// string callback;
func (broker *Broker) interchainInvoke(ctx contractapi.TransactionContextInterface, args[] string) error {
	//_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) < 6 {
		return fmt.Errorf("incorrect number of arguments, expecting 6")
	}

	destChainID := args[0]
	outMeta, err := broker.getMap(ctx, outterMeta)
	if err != nil {
		//return shim.Error(err.Error())
		return err
	}

	if _, ok := outMeta[destChainID]; !ok {
		outMeta[destChainID] = 0
	}

	tx := &Event{
		Index:         outMeta[destChainID] + 1,
		DstChainID:    destChainID,
		SrcContractID: args[1],
		DstContractID: args[2],
		Func:          args[3],
		Args:          args[4],
		Callback:      args[5],
	}

	outMeta[tx.DstChainID]++
	if err := broker.putMap(ctx, outterMeta, outMeta); err != nil {
		//return shim.Error(err.Error())
		return err
	}

	txValue, err := json.Marshal(tx)
	if err != nil {
		//return shim.Error(err.Error())
		return err
	}

	// persist out message
	key := outMsgKey(tx.DstChainID, strconv.FormatUint(tx.Index, 10))
	if err := ctx.GetStub().PutState(key, txValue); err != nil {
		//return shim.Error(fmt.Errorf("persist event: %w", err).Error())
		return err
	}

	if err := ctx.GetStub().SetEvent(interchainEventName, txValue); err != nil {
		//return shim.Error(fmt.Errorf("set event: %w", err).Error())
		return err
	}

	return nil //shim.Success(nil)
}

// polling m(m is the out meta plugin has received. transfer them into string to pass. structure is map[string]uint64
func (broker *Broker) PollingEvent(ctx contractapi.TransactionContextInterface, mStr string) ([]*Event, error) {
	//_, args := ctx.GetStub().GetFunctionAndParameters()
	m := make(map[string]uint64)
	if err := json.Unmarshal([]byte(mStr), &m); err != nil {
		return nil, err
		//return shim.Error(fmt.Errorf("unmarshal out meta: %s", err).Error())
	}
	outMeta, err := broker.getMap(ctx, outterMeta)
	if err != nil {
		//return shim.Error(err.Error())
		return nil, err
	}
	events := make([]*Event, 0)
	for addr, idx := range outMeta {
		startPos, ok := m[addr]
		if !ok {
			startPos = 0
		}
		for i := startPos + 1; i <= idx; i++ {
			eb, err := ctx.GetStub().GetState(outMsgKey(addr, strconv.FormatUint(i, 10)))
			if err != nil {
				fmt.Printf("get out event by key %s fail", outMsgKey(addr, strconv.FormatUint(i, 10)))
				continue
			}
			e := &Event{}
			if err := json.Unmarshal(eb, e); err != nil {
				fmt.Println("unmarshal event fail")
				continue
			}
			events = append(events, e)
		}
	}
	//ret, err := json.Marshal(events)
	//if err != nil {
	//	return nil, err
	//}
	return events, nil
}
//
func main() {
	chaincode, err := contractapi.NewChaincode(new(Broker))

	if err != nil {
		fmt.Printf("Error create broker chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting broker chaincode: %s", err.Error())
	}
}
