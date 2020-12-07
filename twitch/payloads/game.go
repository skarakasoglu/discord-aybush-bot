package payloads

type GamePayload struct{
	Data []Game `json:"data"`
}

type Game struct{
	ID string `json:"id"`
	Name string `json:"name"`
	BoxArtURL string `json:"box_art_url"`
}
