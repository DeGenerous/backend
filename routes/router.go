package routes

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
	"net/http"

	"backend/ai"

	. "backend/config"
)

type Claims struct {
	ID       string                         `json:"id"`
	Messages []openai.ChatCompletionMessage `json:"messages"`
	Step     int                            `json:"step"`
	jwt.RegisteredClaims
}

func signJWT(claims *Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(Config.Key)
	return tokenString, err
}

func Start(c *gin.Context) {
	id := uuid.NewString()
	resp, err := ai.Generate(Config.InitialMessages)
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
		Step: 1,
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

	claims.Step++

	var message string
	if claims.Step >= Config.MaxSteps {
		message = fmt.Sprintf("This is message number %d, you should finish the story now. I choose option %d.", claims.Step, response.Option)
	} else {
		message = fmt.Sprintf("This is message number %d. I choose option %d.", claims.Step, response.Option)
	}

	claims.Messages = append(claims.Messages, openai.ChatCompletionMessage{
		Role:    "user",
		Content: message,
	})

	messages := append([]openai.ChatCompletionMessage{}, Config.InitialMessages...)
	messages = append(messages, claims.Messages...)
	resp, err := ai.Generate(messages)
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
		"end":     resp.End,
		"summary": resp.Summary,
		"jwt":     token,
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
	if err != nil || !tkn.Valid {
		c.String(http.StatusUnauthorized, err.Error())
		return
	}

	resp, err := ai.Image(claims.Messages[len(claims.Messages)-1].Content)
	if err != nil || !tkn.Valid {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.String(http.StatusOK, resp)
}
