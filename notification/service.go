package notification

import (
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/logger"
	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
)

type Notification interface {
	RenderContent() string
	RenderTitle() string
}

type JobNotificationService struct {
	// notificationCenter *NotificationCenter
	mongo *mongo.Service
}

func NewJobNotificationService(m *mongo.Service) *JobNotificationService {
	return &JobNotificationService{
		mongo: m,
	}
}

func (ns *JobNotificationService) Succeed(job *entity.Job) {
	session := ns.mongo.NewSession()
	defer session.Close()
	succeedNotification, err := NewSucceedJobNotification(session, job)
	if err != nil {
		logger.Error(err)
	}

	_ = succeedNotification

}

func (ns *JobNotificationService) Started(job *entity.Job) {
	session := ns.mongo.NewSession()
	defer session.Close()
	startedNotification, err := NewStartedJobNotification(session, job)
	if err != nil {
		logger.Error(err)
	}
	_ = startedNotification
}

func (ns *JobNotificationService) Canceled(job *entity.Job) {
	session := ns.mongo.NewSession()
	defer session.Close()
	canceledNotification, err := NewCanceledJobNotification(session, job)
	if err != nil {
		logger.Error(err)
	}
	_ = canceledNotification
}
