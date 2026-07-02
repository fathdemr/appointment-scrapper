package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type TelegramSetupHandler struct {
	logger *zap.Logger
}

func NewTelegramSetupHandler(logger *zap.Logger) *TelegramSetupHandler {
	return &TelegramSetupHandler{logger: logger}
}

type tgResponse[T any] struct {
	OK     bool   `json:"ok"`
	Result T      `json:"result"`
	Desc   string `json:"description"`
}

type tgBotInfo struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
}

type tgUpdate struct {
	Message *struct {
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
	} `json:"message"`
}

func callTelegram[T any](token, method string) (*tgResponse[T], error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/%s", token, method)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result tgResponse[T]
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// VerifyBotToken godoc
// @Summary  Bot token'ını doğrular, bot adını döner
// @Tags     telegram
// @Produce  json
// @Param    token query string true "Bot token"
// @Router   /telegram/verify-token [get]
func (h *TelegramSetupHandler) VerifyBotToken(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token gerekli"})
		return
	}

	res, err := callTelegram[tgBotInfo](token, "getMe")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	if !res.OK {
		c.JSON(http.StatusUnauthorized, gin.H{"error": res.Desc})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":         res.Result.ID,
		"name":       res.Result.FirstName,
		"username":   res.Result.Username,
		"bot_link":   "https://t.me/" + res.Result.Username,
	})
}

// DetectChatID godoc
// @Summary  Son güncellemeye bakarak chat_id'yi döner
// @Tags     telegram
// @Produce  json
// @Param    token query string true "Bot token"
// @Router   /telegram/detect-chat [get]
func (h *TelegramSetupHandler) DetectChatID(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token gerekli"})
		return
	}

	res, err := callTelegram[[]tgUpdate](token, "getUpdates?limit=10&allowed_updates=message")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	if !res.OK {
		c.JSON(http.StatusUnauthorized, gin.H{"error": res.Desc})
		return
	}
	if len(res.Result) == 0 {
		c.JSON(http.StatusOK, gin.H{"chat_id": nil, "found": false})
		return
	}
	// En son mesajdan chat_id al
	for i := len(res.Result) - 1; i >= 0; i-- {
		u := res.Result[i]
		if u.Message != nil {
			c.JSON(http.StatusOK, gin.H{
				"chat_id": fmt.Sprintf("%d", u.Message.Chat.ID),
				"found":   true,
			})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"chat_id": nil, "found": false})
}
