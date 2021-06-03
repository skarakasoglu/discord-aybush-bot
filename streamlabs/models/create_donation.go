package models

type CreateDonation struct {
	Name string
	Message string
	Identifier string
	Amount float64
	CreatedAt int64
	Currency string
	SkipAlert string
}

type CreateDonationResponse struct{
	DonationId int64 `json:"donation_id"`
}
