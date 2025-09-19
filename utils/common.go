package utils

import (
	"fmt"
	"strconv"

	"qrcode-server/config"
)

// CheckQRType 校验二维码格式是否在支持列表中
func CheckQRType(imgType string) bool {
	cfg := config.Get()
	for _, t := range cfg.QRCode.SupportTypes {
		if t == imgType {
			return true
		}
	}
	return false
}

// ParseQRSize 解析二维码尺寸参数（处理默认值和范围校验）
func ParseQRSize(sizeStr string) (int, error) {
	cfg := config.Get()
	defaultSize := cfg.QRCode.DefaultSize

	// 无参数时返回默认值
	if sizeStr == "" {
		return defaultSize, nil
	}

	// 字符串转整数
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		return 0, fmt.Errorf("大小必须是整数")
	}

	// 范围校验（100-2000像素，避免过小/过大）
	if size < 100 || size > 2000 {
		return 0, fmt.Errorf("尺寸必须在100到2000像素之间")
	}
	return size, nil
}

// GetQRSupportTypes 获取支持的二维码格式列表（字符串形式，用于错误提示）
func GetQRSupportTypes() string {
	cfg := config.Get()
	typesStr := ""
	for i, t := range cfg.QRCode.SupportTypes {
		if i > 0 {
			typesStr += ","
		}
		typesStr += t
	}
	return typesStr
}
