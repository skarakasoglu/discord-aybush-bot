package payloads

type ChannelPayload struct{
	Data []ChannelInfo `json:"data"`
}

type ChannelInfo struct {
	BroadcasterId string `json:"broadcaster_id"`
	BroadcasterName string `json:"broadcaster_name"`
	GameName string `json:"game_name"`
	GameId string `json:"game_id"`
	BroadcasterLanguage string `json:"broadcaster_language"`
	Title string `json:"title"`
	Delay string `json:"delay"`
}
