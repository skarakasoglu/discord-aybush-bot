package models

type GetDonationResponse struct{
	Data []GetDonation `json:"data"`
}

type GetDonation struct{
	Id int `json:"donate_id"`
	CreatedAt uint64 `json:"created_at"`
	Currency string `json:"currency"`
	Amount float64 `json:"amount"`
	Name string `json:"name"`
	Message string `json:"message"`
	Email string `json:"email"`
}
