package utils

func BoolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func TimeframeToMilliseconds(tf string) int64 {
	switch tf {
	case "1m":
		return 60 * 1000
	case "15m":
		return 900 * 1000
	case "30m":
		return 1800 * 1000
	case "1h":
		return 3600 * 1000
	case "30m":
		return 1800 * 1000
	case "1d":
		return 86400 * 1000
	}

	return 0
}
