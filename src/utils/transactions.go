package utils

import (
	"context"
	"fmt"
	"log"
	"time"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"

	"github.com/gagliardetto/solana-go/rpc"
)

func GetGenesisTransaction(rpcClient *rpc.Client, account solana.PublicKey) string {

	var firstTransactionSigature solana.Signature

	depth := 0
	for !false {
		time.Sleep(1 * time.Second)
		signatures, err := rpcClient.GetSignaturesForAddressWithOpts(context.TODO(), account, func(_depth int, _sig solana.Signature) *rpc.GetSignaturesForAddressOpts {
			if _depth == 0 {
				return nil
			} else {
				return &rpc.GetSignaturesForAddressOpts{
					Before: _sig,
				}
			}
		}(depth, firstTransactionSigature))

		if err != nil {
			panic(err)
		}
		if len(signatures) == 0 && depth != 0 {
			break
		} else {
			firstTransactionSigature = signatures[len(signatures)-1].Signature
		}

		depth += 1
	}

	if firstTransactionSigature.Equals(solana.Signature{}) {
		log.Println("bad account", account, firstTransactionSigature.String())
		return ""
	}

	time.Sleep(4 * time.Second)
	transaction, err := rpcClient.GetTransaction(context.TODO(), firstTransactionSigature, &rpc.GetTransactionOpts{
		Encoding: solana.EncodingBase64,
	})
	if err != nil {
		log.Println("bad transaction signature", firstTransactionSigature.String())
		panic(err)
	}

	if transaction.Transaction == nil {
		log.Println("bad transaction", transaction)
		return ""
	}
	// log.Println(transaction.Transaction,.)
	tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(transaction.Transaction.GetBinary()))
	if err != nil {
		panic(err)
	}

	// lets assume that transaction is never nil
	genesisInvoker := tx.Message.AccountKeys[0]

	return genesisInvoker.String()
}

func SendTx(
	rpcClient *rpc.Client,
	doc string,
	instructions []solana.Instruction,
	signers []solana.PrivateKey,
	feePayer solana.PublicKey,
) {
	recent, err := rpcClient.GetRecentBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		log.Println("PANIC!!!", fmt.Errorf("unable to fetch recent blockhash - %w", err))
		return
	}

	tx, err := solana.NewTransaction(
		instructions,
		recent.Value.Blockhash,
		solana.TransactionPayer(feePayer),
	)
	if err != nil {
		log.Println("PANIC!!!", fmt.Errorf("unable to create transaction"))
		return
	}

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		for _, candidate := range signers {
			if candidate.PublicKey().Equals(key) {
				return &candidate
			}
		}
		return nil
	})
	if err != nil {
		log.Println("PANIC!!!", fmt.Errorf("unable to sign transaction: %w", err))
		return
	}

	/*
		        tx.EncodeTree(text.NewTreeEncoder(os.Stdout, doc))
				sig, err := rpcClient.SimulateTransaction(
					context.TODO(),
					tx,
				)
				if err != nil {
					panic(err)
				}
	*/
	sig, err := rpcClient.SendTransaction(
		context.TODO(),
		tx,
	)
	if err != nil {
		log.Println("PANIC!!!", fmt.Errorf("unable to send transaction - %w", err))
		return
	}
	log.Println(sig)
}
