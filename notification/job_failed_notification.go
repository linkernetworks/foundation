package notification

import (
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/logger"
	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"bytes"

	"gopkg.in/mgo.v2/bson"
	"log"
	"text/template"
	"time"
)

type FailedJobNotification struct {
	Job       *entity.Job
	Send      *entity.User
	CreatedAt time.Time
}

func NewFailedJobNotification(session *mongo.Session, job *entity.Job) (*FailedJobNotification, error) {
	uid := job.CreatedBy

	user := entity.User{}
	query := bson.M{"_id": uid}
	if err := session.C(entity.UserCollectionName).Find(query).One(&user); err != nil {
		logger.Error(err)
		return nil, err
	}

	return &FailedJobNotification{
		Job:       job,
		Send:      &user,
		CreatedAt: time.Now(),
	}, nil
}

func (fn *FailedJobNotification) RenderContent() (string, error) {
	// Define a template.
	const letter = `
Dear {{.Name}},
 
Your job {{.JobId}} {{.Link}} has failed. 
Please login to aiForge for more details.

Note: This is automatic message by aiForge system. Please do not reply.

Linker Networks Team
`
	// Prepare some data to insert into the template.
	type Recipient struct {
		Name  string
		JobId string
		Link  string
	}
	// FIXME
	fullName := fn.Send.FirstName + fn.Send.LastName
	link := "https://www.google.com"
	var recipient = Recipient{
		Name:  fullName,
		JobId: fn.Job.ID.Hex(),
		Link:  link,
	}

	// Create a new template and parse the letter into it.
	t := template.Must(template.New("letter").Parse(letter))

	var buf bytes.Buffer

	// Execute the template for each recipient.
	err := t.Execute(&buf, recipient)
	if err != nil {
		log.Printf("Render Content Failed when executing template:", err)
		return "", err
	}
	return buf.String(), nil
}

func (fn *FailedJobNotification) RenderTitle() (string, error) {
	const subjectTml = `[aiForge] Job {{.JobId}} has failed`
	type Content struct {
		JobId string
	}

	var c = Content{
		JobId: fn.Job.ID.Hex(),
	}

	t := template.Must(template.New("subjectTml").Parse(subjectTml))
	var buf bytes.Buffer

	err := t.Execute(&buf, c)
	if err != nil {
		log.Printf("Render Content Failed when executing template:", err)
		return "", err
	}
	return buf.String(), nil
}
