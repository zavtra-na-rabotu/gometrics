package storage

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

func ConfigureStorage(memStorage *MemStorage, fileStoragePath string, restore bool, storeInterval int) error {
	if fileStoragePath == "" {
		return nil
	}

	// Try to restore metrics from file
	if restore {
		err := RestoreMetricsFromFile(memStorage, fileStoragePath)
		if err != nil {
			zap.L().Error("Error restoring metrics", zap.Error(err))
		}
	}

	if storeInterval == 0 {
		writer, err := NewWriter(fileStoragePath)
		if err != nil {
			return fmt.Errorf("could not create file writer: %w", err)
		}

		memStorage.SetSyncMode(true)
		memStorage.SetFileWriter(writer)

		return nil
	}

	// Save metrics to file every {config.storeInterval} seconds
	storeToFileTicker := time.NewTicker(time.Duration(storeInterval) * time.Second)
	go func() {
		for {
			<-storeToFileTicker.C
			err := WriteMetricsToFile(memStorage, fileStoragePath)
			if err != nil {
				zap.L().Error("Error storing metrics", zap.Error(err))
			}
		}
	}()

	return nil
}
