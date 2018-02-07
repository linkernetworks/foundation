package messenger

type MessageConfig interface {
	LoadAllReceivers(MessageEvent) []string
	GetAllSender(MessageEvent) []MessageSender
}
