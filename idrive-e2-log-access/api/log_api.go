package api

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
)

type FileInfo struct {
	Name    string    `json:"name"`
	ModTime time.Time `json:"mod_time"`
}

type LogApi struct {
	appLogDirMap map[string]string
}

func NewLogApi(appLogDirMap map[string]string) *LogApi {
	return &LogApi{appLogDirMap: appLogDirMap}
}

func (la *LogApi) ListLogFiles(c *gin.Context) {
	app := c.Query("app")
	//logDir := "/var/log/letsencrypt"
	logDir, ok := la.appLogDirMap[app]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "app not found"})
		return
	}
	files, err := os.ReadDir(logDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var logFiles []FileInfo
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".log" {
			info, err := file.Info()
			if err != nil {
				continue
			}
			logFiles = append(logFiles, FileInfo{Name: file.Name(), ModTime: info.ModTime()})
		}
	}

	// Sort by last modified date and file name
	sort.Slice(logFiles, func(i, j int) bool {
		if logFiles[i].ModTime.Equal(logFiles[j].ModTime) {
			return logFiles[i].Name < logFiles[j].Name
		}
		return logFiles[i].ModTime.After(logFiles[j].ModTime)
	})

	c.JSON(http.StatusOK, logFiles)
}
func (la *LogApi) DownloadLogFile(c *gin.Context) {
	app := c.Query("app")
	logFileName := c.Query("log-file")
	logDir, ok := la.appLogDirMap[app]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "app not found"})
		return
	}
	logFilePath := filepath.Join(logDir, logFileName)
	// date := c.Param("date")
	// logFileName := fmt.Sprintf("cert-manager-%s.log", date)
	// logFilePath := filepath.Join("/var/log/letsencrypt", logFileName)

	c.FileAttachment(logFilePath, logFileName)
}
