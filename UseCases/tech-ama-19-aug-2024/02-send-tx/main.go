package main

import (
	"context"
	"github.com/pkg/errors"
	"github.com/qubic/go-node-connector/types"
	"github.com/qubic/qubic-http/protobuff"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
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
	// source id(pk/seed), dest id, amount

	targetTick := tickInfoRes.TickInfo.Tick + 5
	log.Printf("target tick: %d\n", targetTick)

	tx, err := types.NewSimpleTransferTransaction(
		"ZFXNTCRVPEUTHEMWRWWPMVEPNXLBHDNKMJGDTNCUQGDCCRJUEAHKIVNFGJKK",
		"FDVORCTKJZVEBFYUXRVUHMPXLMADKSQKAOXLEXUASDGNXXGSXDIACIGHPYSF",
		1,
		targetTick,
	)
	if err != nil {
		return errors.Wrap(err, "creating new transaction")
	}
	err = tx.Sign("vifgmbvarewaclfbmgnxkrnijdmrjdksaglooamqfbxvfhystpqpnde")
	if err != nil {
		return errors.Wrap(err, "signing tx")
	}

	id, err := tx.ID()
	if err != nil {
		return errors.Wrap(err, "getting tx id")
	}

	encodedTx, err := tx.EncodeToBase64()
	if err != nil {
		return errors.Wrap(err, "encoding tx")
	}

	log.Printf("tx ID: %s\n", id)
	
	// curl -X POST https://testapi.qubic.org/v1/broadcast-transaction -d `{"encodedTransaction": ""}`
	broadcastTxRes, err := grpcLiveClient.BroadcastTransaction(ctx, &protobuff.BroadcastTransactionRequest{EncodedTransaction: encodedTx})
	if err != nil {
		return errors.Wrap(err, "broadcasting tx")
	}

	log.Printf("%+v\n", broadcastTxRes)

	return nil
}
