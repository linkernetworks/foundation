package messenger

import (
	"bitbucket.org/linkernetworks/aurora/src/logger"

	_ "bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/service/mongo"

	twilio "github.com/carlosdp/twiliogo"
	mailgun "gopkg.in/mailgun/mailgun-go.v1"
)

type MessageSender interface {
	Send() error
	NewService()
}

type Mailgun struct {
	domain       string
	ApiKey       string
	publicApiKey string
}

type Twilio struct {
	accountSid string
	authToken  string
}

func (mg *Mailgun) NewService(m *mongo.Service) {
	context := m.NewSession()
	defer context.Close()
	// TODO Read config from mongo

	mg.domain = "linkernetworks.com"
	mg.ApiKey = "blahblah"
	mg.publicApiKey = "blahblah"
}

func (twlo *Twilio) NewService(m *mongo.Service) {
	context := m.NewSession()
	defer context.Close()
	// TODO Read config from mongo

	twlo.accountSid = "xxxxx"
	twlo.authToken = "ddddd"
}

func (mg *Mailgun) Send(msg Message) error {
	client := mailgun.NewMailgun(mg.domain, mg.ApiKey, mg.publicApiKey)
	message := client.NewMessage(msg.From(), msg.Title(), msg.Content(), msg.To())
	resp, id, err := client.Send(message)
	if err != nil {
		return err
	}
	logger.Infof("ID: %s Resp: %s\n", id, resp)
	return nil
}

func (twlo *Twilio) Send(msg Message) error {
	client := twilio.NewClient(twlo.accountSid, twlo.authToken)
	message, err := twilio.NewMessage(client, msg.Title(), msg.From(), twilio.Body(msg.Content()))
	if err != nil {
		return err
	}
	logger.Info("no err:", message.Status)
	return nil
}

// func main() {
//
// 	e := &Email{}
//
// 	a := &Mailgun{}
// 	a.NewService()
//
// 	a.Send(e)
//
// }
