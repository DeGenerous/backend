package discord

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"encoding/json"
	"net/http"

	"github.com/bwmarrin/discordgo"

	. "backend/config"
)

type response struct {
	Data struct {
		SettledsSimpleAuctions []struct {
			Price         string `json:"price"`
			SimpleAuction struct {
				AssetId string `json:"assetId"`
			} `json:"simpleAuction"`
		} `json:"settledsSimpleAuctions"`
	} `json:"data"`
}

type imgResponse struct {
	Data struct {
		Tokens []struct {
			Metadata struct {
				Image string `json:"image"`
			} `json:"metadata"`
		} `json:"tokens"`
	} `json:"data"`
}

func getQuery(limit int) (string, *bytes.Buffer) {
	const gqlUrl = "https://squid.subsquid.io/raresama-auction-exosama/graphql"
	const gqlQuery = `
		query settledSimpleAuctions {
		  settledsSimpleAuctions(orderBy: endTime_DESC, where: {assetAddress_eq: \"%s\"}, limit: %d) {
			price
			simpleAuction {
			  assetId
			}
		  }
		}
	`

	query := strings.Replace(strings.Replace(fmt.Sprintf(gqlQuery, Config.ContractAddress, limit), "\n", " ", -1), "\t", " ", -1)
	jsonQuery := fmt.Sprintf("{\"query\": \"%s\"}", query)

	return gqlUrl, bytes.NewBuffer([]byte(jsonQuery))
}

func getImageQuery(id int) (string, *bytes.Buffer) {
	const gqlUrl = "https://squid.subsquid.io/raresama-nft-exosama/graphql"
	const gqlQuery = `
		query getNftStatsPage {
		  tokens(where: {contract: {id_eq: \"%s\"}, numericId_eq: %d}, limit: 1) {
			metadata {
			  image
			}
		  }
		}
	`

	query := strings.Replace(strings.Replace(fmt.Sprintf(gqlQuery, Config.ContractAddress, id), "\n", " ", -1), "\t", " ", -1)
	jsonQuery := fmt.Sprintf("{\"query\": \"%s\"}", query)

	return gqlUrl, bytes.NewBuffer([]byte(jsonQuery))
}

func formatBigInt(number string) string {
	return number[:len(number)-18]
}

func latestSales(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	url, query := getQuery(5)
	resp, err := http.Post(url, "application/json", query)
	if err != nil {
		fmt.Println(err)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	var r response
	if err := json.Unmarshal(body, &r); err != nil {
		fmt.Println(err)
		return
	}

	var fields []*discordgo.MessageEmbedField

	for _, nft := range r.Data.SettledsSimpleAuctions {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Inception Ark #" + nft.SimpleAuction.AssetId,
			Value:  formatBigInt(nft.Price) + " $SAMA",
			Inline: false,
		})
	}

	err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:  "Latest sales:",
					Fields: fields,
				},
			},
		},
	})

	if err != nil {
		fmt.Println(err)
		return
	}
}

var latestSale = ""

func explorerInit() error {
	url, query := getQuery(5)
	resp, err := http.Post(url, "application/json", query)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}

	var r response
	if err := json.Unmarshal(body, &r); err != nil {
		fmt.Println(err)
		return err
	}

	latestSale = r.Data.SettledsSimpleAuctions[0].SimpleAuction.AssetId
	return nil
}

func tick() {
	for range time.Tick(time.Second * 15) {
		newSale()
	}
}

func newSale() {
	url, query := getQuery(5)
	resp, err := http.Post(url, "application/json", query)
	if err != nil {
		fmt.Println(err)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	var r response
	if err := json.Unmarshal(body, &r); err != nil {
		fmt.Println(err)
		return
	}

	if latestSale != r.Data.SettledsSimpleAuctions[0].SimpleAuction.AssetId {
		latestSale = r.Data.SettledsSimpleAuctions[0].SimpleAuction.AssetId

		id, err := strconv.Atoi(r.Data.SettledsSimpleAuctions[0].SimpleAuction.AssetId)
		if err != nil {
			fmt.Println(err)
			return
		}

		url, query := getImageQuery(id)
		resp, err := http.Post(url, "application/json", query)
		if err != nil {
			fmt.Println(err)
			return
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			return
		}

		var img imgResponse
		if err := json.Unmarshal(body, &img); err != nil {
			fmt.Println(err)
			return
		}

		_, err = session.ChannelMessageSendEmbed(Config.ChannelID, &discordgo.MessageEmbed{
			Title: "New Sale",
			Description: fmt.Sprintf("Inception Ark #%s\nBought for: %s $SAMA",
				r.Data.SettledsSimpleAuctions[0].SimpleAuction.AssetId,
				formatBigInt(r.Data.SettledsSimpleAuctions[0].Price),
			),
			Image: &discordgo.MessageEmbedImage{
				URL: "https://gateway.moonsama.com/ipfs/QmZQFXFZ2KWmAiHPUajez1mbxyPqsiRLYGw4oftQKxhb15",
			},
		})

		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
