package log

import (
	"fmt"
	"time"

	"github.com/pafthang/servicebase/core"
	logmodels "github.com/pafthang/servicebase/services/log/models"
	"github.com/pocketbase/dbx"
)

func CleanupLogs(app core.App) error {
	if app == nil {
		return fmt.Errorf("app is required")
	}

	var projects []*logmodels.LoggingProject
	if err := app.Dao().ModelQuery(&logmodels.LoggingProject{}).All(&projects); err != nil {
		return fmt.Errorf("failed to list logging projects: %w", err)
	}

	for _, project := range projects {
		if project.Retention <= 0 {
			continue
		}

		cutoff := time.Now().AddDate(0, 0, -project.Retention)
		for {
			result, err := app.Dao().NonconcurrentDB().NewQuery(
				"DELETE FROM {{logs}} WHERE [[id]] IN (" +
					"SELECT [[id]] FROM {{logs}} WHERE [[project]] = {:project} AND [[timestamp]] <= {:cutoff} LIMIT 1000" +
					")",
			).Bind(dbx.Params{
				"project": project.Id,
				"cutoff":  cutoff,
			}).Execute()
			if err != nil {
				return fmt.Errorf("failed to delete old logs for project %s: %w", project.Id, err)
			}

			rowsAffected, err := result.RowsAffected()
			if err != nil {
				return fmt.Errorf("failed to inspect deleted rows for project %s: %w", project.Id, err)
			}

			if rowsAffected == 0 {
				break
			}

			time.Sleep(50 * time.Millisecond)
		}
	}

	return nil
}
