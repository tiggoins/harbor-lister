package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
)

type Config struct {
	HarborURL   string
	Username    string
	Password    string
	OutputFile  string
	InsecureSSL bool
}

func ParseFlags() *Config {
	config := &Config{}

	pflag.StringVarP(&config.HarborURL, "url", "u", "", "Harbor 服务器地址 (必填)")
	pflag.StringVarP(&config.Username, "username", "U", "", "Harbor 用户名 (必填)")
	pflag.StringVarP(&config.Password, "password", "P", "", "Harbor 密码 (必填)")
	pflag.StringVarP(&config.OutputFile, "output", "o", "harbor_images.xlsx", "输出的 Excel 文件路径")
	pflag.BoolVar(&config.InsecureSSL, "insecure-ssl", true, "是否跳过 SSL 证书验证（默认：是）")

	pflag.Parse()

	// Validate required flags
	if config.HarborURL == "" || config.Username == "" || config.Password == "" {
		fmt.Fprintln(os.Stderr, "错误: url, username 和 password 参数为必填项")
		pflag.Usage()
		os.Exit(1)
	}

	// Process Harbor URL
	config.HarborURL = normalizeHarborURL(config.HarborURL)

	// Ensure the output directory exists
	outputDir := filepath.Dir(config.OutputFile)
	if outputDir != "." {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "创建输出目录失败: %v\n", err)
			os.Exit(1)
		}
	}

	return config
}

func normalizeHarborURL(url string) string {
	url = strings.TrimSpace(url) // 去掉首尾空格
	url = strings.TrimRight(url, "/")

	if !strings.HasPrefix(url, "https://") {
		if strings.HasPrefix(url, "http://") {
			url = "https://" + strings.TrimPrefix(url, "http://")
		} else {
			url = "https://" + url
		}
	}

	if !strings.HasSuffix(url, "/api/v2.0") {
		url += "/api/v2.0"
	}

	return url
}
