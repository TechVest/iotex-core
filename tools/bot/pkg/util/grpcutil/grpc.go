package grpcutil

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/iotexproject/go-pkgs/hash"
	"github.com/iotexproject/iotex-antenna-go/v2/iotex"
	"github.com/iotexproject/iotex-core/v2/action"
	"github.com/iotexproject/iotex-proto/golang/iotexapi"
	"github.com/iotexproject/iotex-proto/golang/iotextypes"
)

// ConnectToEndpoint connect to endpoint
func ConnectToEndpoint(url string) (*grpc.ClientConn, error) {
	endpoint := url
	if endpoint == "" {
		return nil, errors.New(`endpoint is empty`)
	}
	return grpc.Dial(endpoint, grpc.WithInsecure())
}

// GetReceiptByActionHash get receipt by action hash
func GetReceiptByActionHash(url, hs string) error {
	conn, err := ConnectToEndpoint(url)
	if err != nil {
		return err
	}
	defer conn.Close()
	c := iotexapi.NewAPIServiceClient(conn)
	if c == nil {
		return errors.New("NewAPIServiceClient error")
	}
	cli := iotex.NewReadOnlyClient(c)
	hash, err := hash.HexStringToHash256(hs)
	if err != nil {
		return err
	}
	caller := cli.GetReceipt(hash)
	response, err := caller.Call(context.Background())
	if err != nil {
		return err
	}
	if response.ReceiptInfo.Receipt.Status != uint64(iotextypes.ReceiptStatus_Success) {
		if response.ReceiptInfo.Receipt.ExecutionRevertMsg != "" {
			return errors.New("action is reverted: " + response.ReceiptInfo.Receipt.ExecutionRevertMsg)
		}
		return errors.New("action failed: " + hs)
	}
	return nil
}

// SendAction send action to endpoint
func SendAction(url string, action *iotextypes.Action) error {
	conn, err := ConnectToEndpoint(url)
	if err != nil {
		return err
	}
	defer conn.Close()
	cli := iotexapi.NewAPIServiceClient(conn)
	req := &iotexapi.SendActionRequest{Action: action}
	if _, err = cli.SendAction(context.Background(), req); err != nil {
		return err
	}
	return nil
}

// GetNonce get nonce of address
func GetNonce(url string, address string) (nonce uint64, err error) {
	conn, err := ConnectToEndpoint(url)
	if err != nil {
		return
	}
	defer conn.Close()
	cli := iotexapi.NewAPIServiceClient(conn)
	request := iotexapi.GetAccountRequest{Address: address}
	response, err := cli.GetAccount(context.Background(), &request)
	if err != nil {
		return
	}
	nonce = response.AccountMeta.PendingNonce
	return
}

// FixGasLimit estimate action gas
func FixGasLimit(url string, caller string, execution *action.Execution) (uint64, error) {
	conn, err := ConnectToEndpoint(url)
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	cli := iotexapi.NewAPIServiceClient(conn)
	request := &iotexapi.EstimateActionGasConsumptionRequest{
		Action: &iotexapi.EstimateActionGasConsumptionRequest_Execution{
			Execution: execution.Proto(),
		},
		CallerAddress: caller,
	}
	res, err := cli.EstimateActionGasConsumption(context.Background(), request)
	if err != nil {
		return 0, err
	}
	return res.Gas, nil
}
