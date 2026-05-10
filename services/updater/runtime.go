package updater

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/pafthang/servicebase/core"
	"github.com/pafthang/servicebase/tools/archive"
	"github.com/spf13/cobra"
)

// CreateCommand builds the cobra update command bound to the service.
func (s *Service) CreateCommand() *cobra.Command {
	var withBackup bool

	command := &cobra.Command{
		Use:          "update",
		Short:        "Automatically updates the current app executable with the latest available version",
		SilenceUsage: true,
		RunE: func(command *cobra.Command, args []string) error {
			var needConfirm bool
			if isMaybeRunningInDocker() {
				needConfirm = true
				color.Yellow("NB! It seems that you are in a Docker container.")
				color.Yellow("The update command may not work as expected in this context because usually the version of the app is managed by the container image itself.")
			} else if isMaybeRunningInNixOS() {
				needConfirm = true
				color.Yellow("NB! It seems that you are in a NixOS.")
				color.Yellow("Due to the non-standard filesystem implementation of the environment, the update command may not work as expected.")
			}

			if needConfirm {
				confirm := false
				prompt := &survey.Confirm{
					Message: "Do you want to proceed with the update?",
				}
				survey.AskOne(prompt, &confirm)
				if !confirm {
					fmt.Println("The command has been cancelled.")
					return nil
				}
			}

			return s.Update(withBackup)
		},
	}

	command.PersistentFlags().BoolVar(
		&withBackup,
		"backup",
		true,
		"Creates a pb_data backup at the end of the update process",
	)

	return command
}

// Update updates the current executable to the latest release.
func (s *Service) Update(withBackup bool) error {
	color.Yellow("Fetching release information...")

	latest, err := fetchLatestRelease(
		s.config.Context,
		s.config.HttpClient,
		s.config.Owner,
		s.config.Repo,
	)
	if err != nil {
		return err
	}

	if compareVersions(strings.TrimPrefix(s.currentVersion, "v"), strings.TrimPrefix(latest.Tag, "v")) <= 0 {
		color.Green("You already have the latest version %s.", s.currentVersion)
		return nil
	}

	if compareVersions(strings.TrimPrefix(latest.Tag, "v"), "0.23.0") <= 0 {
		color.Green("%s contains breaking changes and cannot be updated directly from v0.22.x. Please check the releases CHANGELOG for more details.", latest.Tag)
		return nil
	}

	suffix := archiveSuffix(runtime.GOOS, runtime.GOARCH)
	if suffix == "" {
		return errors.New("unsupported platform")
	}

	asset, err := latest.findAssetBySuffix(suffix)
	if err != nil {
		return err
	}

	releaseDir := filepath.Join(s.App().DataDir(), core.LocalTempDirName)
	defer os.RemoveAll(releaseDir)

	color.Yellow("Downloading %s...", asset.Name)

	assetZip := filepath.Join(releaseDir, asset.Name)
	if err := downloadFile(s.config.Context, s.config.HttpClient, asset.DownloadUrl, assetZip); err != nil {
		return err
	}

	color.Yellow("Extracting %s...", asset.Name)

	extractDir := filepath.Join(releaseDir, "extracted_"+asset.Name)
	defer os.RemoveAll(extractDir)

	if err := archive.Extract(assetZip, extractDir); err != nil {
		return err
	}

	color.Yellow("Replacing the executable...")

	oldExec, err := os.Executable()
	if err != nil {
		return err
	}
	renamedOldExec := oldExec + ".old"
	defer os.Remove(renamedOldExec)

	newExec := filepath.Join(extractDir, s.config.ArchiveExecutable)
	if _, err := os.Stat(newExec); err != nil {
		newExec = newExec + ".exe"
		if _, fallbackErr := os.Stat(newExec); fallbackErr != nil {
			return fmt.Errorf("The executable in the extracted path is missing or it is inaccessible: %v, %v", err, fallbackErr)
		}
	}

	if err := os.Rename(oldExec, renamedOldExec); err != nil {
		return fmt.Errorf("Failed to rename the current executable: %w", err)
	}

	tryToRevertExecChanges := func() {
		if revertErr := os.Rename(renamedOldExec, oldExec); revertErr != nil {
			s.App().Logger().Debug(
				"Failed to revert executable",
				slog.String("old", renamedOldExec),
				slog.String("new", oldExec),
				slog.String("error", revertErr.Error()),
			)
		}
	}

	if err := os.Rename(newExec, oldExec); err != nil {
		tryToRevertExecChanges()
		return fmt.Errorf("Failed replacing the executable: %w", err)
	}

	if withBackup {
		color.Yellow("Creating pb_data backup...")

		backupName := fmt.Sprintf("@update_%s.zip", latest.Tag)
		if s.config.CreateBackup == nil {
			tryToRevertExecChanges()
			return fmt.Errorf("backup handler is not configured")
		}

		if err := s.config.CreateBackup(s.config.Context, backupName); err != nil {
			tryToRevertExecChanges()
			return err
		}
	}

	color.HiBlack("---")
	color.Green("Update completed successfully! You can start the executable as usual.")

	if latest.Body != "" {
		fmt.Print("\n")
		color.Cyan("Here is a list with some of the %s changes:", latest.Tag)
		releaseNotes := strings.TrimSpace(strings.Replace(latest.Body, "> _To update the prebuilt executable you can run `./"+s.config.ArchiveExecutable+" update`._", "", 1))
		color.Cyan(releaseNotes)
		fmt.Print("\n")
	}

	return nil
}

func fetchLatestRelease(
	ctx context.Context,
	client HttpClient,
	owner string,
	repo string,
) (*release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	rawBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("(%d) failed to fetch latest releases:\n%s", res.StatusCode, string(rawBody))
	}

	result := &release{}
	if err := json.Unmarshal(rawBody, result); err != nil {
		return nil, err
	}

	return result, nil
}

func downloadFile(
	ctx context.Context,
	client HttpClient,
	url string,
	destPath string,
) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return fmt.Errorf("(%d) failed to send download file request", res.StatusCode)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
		return err
	}

	dest, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dest.Close()

	if _, err := io.Copy(dest, res.Body); err != nil {
		return err
	}

	return nil
}

func archiveSuffix(goos, goarch string) string {
	switch goos {
	case "linux":
		switch goarch {
		case "amd64":
			return "_linux_amd64.zip"
		case "arm64":
			return "_linux_arm64.zip"
		case "arm":
			return "_linux_armv7.zip"
		}
	case "darwin":
		switch goarch {
		case "amd64":
			return "_darwin_amd64.zip"
		case "arm64":
			return "_darwin_arm64.zip"
		}
	case "windows":
		switch goarch {
		case "amd64":
			return "_windows_amd64.zip"
		case "arm64":
			return "_windows_arm64.zip"
		}
	}

	return ""
}

func compareVersions(a, b string) int {
	aSplit := strings.Split(a, ".")
	aTotal := len(aSplit)

	bSplit := strings.Split(b, ".")
	bTotal := len(bSplit)

	limit := aTotal
	if bTotal > aTotal {
		limit = bTotal
	}

	for i := 0; i < limit; i++ {
		var x, y int

		if i < aTotal {
			x, _ = strconv.Atoi(aSplit[i])
		}

		if i < bTotal {
			y, _ = strconv.Atoi(bSplit[i])
		}

		if x < y {
			return 1
		}

		if x > y {
			return -1
		}
	}

	return 0
}

func isMaybeRunningInDocker() bool {
	_, err := os.Stat("/.dockerenv")
	return err == nil
}

func isMaybeRunningInNixOS() bool {
	_, err := os.Stat("/etc/NIXOS")
	return err == nil
}
