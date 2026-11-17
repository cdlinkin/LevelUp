package core_redis_cache

import (
	"fmt"
	"time"
)

func versionKey(userID int) string {
	return fmt.Sprintf("stats:ver:%d", userID)
}

func statisticsKey(
	userID int,
	version int64,
	from *time.Time,
	to *time.Time,
) string {
	return fmt.Sprintf(
		"stats:%d:%d:%s:%s",
		userID,
		version,
		dateKey(from),
		dateKey(to),
	)
}

func dateKey(date *time.Time) string {
	if date == nil {
		return "all"
	}

	return date.UTC().Format("2006-01-02")
}
