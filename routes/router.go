package routes

import (
	"backend/contracts"
	"bytes"
	"fmt"
	"strconv"
	"time"

	"encoding/json"
	"net/http"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"

	"backend/ai"
	"backend/database"

	. "backend/config"
)

type Claims struct {
	ID          string                         `json:"id"`
	Messages    []openai.ChatCompletionMessage `json:"messages"`
	Step        int                            `json:"step"`
	Compression *ai.Compression                `json:"compression"`
	jwt.RegisteredClaims
}

func signJWT(claims *Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(Config.Key)
	return tokenString, err
}

const storiesPerToken = 1

func AvailableStories(c *gin.Context) {
	wallet := c.GetString("wallet")

	tokens, err := contracts.TokensOfUser(wallet)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	used, err := database.UsedStories(wallet)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"available": storiesPerToken * len(tokens),
		"used":      used,
	})
}

func Start(c *gin.Context) {
	wallet := c.GetString("wallet")

	tokens, err := contracts.TokensOfUser(wallet)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	used, err := database.UsedStories(wallet)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if used >= storiesPerToken*len(tokens) {
		c.String(http.StatusBadRequest, "Too many stories played, try again tomorrow")
		return
	}

	id := uuid.NewString()
	resp, err := ai.Generate(Config.PromptMessages)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	claims := &Claims{
		ID: id,
		Messages: []openai.ChatCompletionMessage{{
			Role:    "assistant",
			Content: resp.OriginalMessage,
		}},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: nil,
		},
		Step:        1,
		Compression: nil,
	}

	token, err := signJWT(claims)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if err := database.NewStory(wallet, id); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": resp.Message,
		"options": resp.Options,
		"jwt":     token,
		"step":    claims.Step,
	})
}

func Respond(c *gin.Context) {
	type Body struct {
		JWT    string `json:"jwt"`
		Option int    `json:"option"`
	}

	var response Body
	err := c.BindJSON(&response)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(response.JWT, claims, func(token *jwt.Token) (interface{}, error) {
		return Config.Key.Public(), nil
	})
	if err != nil || !tkn.Valid {
		c.String(http.StatusUnauthorized, err.Error())
		return
	}

	step, err := database.GetStep(claims.ID)
	if err != nil {
		c.String(http.StatusUnauthorized, err.Error())
		return
	}

	if step != claims.Step {
		c.String(http.StatusBadRequest, "You already made a choice for this step")
		return
	}

	claims.Step++

	lastNode := ai.Node{}
	lastMessage := claims.Messages[len(claims.Messages)-1].Content
	if err = json.Unmarshal([]byte(lastMessage), &lastNode); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	var message string
	if claims.Step >= Config.MaxSteps {
		message = fmt.Sprintf("This is message number %d, finish the story now. I choose option number %d: %s. Remember to only answer in JSON format.", claims.Step, response.Option, lastNode.Options[response.Option-1])
	} else {
		message = fmt.Sprintf("This is message number %d. I choose option number %d: %s. Remember to only answer in JSON format.", claims.Step, response.Option, lastNode.Options[response.Option-1])
	}

	claims.Messages = append(claims.Messages, openai.ChatCompletionMessage{
		Role:    "user",
		Content: message,
	})

	var messages []openai.ChatCompletionMessage

	if claims.Compression != nil {
		messages = append(messages, Config.CompressionPromptMessages(claims.Compression.Step, claims.Compression.Message)...)
	} else {
		messages = append(messages, Config.PromptMessages...)
	}

	messages = append(messages, claims.Messages...)
	resp, err := ai.Generate(messages)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if len(claims.Messages) > Config.CompressionLimit {
		claims.Compression, err = ai.Compress(messages, claims.Step)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		claims.Messages = nil
	}

	claims.Messages = append(claims.Messages, openai.ChatCompletionMessage{
		Role:    "assistant",
		Content: resp.OriginalMessage,
	})

	token, err := signJWT(claims)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if err := database.SetStep(claims.ID, claims.Step); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": resp.Message,
		"options": resp.Options,
		"end":     resp.End,
		"summary": resp.Summary,
		"jwt":     token,
		"step":    claims.Step,
	})
}

func Image(c *gin.Context) {
	type Body struct {
		JWT string `json:"jwt"`
	}

	var response Body
	err := c.BindJSON(&response)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(response.JWT, claims, func(token *jwt.Token) (interface{}, error) {
		return Config.Key.Public(), nil
	})

	generated, err := database.ImageGenerated(claims.ID)

	if err != nil || !tkn.Valid {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if generated {
		c.String(http.StatusBadRequest, "Image for this step already generated")
		return
	}

	if err != nil || !tkn.Valid {
		c.String(http.StatusUnauthorized, err.Error())
		return
	}

	resp, err := ai.Image(claims.Messages)
	if err != nil || !tkn.Valid {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.String(http.StatusOK, resp)
}

func GetNonce(c *gin.Context) {
	type Body struct {
		Wallet string `json:"wallet"`
	}

	var response Body
	err := c.BindJSON(&response)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	wallet := response.Wallet

	nonce, err := database.GetNonce(wallet)
	if err != nil {
		if err.Error() == "no user" {
			nonceID, err := uuid.NewRandom()
			if err != nil {
				c.String(http.StatusInternalServerError, "Error creating nonce")
				return
			}

			err = database.Register(wallet, nonceID.String())
			if err != nil {
				c.String(http.StatusInternalServerError, "Error creating user")
				return
			}

			nonce = nonceID.String()
		} else {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	}

	c.String(http.StatusOK, nonce)
}

func VerifySignature(wallet, signature, nonce string) (bool, error) {
	nonceHash := crypto.Keccak256Hash([]byte("\x19Ethereum Signed Message:\n" + strconv.Itoa(len(nonce)) + nonce))
	sigBytes, err := hexutil.Decode(signature)
	if err != nil {
		return false, err
	}

	sigBytes[64] -= 27
	pub, err := crypto.Ecrecover(nonceHash.Bytes(), sigBytes)
	if err != nil {
		return false, err
	}

	walletHex, err := hexutil.Decode(wallet)
	if err != nil {
		return false, err
	}

	pubHash := crypto.Keccak256Hash(pub[1:])

	return bytes.Compare(pubHash.Bytes()[12:], walletHex) == 0, nil
}

func Login(context *gin.Context) {
	type Body struct {
		Wallet    string `json:"wallet"`
		Signature string `json:"signature"`
	}

	var response Body
	err := context.BindJSON(&response)
	if err != nil {
		context.String(http.StatusInternalServerError, err.Error())
		return
	}

	nonce, err := database.GetNonce(response.Wallet)
	if err != nil {
		context.String(http.StatusInternalServerError, "Error getting nonce")
		return
	}

	if ok, err := VerifySignature(response.Wallet, response.Signature, nonce); err != nil || !ok {
		context.String(http.StatusUnauthorized, "Verification check error")
		return
	}

	nonceUUID, err := uuid.NewRandom()
	if err != nil {
		context.String(http.StatusInternalServerError, "Error creating nonce")
		return
	}

	session, err := context.Cookie("session")
	if err == nil {
		if status := database.RedisClient.Del(database.RedisContext, "session: "+session); status.Err() != nil {
			context.String(http.StatusInternalServerError, "Error connecting to session database")
			return
		}
	}

	err = database.SetNonce(response.Wallet, nonceUUID.String())
	if err != nil {
		context.String(http.StatusInternalServerError, "Error updating nonce")
		return
	}

	sessionUUID, err := uuid.NewRandom()
	if err != nil {
		context.String(http.StatusInternalServerError, "Error creating session")
		return
	}

	status := database.RedisClient.SetEX(database.RedisContext, "session: "+sessionUUID.String(), response.Wallet, 24*time.Hour)
	if status.Err() != nil {
		context.String(http.StatusInternalServerError, "Error creating session")
		return
	}

	http.SetCookie(context.Writer, &http.Cookie{
		Name:     "session",
		Value:    sessionUUID.String(),
		MaxAge:   int((24 * time.Hour).Seconds()),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	})

	context.String(http.StatusOK, "")
}

func IsAuth(context *gin.Context) {
	session, err := context.Cookie("session")
	if err != nil {
		context.String(http.StatusUnauthorized, "Unauthorized")
		context.Abort()
		return
	}

	result := database.RedisClient.Get(database.RedisContext, "session: "+session)
	wallet, err := result.Result()
	if err != nil {
		context.String(http.StatusUnauthorized, "Unauthorized")
		context.Abort()
		return
	}

	context.Set("wallet", wallet)

	context.Next()
}

func LoggedIn(context *gin.Context) {
	context.String(http.StatusOK, "User logged in")
}
