package utils

import (
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
)

type TokensMetadataResponse struct {
	TokenMintMeta token.Mint       `json:"tokenMintMeta"`
	TokenMetadata MetadataResponse `json:"tokenMetadata"`
}

func FindTokenResponse(tokenMint solana.PublicKey, mintsData map[solana.PublicKey]token.Mint, metadatasData []MetadataResponse) TokensMetadataResponse {
	mintMeta := mintsData[tokenMint]
	metadata := func() MetadataResponse {
		for _, metadata := range metadatasData {
			if metadata.Mint == tokenMint {
				return metadata
			}
		}
		panic("unforeseen")
	}()

	return TokensMetadataResponse{
		TokenMintMeta: mintMeta,
		TokenMetadata: metadata,
	}
}
