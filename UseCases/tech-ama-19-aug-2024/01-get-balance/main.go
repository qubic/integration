package main

import (
	"context"
	"github.com/pkg/errors"
	"github.com/qubic/qubic-http/protobuff"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	id := os.Args[1]

	grpcClient, err := grpc.NewClient("213.170.135.5:8004", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return errors.Wrap(err, "creating new grpc client")
	}

	grpcLiveClient := protobuff.NewQubicLiveServiceClient(grpcClient)

	ctx := context.Background()

	// https://testapi.qubic.org/v1/tick-info
	tickInfoRes, err := grpcLiveClient.GetTickInfo(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "getting tick info")
	}

	log.Printf("tick info: %+v", tickInfoRes)

	// https://testapi.qubic.org/v1/balances/ZFXNTCRVPEUTHEMWRWWPMVEPNXLBHDNKMJGDTNCUQGDCCRJUEAHKIVNFGJKK
	balanceRes, err := grpcLiveClient.GetBalance(ctx, &protobuff.GetBalanceRequest{Id: id})
	if err != nil {
		return errors.Wrap(err, "getting balance")
	}

	log.Printf("balance: %+v", balanceRes)

	return nil
}
