package utils

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/runtipi/cli/internal/config"
)

type AppBackup struct {
	Name      string
	Size      string
	CreatedAt string
}

func GetAppBackups(urn string) ([]AppBackup, error) {
	uparts := strings.SplitN(urn, ":", 2)
	if len(uparts) != 2 {
		return nil, errors.New("error invalid URN")
	}
	backupsDir := path.Join(config.RootFolder, "backups", uparts[1], uparts[0])
	files, err := os.ReadDir(backupsDir)
	if err != nil {
		return nil, err
	}
	var backups []AppBackup
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".tar.gz") {
			continue
		}
		bparts := strings.SplitN(strings.TrimSuffix(file.Name(), ".tar.gz"), "-", 2)
		if len(bparts) < 2 || bparts[0] != urn {
			continue
		}
		sTimestamp := bparts[1]
		iTimestamp, err := strconv.ParseInt(sTimestamp, 10, 64)
		if err != nil {
			continue
		}
		tz, err := time.LoadLocation(os.Getenv("TZ"))
		if err != nil {
			continue
		}
		createdAt := time.UnixMilli(iTimestamp).In(tz).String()
		info, err := file.Info()
		if err != nil {
			continue
		}
		size := info.Size()
		sizeStr := fmt.Sprintf("%.2f MB", float64(size)/1024/1024)
		backups = append(backups, AppBackup{
			Name:      file.Name(),
			Size:      sizeStr,
			CreatedAt: createdAt,
		})
	}
	return backups, nil
}
