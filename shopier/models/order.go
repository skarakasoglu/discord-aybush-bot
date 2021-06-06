package models

type Order struct {
	Email string `json:"email"`
	OrderId string `json:"orderid"`
	Currency string `json:"currency"`
	CurrencyString string
	Price string `json:"price"`
	Name string `json:"buyername"`
	Surname string `json:"buyersurname"`
	ProductCount string `json:"productcount"`
	ProductId string `json:"productid"`
	CustomerNote string `json:"customernote"`
	IsTest string `json:"istest"`
}
