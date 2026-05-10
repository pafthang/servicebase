package migrate

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/AlecAivazis/survey/v2"
	collectionmodels "github.com/pafthang/servicebase/services/collection/models"
	"github.com/pafthang/servicebase/services/migrate/registry"
	"github.com/pafthang/servicebase/tools/inflector"
	migratetool "github.com/pafthang/servicebase/tools/migrate"
	"github.com/pocketbase/dbx"
	"github.com/spf13/cobra"
)

// CreateCommand builds the cobra migrate command bound to the service.
func (s *Service) CreateCommand() *cobra.Command {
	const cmdDesc = `Supported arguments are:
- up            - runs all available migrations
- down [number] - reverts the last [number] applied migrations
- create name   - creates new blank migration template file
- collections   - creates new migration file with snapshot of the local collections configuration
- history-sync  - ensures that the _migrations history table doesn't have references to deleted migration files
`

	command := &cobra.Command{
		Use:          "migrate",
		Short:        "Executes app DB migration scripts",
		Long:         cmdDesc,
		ValidArgs:    []string{"up", "down", "create", "collections"},
		SilenceUsage: true,
		RunE: func(command *cobra.Command, args []string) error {
			cmd := ""
			if len(args) > 0 {
				cmd = args[0]
			}

			switch cmd {
			case "create":
				if _, err := s.migrateCreateHandler("", args[1:], true); err != nil {
					return err
				}
			case "collections":
				if _, err := s.migrateCollectionsHandler(args[1:], true); err != nil {
					return err
				}
			default:
				db, ok := s.App().Dao().DB().(*dbx.DB)
				if !ok || db == nil {
					return fmt.Errorf("Failed to resolve app db")
				}

				runner, err := migratetool.NewRunner(db, registry.AppMigrations)
				if err != nil {
					return err
				}

				if err := runner.Run(args...); err != nil {
					return err
				}
			}

			return nil
		},
	}

	return command
}

// CreateMigration creates a new migration file using the blank template.
func (s *Service) CreateMigration(name string, interactive bool) (string, error) {
	return s.migrateCreateHandler("", []string{name}, interactive)
}

// CreateCollectionsSnapshot creates a snapshot migration for current collections.
func (s *Service) CreateCollectionsSnapshot(interactive bool, extraArgs ...string) (string, error) {
	return s.migrateCollectionsHandler(extraArgs, interactive)
}

func (s *Service) migrateCreateHandler(template string, args []string, interactive bool) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("Missing migration file name")
	}

	name := args[0]
	dir := s.config.Dir
	filename := fmt.Sprintf("%d_%s.%s", time.Now().Unix(), inflector.Snakecase(name), s.templateLang())
	resultFilePath := path.Join(dir, filename)

	if interactive {
		confirm := false
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Do you really want to create migration %q?", resultFilePath),
		}
		survey.AskOne(prompt, &confirm)
		if !confirm {
			fmt.Println("The command has been cancelled")
			return "", nil
		}
	}

	if template == "" {
		var templateErr error
		template, templateErr = s.blankTemplate()
		if templateErr != nil {
			return "", fmt.Errorf("Failed to resolve create template: %v\n", templateErr)
		}
	}

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", err
	}

	if err := os.WriteFile(resultFilePath, []byte(template), 0644); err != nil {
		return "", fmt.Errorf("Failed to save migration file %q: %v\n", resultFilePath, err)
	}

	if interactive {
		fmt.Printf("Successfully created file %q\n", resultFilePath)
	}

	return filename, nil
}

func (s *Service) templateLang() string {
	return TemplateLangGo
}

func (s *Service) blankTemplate() (string, error) {
	return s.goBlankTemplate()
}

func (s *Service) snapshotTemplate(collections []*collectionmodels.Collection) (string, error) {
	return s.goSnapshotTemplate(collections)
}

func (s *Service) diffTemplate(new, old *collectionmodels.Collection) (string, error) {
	return s.goDiffTemplate(new, old)
}

func (s *Service) migrateCollectionsHandler(args []string, interactive bool) (string, error) {
	createArgs := []string{"collections_snapshot"}
	createArgs = append(createArgs, args...)

	collections := []*collectionmodels.Collection{}
	if err := s.App().Dao().CollectionQuery().OrderBy("created ASC").All(&collections); err != nil {
		return "", fmt.Errorf("Failed to fetch migrations list: %v", err)
	}

	template, templateErr := s.snapshotTemplate(collections)
	if templateErr != nil {
		return "", fmt.Errorf("Failed to resolve template: %v", templateErr)
	}

	return s.migrateCreateHandler(template, createArgs, interactive)
}
