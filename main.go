package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	token_metadata "github.com/gagliardetto/metaplex-go/clients/token-metadata"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"triptych.labs/metaplex-go-cli/src/utils"
)

var MINT solana.PublicKey = solana.MustPublicKeyFromBase58("4rBVnELJeHi9GgvCUi9Jzzr8dnUUSbXB9gtsoHtsuaUn")

func main() {
	op := os.Args[1]
	switch op {
	case "update_royalties_using_hashlist":
		{
			updateRoyaltiesUsingHashList()
		}
	case "create_metadata":
		{
			createMetadata()
		}
	case "update":
		{
			update()
		}
	case "fetch_numbered":
		{
			fetchNumbered()
		}
	default:
		{
			panic(errors.New("invalid operation"))
		}
	}
}

func createMetadata() {
	rpcClient := rpc.New(utils.NETWORK)
	oracle, err := solana.PrivateKeyFromSolanaKeygenFile("./oracle.key")
	if err != nil {
		panic(err)
	}

	mint := solana.MustPublicKeyFromBase58(os.Args[2])
	metadataData, err := ioutil.ReadFile(os.Args[3])
	if err != nil {
		panic(err)
	}
	var metadata token_metadata.DataV2
	err = json.Unmarshal(metadataData, &metadata)
	if err != nil {
		panic(err)
	}

	metadataPda, _ := utils.GetMetadata(mint)

	metadataArgs := token_metadata.CreateMetadataAccountArgsV2{
		Data: token_metadata.DataV2{
			Name:                 metadata.Name,
			Symbol:               metadata.Symbol,
			Uri:                  metadata.Uri,
			SellerFeeBasisPoints: 0,
			Creators:             nil,
			Collection:           nil,
			Uses:                 nil,
		},
		IsMutable: true,
	}

	ix := token_metadata.NewCreateMetadataAccountV2Instruction(
		metadataArgs,
		metadataPda,
		mint,
		oracle.PublicKey(),
		oracle.PublicKey(),
		oracle.PublicKey(),
		solana.SystemProgramID,
		solana.SysVarRentPubkey,
	).Build()

	utils.SendTx(
		rpcClient,
		"airdrop",
		append(make([]solana.Instruction, 0), ix),
		append(make([]solana.PrivateKey, 0), oracle),
		oracle.PublicKey(),
	)

}
func update() {
	rpcClient := rpc.New(utils.NETWORK)
	oracle, err := solana.PrivateKeyFromSolanaKeygenFile("./oracle.key")
	if err != nil {
		panic(err)
	}

	collectionMint := solana.MustPublicKeyFromBase58("")

	metadataPda, _ := utils.GetMetadata(collectionMint)
	metadataPdaData := utils.GetMetadataData(rpcClient, metadataPda)

	utils.UpdateCollectionMetadata(rpcClient, *metadataPdaData, metadataPda, oracle, "", "")

	metadata := token_metadata.CreateMetadataAccountArgsV2{Data: token_metadata.DataV2{Name: "", Symbol: "", Uri: "", SellerFeeBasisPoints: 0, Creators: nil, Collection: nil, Uses: nil}, IsMutable: true}

	ix := token_metadata.NewCreateMetadataAccountV2Instruction(metadata, metadataPda, MINT, oracle.PublicKey(), oracle.PublicKey(), oracle.PublicKey(), solana.SystemProgramID, solana.SysVarRentPubkey).Build()
	utils.SendTx(
		rpcClient,
		"airdrop",
		append(make([]solana.Instruction, 0), ix),
		append(make([]solana.PrivateKey, 0), oracle),
		oracle.PublicKey(),
	)

}

func updateRoyaltiesUsingHashList() {
	rpcClient := rpc.New(utils.NETWORK)
	oracle, err := solana.PrivateKeyFromSolanaKeygenFile(os.Args[2])
	if err != nil {
		panic(err)
	}

	hashListBytes, err := ioutil.ReadFile(os.Args[3])
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	royaltiesArg, err := strconv.Atoi(os.Args[4])
	if err != nil {
		log.Println("Royalties Argument is not a number")
	}

	var hashList []solana.PublicKey
	err = json.Unmarshal(hashListBytes, &hashList)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}
	fmt.Println(len(hashList))

	metadatasPdas := make([]solana.PublicKey, len(hashList))
	for i, mint := range hashList {
		metadatasPdas[i], _ = utils.GetMetadata(mint)
	}

	var wg sync.WaitGroup

	metadatasData := utils.GetMetadatasData(rpcClient, metadatasPdas)
	if metadatasData == nil {
		panic(errors.New("empty metadatas data"))
	}

	royalties := uint16(royaltiesArg)

	for i, metadata := range *metadatasData {
		if i%1000 == 0 {
			log.Println("Sleeping...", i, len(hashList))
			wg.Wait()
			time.Sleep(5 * time.Second)
		}
		wg.Add(1)
		go func(_wg *sync.WaitGroup, _metadata token_metadata.Metadata, _metadataPda solana.PublicKey) {
			utils.UpdateMetadata(rpcClient, _metadata, _metadataPda, oracle, royalties)
			_wg.Done()
		}(&wg, metadata, metadatasPdas[i])
	}

	log.Println("Closing...")
	wg.Wait()

	log.Println("Done.")
}

func fetchNumbered() {
	log.Println("Starting...")
	numbersArg := os.Args[2]
	numbersStrings := strings.Split(numbersArg, ",")
	numbers := make([]int, len(numbersStrings))
	for i, numberString := range numbersStrings {
		numbers[i], _ = strconv.Atoi(numberString)

	}

	rpcClient := rpc.New(utils.NETWORK)
	hashListBytes, err := ioutil.ReadFile("./hashlist.json")
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	// Now let's unmarshall the data into `payload`
	var hashList []solana.PublicKey
	err = json.Unmarshal(hashListBytes, &hashList)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}

	metadataPdas := make([]solana.PublicKey, len(hashList))
	for i, mint := range hashList {
		metadataPdas[i], _ = utils.GetMetadata(mint)
	}

	metadatasData := utils.GetMetadatasData(rpcClient, metadataPdas)
	if metadatasData == nil {
		panic(errors.New("empty metadatas data"))
	}

	// map[metadataPda]tokenMint
	type record struct {
		Mint solana.PublicKey
		Name string
	}
	selection := make(map[solana.PublicKey]record)

	qualifiedAccounts := make([]solana.PublicKey, 0)
	for metadataIndex, metadata := range *metadatasData {
		name := strings.Replace(metadata.Data.Name, "\u0000", "", -1)
		numberString := strings.Split(name, "#")[1]
		number, _ := strconv.Atoi(numberString)
		for _, selectedNumber := range numbers {
			if number == selectedNumber {
				qualifiedAccounts = append(qualifiedAccounts, metadataPdas[metadataIndex])
				selection[metadataPdas[metadataIndex]] = record{
					Mint: metadata.Mint,
					Name: strings.Replace(metadata.Data.Name, "\u0000", "", -1),
				}
			}
		}
	}
	fmt.Println(qualifiedAccounts)

	var wg sync.WaitGroup
	log.Println("sleeping to settle rpc limits")
	time.Sleep(5 * time.Second)
	for i, qualifiedAccount := range qualifiedAccounts {
		if i%50 == 0 && i != 0 {
			log.Println("Sleeping...", i, len(qualifiedAccounts))
			wg.Wait()
			time.Sleep(5 * time.Second)
		}
		wg.Add(1)
		go func(_wg *sync.WaitGroup, _rpcClient *rpc.Client, _index int, _account solana.PublicKey) {

			genesisInvoker := utils.GetGenesisTransaction(_rpcClient, _account)
			fmt.Println(selection[_account].Mint, selection[_account].Name, genesisInvoker)

			_wg.Done()
		}(&wg, rpcClient, i, qualifiedAccount)

	}
	wg.Wait()

	log.Println("Done.")
}

