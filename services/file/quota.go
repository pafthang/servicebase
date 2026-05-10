package file

import (
	"fmt"

	"github.com/pafthang/servicebase/core"
	filemodels "github.com/pafthang/servicebase/services/file/models"
)

const DefaultQuotaBytes = 1073741824

func GetQuotaInfo(app core.App, userId string) (quotaBytes int64, usedBytes int64, err error) {
	quota, err := findUserStorageQuota(app, userId)

	if err != nil {
		return DefaultQuotaBytes, 0, nil
	}

	quotaBytes = quota.QuotaBytes
	if quotaBytes == 0 {
		quotaBytes = DefaultQuotaBytes
	}

	usedBytes = quota.UsedBytes
	return quotaBytes, usedBytes, nil
}

func CheckQuotaAvailable(app core.App, userId string, fileSize int64) (bool, error) {
	quotaBytes, usedBytes, err := GetQuotaInfo(app, userId)
	if err != nil {
		return false, err
	}

	availableBytes := quotaBytes - usedBytes
	return availableBytes >= fileSize, nil
}

func UpdateQuotaUsage(app core.App, userId string, deltaBytes int64) error {
	quota, err := findUserStorageQuota(app, userId)

	if err != nil {
		newQuota := &filemodels.UserStorageQuota{
			User:       userId,
			QuotaBytes: DefaultQuotaBytes,
			UsedBytes:  deltaBytes,
		}

		if err := saveUserStorageQuota(app, newQuota); err != nil {
			return fmt.Errorf("failed to create quota record: %w", err)
		}
		return nil
	}

	newUsed := quota.UsedBytes + deltaBytes
	if newUsed < 0 {
		newUsed = 0
	}

	quota.UsedBytes = newUsed
	if err := saveUserStorageQuota(app, quota); err != nil {
		return fmt.Errorf("failed to update quota: %w", err)
	}

	return nil
}
