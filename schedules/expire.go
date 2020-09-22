package schedules

import (
	"log"

	"u9k/db"
	"u9k/storage"
)

func expireFiles() {
	files, err := db.GetExpiredFiles()
	if err != nil {
		log.Printf("Failed to get expired files: %s\n", err)
		return
	}

	var num int64
	for _, file := range files {
		log.Printf("Expiring file %s (created on %s)\n", file.Id, file.CreateTimestamp)
		err := db.DeleteFile(file.Id)
		if err != nil {
			log.Printf("Failed to delete file %s from DB: %s\n", file.Id, err)
			continue
		}
		err = storage.DeleteFile(file.Id)
		if err != nil {
			log.Printf("Failed to delete file %s from storage: %s\n", file.Id, err)
		}
		num++
	}

	log.Printf("Successfully expired %d files\n", num)
}
