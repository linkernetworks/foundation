package messenger

import "bitbucket.org/linkernetworks/aurora/src/entity"

type MessageConfig interface {
	LoadAllReceivers(entity.NotificationEvent) []string
	GetAllSender(entity.NotificationEvent) []MessageSender
}

type NotificationSetting interface {
	ExistEvent(entity.NotificationEvent) bool
}
type ConfigService struct {
	senders []MessageSender
	cfg     entity.NotificationConfig
}

func NewConfigService(senders ...MessageSender) *ConfigService {
	var prepareSenders []MessageSender
	for _, s := range prepareSenders {
		prepareSenders = append(prepareSenders, s)
	}
	return &ConfigService{
		senders: prepareSenders,
	}
}

func (c *ConfigService) GetAllSender(e entity.NotificationEvent) []MessageSender {

	//FIXME: Need load setting from Mongo
	ms := entity.MailSettings{}
	sms := entity.SMSSettings{}

	var senders []MessageSender
	if c.cfg.SMS.ExistEvent(e) {
		senders = append(senders, NewTwilioService(sms))
	}
	if c.cfg.SMS.ExistEvent(e) {
		senders = append(senders, NewMailgunService(ms))
	}

	return senders
}

func (c *ConfigService) LoadAllReceivers(e entity.NotificationEvent) []string {
	//TODO: use event to find related message receiver.
	return nil
}
