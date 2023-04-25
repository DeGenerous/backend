package routes

import (
	"backend/ai"
	"github.com/sashabaranov/go-openai"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"

	. "backend/config"
)

type Claims struct {
	ID       string                         `json:"id"`
	Messages []openai.ChatCompletionMessage `json:"messages"`
	jwt.RegisteredClaims
}

func signJWT(claims *Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(Config.Key)
	return tokenString, err
}

func Start(c *gin.Context) {
	id := uuid.NewString()
	resp, err := ai.Generate(Config.InitialMessages) // + JWT messages
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	claims := &Claims{
		ID: id,
		Messages: []openai.ChatCompletionMessage{{
			Role:    "system",
			Content: resp.OriginalMessage,
		}},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: nil,
		},
	}

	token, err := signJWT(claims)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": resp.Message,
		"options": resp.Options,
		"jwt":     token,
	})
}

type StoryResponse struct {
	JWT    string `json:"jwt"`
	Option int    `json:"option"`
}

func Respond(c *gin.Context) {
	var response StoryResponse
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

	claims.Messages = append(claims.Messages, openai.ChatCompletionMessage{
		Role:    "user",
		Content: strconv.Itoa(response.Option),
	})

	resp, err := ai.Generate(append(Config.InitialMessages, claims.Messages...))
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	claims.Messages = append(claims.Messages, openai.ChatCompletionMessage{
		Role:    "system",
		Content: resp.OriginalMessage,
	})

	token, err := signJWT(claims)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": resp.Message,
		"options": resp.Options,
		"jwt":     token,
	})
}
