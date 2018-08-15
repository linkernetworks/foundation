package migration

import (
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/linkernetworks/foundation/logger"
	"github.com/linkernetworks/foundation/service/mongo"
	"github.com/linkernetworks/foundation/utils/timeutils"
	"gopkg.in/mgo.v2/bson"
)

const MigrationCollectionName = "migrations"

type Migration struct {
	Id           bson.ObjectId     `bson:"_id" json:"id"`
	ScriptId     string            `bson:"sid" json:"sid"`
	Version      string            `bson:"version" json:"version"`
	Description  string            `bson:"description" json:"description"`
	Function     MigrationFunction `bson:"-" json:"-"`
	Retry        int               `bson:"retry" json:"retry"`
	Running      bool              `bson:"running" json:"running"`
	DefinedAt    *time.Time        `bson:"definedAt" json:"definedAt"`
	StartedAt    *time.Time        `bson:"startedAt" json:"startedAt"`
	CompletedAt  *time.Time        `bson:"completedAt" json:"completedAt"`
	Completed    bool              `bson:"completed" json:"completed"`
	Errored      bool              `bson:"errored" json:"errored"`
	ErroredAt    *time.Time        `bson:"erroredAt" json:"erroredAt"`
	ErrorMessage string            `bson:"errorMessage" json:"errorMessage"`
}

func (m Migration) GetCollection() string {
	return MigrationCollectionName
}

type MigrationFunction func(ServiceContainer) error

func NewMigration(scriptId string, description string, f MigrationFunction) Migration {
	ret := strings.Split(scriptId, "-")
	if len(ret) < 2 {
		panic("Invalid migration script id. Please use {version}-{date} format")
	}
	version := ret[0]

	layout := "20060102"
	definedAt, err := time.Parse(layout, ret[1])

	if err != nil {
		panic(err)
	}

	return Migration{
		Version:     version,
		ScriptId:    scriptId,
		Description: description,
		DefinedAt:   &definedAt,
		Function:    f,
	}
}

func FindMigrationByScriptId(session *mongo.Session, sid string) (*Migration, error) {
	m := new(Migration)
	query := bson.M{"sid": sid}
	if err := session.FindOne(MigrationCollectionName, query, m); err != nil {
		return nil, err
	}

	return m, nil
}

func RunMigration(session *mongo.Session, ms Migration, as ServiceContainer) error {
	defer func() {
		if r := recover(); r != nil {
			var errMsg = fmt.Sprintf("handling panic script=%s error=%v: \n===STACKTRACE===\n%s\n===END OF STACKTRACE===", ms.ScriptId, r, debug.Stack())
			logger.Errorf("migration script panic: %s", errMsg)

			if err := session.Update(MigrationCollectionName, bson.M{"_id": ms.Id}, bson.M{
				"$set": bson.M{
					"running":      false,
					"errored":      r,
					"erroredAt":    time.Now(),
					"errorMessage": errMsg,
				},
			}); err != nil {
				logger.Errorf("failed to reset migration script state: %v", err)
			}
		}
	}()

	err := ms.Function(as)
	if err != nil {
		ms.Running = false
		ms.Errored = true
		ms.ErroredAt = timeutils.Now()
		ms.ErrorMessage = err.Error()

		if err := session.UpdateBy(MigrationCollectionName, "_id", ms.Id, ms); err != nil {
			logger.Error("Migration update failed", ms.ScriptId)
			return err
		}

		// when error occurred, we should stop the migration immediately
		return err
	}

	// Write this migration
	ms.CompletedAt = timeutils.Now()
	ms.Completed = true
	ms.Running = false

	if err := session.UpdateBy(MigrationCollectionName, "_id", ms.Id, ms); err != nil {
		logger.Error("Migration update failed", ms.ScriptId)
		return err
	}
	return nil
}

func Migrate(as ServiceContainer, dbVersion string, migrationScripts []Migration) ([]Migration, error) {
	session := as.GetMongo().NewSession()
	defer session.Close()

	var executedMigrations []Migration

	// Valid config dbVersion
	migrationScripts, err := filterRequiredMigrationScripts(dbVersion, migrationScripts)
	if err != nil {
		return executedMigrations, err
	}

	for _, ms := range migrationScripts {
		logger.Info("Checking migration: ", ms.ScriptId)

		record, _ := FindMigrationByScriptId(session, ms.ScriptId)
		if record != nil {
			logger.Info("Migration record exists: ", ms.ScriptId)
			record.Function = ms.Function
			ms = *record
		}

		if ms.Completed {
			logger.Infof("Migration %s was completed at %s", ms.ScriptId, ms.CompletedAt)
			continue
		}

		if ms.Running {
			logger.Infof("Found running migration %s, aborting", ms.ScriptId)
			return executedMigrations, errors.New("Found running migration")
		}

		if ms.Errored {
			logger.Info("Retrying Migration: ", ms.ScriptId)
			ms.Retry++
		}

		ms.StartedAt = timeutils.Now()
		ms.Completed = false
		ms.Running = true

		// If it's a new migration script
		if ms.Id == "" {
			ms.Id = bson.NewObjectId()
			// add the new migration script to the database
			if err := session.Insert(MigrationCollectionName, ms); err != nil {
				logger.Error("Migration insert failed", ms.ScriptId)
				return executedMigrations, err
			}
		} else {
			if err := session.UpdateBy(MigrationCollectionName, "_id", ms.Id, ms); err != nil {
				logger.Error("Migration update failed", ms.ScriptId)
				return executedMigrations, err
			}
		}

		// Execute the migration and check the error
		logger.Infof("Executing migration %s...", ms.ScriptId)
		err = RunMigration(session, ms, as)

		executedMigrations = append(executedMigrations, ms)

		if err != nil {
			logger.Errorf("Error occured during migration: %s. Migration aborted.", ms.ScriptId)
			logger.Errorf("Migration aborted.")
			return executedMigrations, err
		}

		logger.Infof("Migration %s completed.", ms.ScriptId)
		logger.Info("Continue to the next migration script...")
	}
	logger.Info("Migration Done")

	return executedMigrations, nil
}

func stripVPrefix(v string) string {
	if v[0] == 'v' {
		return v[1:]
	}
	return v
}

func filterRequiredMigrationScripts(cv string, mss []Migration) ([]Migration, error) {
	var scripts []Migration

	configVersion, err := semver.Make(stripVPrefix(cv))

	if err != nil {
		return scripts, err
	}

	for _, ms := range mss {
		mv, err := semver.Make(stripVPrefix(ms.Version))
		if err != nil {
			return scripts, err
		}
		ret := configVersion.Compare(mv)
		if ret == 0 || ret == 1 {
			scripts = append(scripts, ms)
		}
	}
	return scripts, nil
}
