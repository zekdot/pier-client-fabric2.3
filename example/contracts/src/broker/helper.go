package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"strconv"
)

type Response struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
	Data    []byte `json:"data"`
}

func successResponse(data []byte) *Response {
	res := &Response{
		OK:   true,
		Data: data,
	}

	//data, err := json.Marshal(res)
	//if err != nil {
	//	panic(err)
	//}

	return res
}

func errorResponse(msg string) *Response {
	res := &Response{
		OK:      false,
		Message: msg,
	}

	//data, err := json.Marshal(res)
	//if err != nil {
	//	panic(err)
	//}

	return res
}

// putMap for persisting meta state into ledger
func (broker *Broker) putMap(ctx contractapi.TransactionContextInterface, metaName string, meta map[string]uint64) error {
	if meta == nil {
		return nil
	}

	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(metaName, metaBytes)
}

func (broker *Broker) getMap(ctx contractapi.TransactionContextInterface, metaName string) (map[string]uint64, error) {
	metaBytes, err := ctx.GetStub().GetState(metaName)
	if err != nil {
		return nil, err
	}

	meta := make(map[string]uint64)
	if metaBytes == nil {
		return meta, nil
	}

	if err := json.Unmarshal(metaBytes, &meta); err != nil {
		return nil, err
	}
	return meta, nil
}

func getChaincodeID(ctx contractapi.TransactionContextInterface) (string, error) {
	sp, err := ctx.GetStub().GetSignedProposal()
	if err != nil {
		return "", err
	}

	proposal := &pb.Proposal{}
	if err := proto.Unmarshal(sp.ProposalBytes, proposal); err != nil {
		return "", err
	}

	payload := &pb.ChaincodeProposalPayload{}
	if err := proto.Unmarshal(proposal.Payload, payload); err != nil {
		return "", err
	}

	spec := &pb.ChaincodeInvocationSpec{}
	if err := proto.Unmarshal(payload.Input, spec); err != nil {
		return "", err
	}

	return getKey(ctx.GetStub().GetChannelID(), spec.ChaincodeSpec.ChaincodeId.Name), nil
}

func getKey(channel, chaincodeName string) string {
	return channel + delimiter + chaincodeName
}

func (broker *Broker) checkIndex(ctx contractapi.TransactionContextInterface, addr string, index string, metaName string) error {
	idx, err := strconv.ParseUint(index, 10, 64)
	if err != nil {
		return err
	}
	meta, err := broker.getMap(ctx, metaName)
	if err != nil {
		return err
	}
	if idx != meta[addr]+1 {
		return fmt.Errorf("incorrect index, expect %d", meta[addr]+1)
	}
	return nil
}

func (broker *Broker) outMsgKey(to string, idx string) string {
	return fmt.Sprintf("out-msg-%s-%s", to, idx)
}

func (broker *Broker) inMsgKey(from string, idx string) string {
	return fmt.Sprintf("in-msg-%s-%s", from, idx)
}

func (broker *Broker) getList(ctx contractapi.TransactionContextInterface) *Response {
	whiteList, err := broker.getMap(ctx, whiteList)
	if err != nil {
		return errorResponse(fmt.Sprintf("Get white list :%s", err.Error()))
	}
	var list [][]byte
	for k, v := range whiteList {
		if v == 0 {
			list = append(list, []byte(k))
		}
	}
	return successResponse(bytes.Join(list, []byte(",")))
}