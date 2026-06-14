// 負責 api 上傳檔案處裡
package api

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

func SaveUploadedZipFile(c *gin.Context, formKey string, dst string) (int, error) {
	file, err := c.FormFile(formKey)
	if err != nil {
		return http.StatusBadRequest, errors.New("failed to get uploaded file")
	}

	if filepath.Ext(strings.ToLower(file.Filename)) != ".zip" {
		return http.StatusBadRequest, errors.New("only zip files are allowed")
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return http.StatusInternalServerError, errors.New("failed to create directory")
	}

	if err := c.SaveUploadedFile(file, dst); err != nil {
		return http.StatusInternalServerError, errors.New("failed to save uploaded file")
	}

	return http.StatusOK, nil
}

type LimitsConfig struct{
	TotalTime	int	`yaml:"totalTime"`
  	CpuTime		int	`yaml:"cpuTime"`
	Memory		int	`yaml:"memory"`

}
type YAMLConfig struct {
	Title			string			`yaml:"title"`
	LimitsConfig	LimitsConfig	`yaml:"limits"`
}

func ParseYAML(yamlPath string) (string, int, error){
	yamlFile, err := os.ReadFile(yamlPath)
	if err != nil {
		return "", 0, errors.New("Error reading YAML file")
	}

	var config YAMLConfig
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		return "", 0, errors.New("Error parsing YAML")
	}
	
	return config.Title, config.LimitsConfig.TotalTime, nil
}