package tgscrapper

const (
	maxNumberOfSymbolsInEllipsizeMessageTitle = 100 - 3

	maxNumberOfHoursAgoForOldestMessage float64 = 48.0

	maxNumberOfPages    = 10
	maxNumberOfMessages = maxNumberOfPages * 20 // 20 is a total number of messages per-page
)
