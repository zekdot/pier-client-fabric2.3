package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/strategy"

	//"github.com/hyperledger/fabric-protos-go/common"
	//"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	//"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/meshplus/bitxhub-kit/log"
	"github.com/meshplus/bitxhub-model/pb"
	"github.com/meshplus/pier/pkg/model"
	"github.com/meshplus/pier/pkg/plugins/client"
	"github.com/sirupsen/logrus"
)

var logger = log.NewWithModule("client")

var _ client.Client = (*Client)(nil)

const (
	GetInnerMetaMethod    = "GetInnerMeta"    // get last index of each source chain executing tx
	GetOutMetaMethod      = "GetOuterMeta"    // get last index of each receiving chain crosschain event
	GetCallbackMetaMethod = "GetCallbackMeta" // get last index of each receiving chain callback tx
	GetInMessageMethod    = "GetInMessage"
	GetOutMessageMethod   = "GetOutMessage"
	PollingEventMethod    = "PollingEvent"
	FabricType            = "fabric2.0"
)

type ContractMeta struct {
	EventFilter string `json:"event_filter"`
	Username    string `json:"username"`
	CCID        string `json:"ccid"`
	ChannelID   string `json:"channel_id"`
	ORG         string `json:"org"`
}

type Client struct {
	//meta     *ContractMeta
	//consumer *Consumer
	eventC   chan *pb.IBTP
	pierId   string
	name     string
	outMeta  map[string]uint64
	ticker   *time.Ticker
	done     chan bool
	contract *gateway.Contract
}

func NewClient(configPath, pierId string, extra []byte) (client.Client, error) {
	// read config of fabric
	fabricConfig, err := UnmarshalConfig(configPath)

	eventC := make(chan *pb.IBTP)

	m := make(map[string]uint64)
	if err := json.Unmarshal(extra, &m); err != nil {
		return nil, fmt.Errorf("unmarshal extra for plugin :%w", err)
	}
	if m == nil {
		m = make(map[string]uint64)
	}

	done := make(chan bool)

	err = os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	if err != nil {
		return nil, fmt.Errorf("Error setting DISCOVERY_AS_LOCALHOST environemnt variable: %v", err)
	}
	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		return nil, fmt.Errorf("Failed to create wallet: %v", err)
	}

	if !wallet.Exists(fabricConfig.Username) {
		err = populateWallet(wallet, fabricConfig.OrganizationsPath, fabricConfig.Username)
		if err != nil {
			return nil, fmt.Errorf("Failed to populate wallet contents: %v", err)
		}
	}

	ccpPath := filepath.Join(
		fabricConfig.OrganizationsPath,
		"peerOrganizations",
		"org1.example.com",
		"connection-org1.yaml",
	)

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, fabricConfig.Username),
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to gateway: %v", err)
	}
	defer gw.Close()

	network, err := gw.GetNetwork(fabricConfig.ChannelId)
	if err != nil {
		return nil, fmt.Errorf("Failed to get network: %v", err)
	}

	contract := network.GetContract(fabricConfig.CCID)

	return &Client{
		//consumer: csm,
		eventC:   eventC,
		//meta:     c,
		pierId:   pierId,
		name:     fabricConfig.Name,
		outMeta:  m,
		ticker:   time.NewTicker(2 * time.Second),
		done:     done,
		contract: contract,
	}, nil
}



func (c *Client) Start() error {
	logger.Info("Fabric consumer started")
	go c.polling()
	//return c.consumer.Start()
	return nil;
}

// polling event from broker
func (c *Client) polling() {
	for {
		select {
		case <-c.ticker.C:
			args, err := json.Marshal(c.outMeta)
			if err != nil {
				logger.WithFields(logrus.Fields{
					"error": err.Error(),
				}).Error("Marshal outMeta of plugin")
				continue
			}
			result, err := c.contract.EvaluateTransaction(PollingEventMethod, string(args))

			if err != nil {
				logger.WithFields(logrus.Fields{
					"error": err.Error(),
				}).Error("Polling events from contract")
				continue
			}
			if result == nil {
				continue
			}

			//proof, err := c.getProof(response)
			//if err != nil {
			//	continue
			//}

			evs := make([]*Event, 0)
			if err := json.Unmarshal(result, &evs); err != nil {
				logger.WithFields(logrus.Fields{
					"error": err.Error(),
				}).Error("Unmarshal response payload")
				continue
			}
			for _, ev := range evs {
				//ev.Proof = proof
				c.eventC <- ev.Convert2IBTP(c.pierId, pb.IBTP_INTERCHAIN)
				if c.outMeta == nil {
					c.outMeta = make(map[string]uint64)
				}
				c.outMeta[ev.DstChainID]++
			}
		case <-c.done:
			logger.Info("Stop long polling")
			return
		}
	}
}
//
//func (c *Client) getProof(response channel.Response) ([]byte, error) {
//	var proof []byte
//	var handle = func(response channel.Response) ([]byte, error) {
//		// query proof from fabric
//		l, err := ledger.New(c.consumer.channelProvider)
//		if err != nil {
//			return nil, err
//		}
//
//		t, err := l.QueryTransaction(response.TransactionID)
//		if err != nil {
//			return nil, err
//		}
//		pd := &common.Payload{}
//		if err := proto.Unmarshal(t.TransactionEnvelope.Payload, pd); err != nil {
//			return nil, err
//		}
//
//		pt := &peer.Transaction{}
//		if err := proto.Unmarshal(pd.Data, pt); err != nil {
//			return nil, err
//		}
//
//		return pt.Actions[0].Payload, nil
//	}
//
//	if err := retry.Retry(func(attempt uint) error {
//		var err error
//		proof, err = handle(response)
//		if err != nil {
//			logger.Errorf("can't get proof: %s", err.Error())
//			return err
//		}
//		return nil
//	}, strategy.Wait(2*time.Second)); err != nil {
//		logger.Panicf("can't get proof: %s", err.Error())
//	}
//
//	return proof, nil
//}

func (c *Client) Stop() error {
	c.ticker.Stop()
	c.done <- true
	//return c.consumer.Shutdown()
	return nil
}

func (c *Client) Name() string {
	return c.name
}

func (c *Client) Type() string {
	return FabricType
}

func (c *Client) GetIBTP() chan *pb.IBTP {
	return c.eventC
}

func toPublicFunction(funcName string) string {
	var upperStr string
	vv := []rune(funcName)
	for i := 0; i < len(vv); i++ {
		if i == 0 {
			if vv[i] >= 97 && vv[i] <= 122 {
				vv[i] -= 32
				upperStr += string(vv[i])
			} else {
				fmt.Println("Not begins with lowercase letter,")
				return funcName
			}
		} else {
			upperStr += string(vv[i])
		}
	}
	return upperStr
}

// ToChaincodeArgs converts string args to []byte args
func toChaincodeArgs(args ...string) [][]byte {
	bargs := make([][]byte, len(args))
	for i, arg := range args {
		bargs[i] = []byte(arg)
	}
	return bargs
}

func (c *Client) SubmitIBTP(ibtp *pb.IBTP) (*model.PluginResponse, error) {
	pd := &pb.Payload{}
	ret := &model.PluginResponse{}
	if err := pd.Unmarshal(ibtp.Payload); err != nil {
		return ret, fmt.Errorf("ibtp payload unmarshal: %w", err)
	}
	content := &pb.Content{}
	if err := content.Unmarshal(pd.Content); err != nil {
		return ret, fmt.Errorf("ibtp content unmarshal: %w", err)
	}

	args := make([]string, len(content.Args) + 3)
	args = append(args, ibtp.From, strconv.FormatUint(ibtp.Index, 10), content.DstContractId)
	//args := util.ToChaincodeArgs(ibtp.From, strconv.FormatUint(ibtp.Index, 10), content.DstContractId)
	for i := 0; i < len(content.Args); i ++ {
		args = append(args, string(content.Args[i]))
	}
	funcName := toPublicFunction(content.Func)
	//args = append(args, content.Args...)
	//request := channel.Request{
	//	ChaincodeID: c.meta.CCID,
	//	Fcn:         content.Func,
	//	Args:        args,
	//}

	// retry executing
	var res []byte
	var proof []byte
	var err error
	if err := retry.Retry(func(attempt uint) error {
		res, err = c.contract.SubmitTransaction(funcName, args...)
		//res, err = c.consumer.ChannelClient.Execute(request)
		if err != nil {
			//if strings.Contains(err.Error(), "Chaincode status Code: (500)") {
			//	res.ChaincodeStatus = shim.ERROR
			//	return nil
			//}
			return fmt.Errorf("execute request: %w", err)
		}

		return nil
	}, strategy.Wait(2*time.Second)); err != nil {
		logger.Panicf("Can't send rollback ibtp back to bitxhub: %s", err.Error())
	}

	//response := &Response{}
	//if err := json.Unmarshal(res.Payload, response); err != nil {
	//	return nil, err
	//}

	// if there is callback function, parse returned value
	result := toChaincodeArgs(strings.Split(string(res), ",")...)
	newArgs := make([][]byte, 0)
	ret.Status = true// response.OK
	ret.Message = "success"// response.Message

	// If no callback function to invoke, then simply return
	if content.Callback == "" {
		return ret, nil
	}

	//proof, err = c.getProof(res)
	proof = []byte("success")
	//if err != nil {
	//	return ret, err
	//}

	switch content.Func {
	case "interchainGet":
		newArgs = append(newArgs, content.Args[0])
		newArgs = append(newArgs, result...)
	//case "interchainCharge":
	//	newArgs = append(newArgs, []byte(strconv.FormatBool(response.OK)), content.Args[0])
	//	newArgs = append(newArgs, content.Args[2:]...)
	}

	ret.Result, err = c.generateCallback(ibtp, newArgs, proof)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (c *Client) GetOutMessage(to string, idx uint64) (*pb.IBTP, error) {
	result, err := c.contract.EvaluateTransaction(GetOutMessageMethod, to, strconv.FormatUint(idx, 10))
	if err != nil {
		return nil, err
	}
	ret := &Event{}
	if err := json.Unmarshal(result, ret); err != nil {
		return nil, err
	}
	return ret.Convert2IBTP(c.pierId, pb.IBTP_INTERCHAIN), nil
	//return c.unpackIBTP(&response, pb.IBTP_INTERCHAIN)
}

func (c *Client) GetInMessage(from string, idx uint64) ([][]byte, error) {

	result, err := c.contract.EvaluateTransaction(GetInMessageMethod, from, strconv.FormatUint(idx, 10))
	if err != nil {
		return nil, err
	}
	results := strings.Split(string(result), ",")
	return toChaincodeArgs(results...), nil
}

func (c *Client) GetInMeta() (map[string]uint64, error) {
	result, err := c.contract.EvaluateTransaction(GetInnerMetaMethod)
	if err != nil {
		return nil, err
	}
	m := make(map[string]uint64)
	err = json.Unmarshal(result, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (c *Client) GetOutMeta() (map[string]uint64, error) {
	result, err := c.contract.EvaluateTransaction(GetOutMetaMethod)
	if err != nil {
		return nil, err
	}
	m := make(map[string]uint64)
	err = json.Unmarshal(result, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (c Client) GetCallbackMeta() (map[string]uint64, error) {
	result, err := c.contract.EvaluateTransaction(GetCallbackMetaMethod)
	if err != nil {
		return nil, err
	}
	m := make(map[string]uint64)
	err = json.Unmarshal(result, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (c *Client) CommitCallback(ibtp *pb.IBTP) error {
	return nil
}

func (c *Client) unpackIBTP(response *channel.Response, ibtpType pb.IBTP_Type) (*pb.IBTP, error) {
	ret := &Event{}
	if err := json.Unmarshal(response.Payload, ret); err != nil {
		return nil, err
	}

	return ret.Convert2IBTP(c.pierId, ibtpType), nil
}

func (c *Client) unpackMap(response channel.Response) (map[string]uint64, error) {
	if response.Payload == nil {
		return nil, nil
	}
	r := make(map[string]uint64)
	err := json.Unmarshal(response.Payload, &r)
	if err != nil {
		return nil, fmt.Errorf("unmarshal payload :%w", err)
	}

	return r, nil
}

type handler struct {
	eventFilter string
	eventC      chan *pb.IBTP
	ID          string
}

func newFabricHandler(eventFilter string, eventC chan *pb.IBTP, pierId string) (*handler, error) {
	return &handler{
		eventC:      eventC,
		eventFilter: eventFilter,
		ID:          pierId,
	}, nil
}

func (h *handler) HandleMessage(deliveries *fab.CCEvent, payload []byte) {
	if deliveries.EventName == h.eventFilter {
		e := &pb.IBTP{}
		if err := e.Unmarshal(deliveries.Payload); err != nil {
			return
		}
		e.Proof = payload

		h.eventC <- e
	}
}
