package utils

import (
	"context"
	"errors"
	"log"

	ag_binary "github.com/gagliardetto/binary"
	token_metadata "github.com/gagliardetto/metaplex-go/clients/token-metadata"
	"github.com/gagliardetto/solana-go/rpc"

	"github.com/gagliardetto/solana-go"
)

type MetadataResponse struct {
	Name   string           `json:"name"`
	Symbol string           `json:"symbol"`
	Mint   solana.PublicKey `json:"mint"`
}

func GetMetadata(mint solana.PublicKey) (solana.PublicKey, error) {
	addr, _, err := solana.FindProgramAddress(
		[][]byte{
			[]byte("metadata"),
			token_metadata.ProgramID.Bytes(),
			mint.Bytes(),
		},
		token_metadata.ProgramID,
	)
	return addr, err
}

func GetMasterEdition(mint solana.PublicKey) (solana.PublicKey, error) {
	addr, _, err := solana.FindProgramAddress(
		[][]byte{
			[]byte("metadata"),
			token_metadata.ProgramID.Bytes(),
			mint.Bytes(),
			[]byte("edition"),
		},
		token_metadata.ProgramID,
	)
	return addr, err
}

func GetMetadataData(rpcClient *rpc.Client, metadataPda solana.PublicKey) *token_metadata.Metadata {
	bin, _ := rpcClient.GetAccountInfoWithOpts(context.TODO(), metadataPda, &rpc.GetAccountInfoOpts{Commitment: "confirmed"})
	if bin == nil {
		return nil
	}
	var data token_metadata.Metadata
	decoder := ag_binary.NewBorshDecoder(bin.Value.Data.GetBinary())
	err := data.UnmarshalWithDecoder(decoder)
	if err != nil {
		panic(err)
	}

	return &data

}

func GetMetadatasData(rpcClient *rpc.Client, metadataPdas []solana.PublicKey) *[]token_metadata.Metadata {
	metadatasData := make([]token_metadata.Metadata, 0)

	for i := range make([]int, 100) {
		response, err := rpcClient.GetMultipleAccounts(context.TODO(), metadataPdas[i*100:(i+1)*100]...)
		if err != nil {
			panic(err)
		}
		if response == nil {
			return nil
		}
		if len(response.Value) == 0 {
			return nil
		}

		for _, bin := range response.Value {
			var data token_metadata.Metadata
			decoder := ag_binary.NewBorshDecoder(bin.Data.GetBinary())
			err := data.UnmarshalWithDecoder(decoder)
			if err != nil {
				panic(err)
			}

			metadatasData = append(metadatasData, data)
		}
	}

	if len(metadataPdas) != len(metadatasData) {
		panic(errors.New("incomplete metadatas data"))
	}

	return &metadatasData

}

func UpdateMetadata(rpcClient *rpc.Client, metadata token_metadata.Metadata, metadataPda solana.PublicKey, updateAuthority solana.PrivateKey, royalties uint16) {

	if metadata.Data.SellerFeeBasisPoints == royalties {
		log.Println("royalties match", metadata.Data.Name)
		return
	}
	metadata.Data.SellerFeeBasisPoints = royalties

	updateIx := token_metadata.NewUpdateMetadataAccountV2Instruction(
		token_metadata.UpdateMetadataAccountArgsV2{
			Data: &token_metadata.DataV2{
				Name:                 metadata.Data.Name,
				Symbol:               metadata.Data.Symbol,
				Uri:                  metadata.Data.Uri,
				SellerFeeBasisPoints: metadata.Data.SellerFeeBasisPoints,
				Creators:             metadata.Data.Creators,
				Collection:           metadata.Collection,
				Uses:                 metadata.Uses,
			},
		},
		metadataPda,
		updateAuthority.PublicKey(),
	)

	SendTx(
		rpcClient,
		"update metadata",
		append(make([]solana.Instruction, 0), updateIx.Build()),
		append(make([]solana.PrivateKey, 0), updateAuthority),
		updateAuthority.PublicKey(),
	)

	log.Println("Completed for", metadata.Data.Name)
}

func UpdateCollectionMetadata(rpcClient *rpc.Client, metadata token_metadata.Metadata, metadataPda solana.PublicKey, updateAuthority solana.PrivateKey, name, uri string) {

	metadata.Data.Name = name
	metadata.Data.Uri = uri

	updateIx := token_metadata.NewUpdateMetadataAccountV2Instruction(
		token_metadata.UpdateMetadataAccountArgsV2{
			Data: &token_metadata.DataV2{
				Name:                 metadata.Data.Name,
				Symbol:               metadata.Data.Symbol,
				Uri:                  metadata.Data.Uri,
				SellerFeeBasisPoints: metadata.Data.SellerFeeBasisPoints,
				Creators:             metadata.Data.Creators,
				Collection:           metadata.Collection,
				Uses:                 metadata.Uses,
			},
		},
		metadataPda,
		updateAuthority.PublicKey(),
	)

	SendTx(
		rpcClient,
		"update metadata",
		append(make([]solana.Instruction, 0), updateIx.Build()),
		append(make([]solana.PrivateKey, 0), updateAuthority),
		updateAuthority.PublicKey(),
	)

	log.Println("Completed for", metadata.Data.Name)
}
