package shopier

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/skarakasoglu/discord-aybush-bot/shopier/models"
	"log"
	"net/http"
	"strconv"
)

type currency uint8

const (
	currency_TL currency = iota
	currency_USD
	currency_EUR
)

var currencyTypes = map[currency]string{
	currency_TL: "TRY",
	currency_USD: "USD",
	currency_EUR: "EUR",
}

func (c currency) String() string{
	str, ok := currencyTypes[c]
	if !ok {
		return ""
	}

	return str
}

type apiv1 struct {
	username string
	key string

	processedOrders map[string]models.Order

	orderNotifierChan chan<- models.Order
}

func NewApiV1(username string, key string, orderNotifierChan chan<- models.Order) apiv1{
	return apiv1{
		username: username,
		key:      key,
		processedOrders: make(map[string]models.Order),
		orderNotifierChan: orderNotifierChan,
	}
}

func (a apiv1) onOrderNotification(ctx *gin.Context) {
	result := ctx.PostForm("res")
	hash := ctx.PostForm("hash")

	data, err := base64.StdEncoding.DecodeString(result)
	if err != nil {
		log.Printf("[ShopierAPI] Error on decoding base64 string: %v", err)
		ctx.String(http.StatusBadRequest, "")
		return
	}

	valid, expected := a.validateRequest(hash, result, a.username, a.key)
	if !valid {
		log.Printf("[ShopierAPI] Invalid payload signature received. The signature is %v, but it should have been %v", hash, expected)
		ctx.String(http.StatusUnauthorized, "")
		return
	}

	var order models.Order
	err = json.Unmarshal(data, &order)
	if err != nil {
		log.Printf("[ShopierAPI] Error on unmarshalling JSON: %v", err)
		ctx.String(http.StatusBadRequest, "")
		return
	}

	currencyInt, err := strconv.Atoi(order.Currency)
	if err != nil {
		log.Printf("[ShopierAPI] Error on parsing currency to int: %v", err)
	}
	order.CurrencyString = currency(currencyInt).String()

	_, ok := a.processedOrders[order.OrderId]
	if ok {
		log.Printf("[ShopierAPI] Duplicate order notification received. OrderId: %v", order.OrderId)
		ctx.String(http.StatusOK, "success")
		return
	}
	a.processedOrders[order.OrderId] = order
	a.orderNotifierChan <- order

	log.Printf("[ShopierAPI] %+v", order)
	ctx.String(http.StatusOK, "success")
}

func (a apiv1) validateRequest(signature, result string, username string, key string) (bool, string) {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(fmt.Sprintf("%v%v",result, username)))
	signatureShouldBe := fmt.Sprintf("%v", hex.EncodeToString(h.Sum(nil)))

	return signatureShouldBe == signature, signatureShouldBe
}