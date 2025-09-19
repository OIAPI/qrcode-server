package utils

// utils/qrcode.go 中添加导入
// utils/qrcode.go 导入部分（关键修改：给 gozxing/qrcode 加别名）
import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io" // 新增：解决 io 未定义

	"github.com/makiuchi-d/gozxing"
	gozxingQR "github.com/makiuchi-d/gozxing/qrcode" // 别名：gozxingQR，避免冲突
	"github.com/skip2/go-qrcode"				// 保留原包名 qrcode（生成二维码用）
	"qrcode-server/config"  // 替换为你的本地模块名
)



// QRGenerateParam 二维码生成参数
type QRGenerateParam struct {
	Content string // 二维码内容（必传）
	Size		int		// 尺寸（像素）
	Level	 string // 纠错级别（L/M/Q/H）
	Type		string // 图片格式（png/jpeg）
}

// GenerateQR 生成二维码图片（返回image.Image实例）
func GenerateQR(param QRGenerateParam) (image.Image, error) {
	// 1. 解析纠错级别
	qrLevel, err := parseQRErrorLevel(param.Level)
	if err != nil {
		return nil, err
	}

	// 2. 创建二维码基础实例
	qrInst, err := qrcode.New(param.Content, qrLevel)
	if err != nil {
		return nil, fmt.Errorf("创建二维码失败: %w", err)
	}

	// 3. 自定义样式（增强多样性）
	qrInst.BackgroundColor = color.White				 // 背景色
	qrInst.ForegroundColor = color.RGBA{0,0,0,255} // 码点色（黑色，可自定义其他色）
	qrInst.DisableBorder = false								 // 显示边框（false=显示，true=隐藏）

	// 4. 生成图片
	return qrInst.Image(param.Size), nil
}

// EncodeQR 按指定格式编码二维码图片（写入io.Writer，用于HTTP响应）
func EncodeQR(img image.Image, imgType string, w io.Writer) error {
	switch imgType {
	case "jpeg":
		return jpeg.Encode(w, img, &jpeg.Options{Quality: 90}) // 90%质量，平衡清晰度和大小
	default: // png（默认格式，兼容性最好）
		return png.Encode(w, img)
	}
}

func DecodeQR(img image.Image) (string, error) {
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return "", fmt.Errorf("转换为位图失败: %w", err)
	}
	hints := make(map[gozxing.DecodeHintType]interface{})
	hints[gozxing.DecodeHintType_CHARACTER_SET] = "UTF-8"
	reader := gozxingQR.NewQRCodeReader()
	result, err := reader.Decode(bmp, hints)
	if err != nil {
		return "", fmt.Errorf("找不到二维码: %w", err)
	}
	return result.GetText(), nil
}

// parseQRErrorLevel 转换纠错级别参数（从配置默认值+请求参数）
func parseQRErrorLevel(levelStr string) (qrcode.RecoveryLevel, error) {
	cfg := config.Get()
	defaultLevel := cfg.QRCode.DefaultLevel

	// 无参数时用配置默认值
	if levelStr == "" {
		levelStr = defaultLevel
	}

	// 转换为qrcode库的RecoveryLevel
	switch levelStr {
	case "L":
		return qrcode.Low, nil		// 7%纠错（适合内容简单场景）
	case "M":
		return qrcode.Medium, nil // 15%纠错（默认，平衡性能和容错）
	case "Q":
		return qrcode.High, nil	 // 25%纠错（适合复杂环境）
	case "H":
		return qrcode.Highest, nil// 30%纠错（容错最高，二维码密度大）
	default:
		return 0, fmt.Errorf("level 必须是 L/M/Q/H")
	}
}
