package contracts

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"

	. "backend/config"
)

var nftContract *Nft

func Init() error {
	client, err := ethclient.Dial(Config.Contracts.Explorer)
	if err != nil {
		return err
	}

	address := common.HexToAddress(Config.Contracts.Nft)
	nftContract, err = NewNft(address, client)
	if err != nil {
		return err
	}

	return nil
}

func TokensOfUser(userAddress string) ([]*big.Int, error) {
	return nftContract.TokensOfOwner(nil, common.HexToAddress(userAddress))
}
