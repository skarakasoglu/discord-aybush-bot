package twitch

type ChatBot struct {
	username string
	token string
	client *ApiClient
}

func NewChatBot(username string, token string, client *ApiClient) *ChatBot {
	return &ChatBot{
		username: username,
		token:    token,
		client: client,
	}
}

func (cb *ChatBot) Start() {

}

func (cb *ChatBot) onCommandReceived() {

}

func (cb *ChatBot) onFollowage() {

}