package glaze

import "os"

// OpenLog will open a log file for appending, creating it if necessary
func OpenLog(logPath string) (*os.File, error) {
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		_, err := os.Create(logPath)
		if err != nil {
			return nil, err
		}
	}

	logFileHandle, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}

	return logFileHandle, nil
}
