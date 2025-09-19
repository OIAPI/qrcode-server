package router

import (
	"image"
	"log/slog"
	"net/http"

	"qrcode-server/utils"

	"github.com/gin-gonic/gin"
)

// router/router.go 中的 InitRouter 函数（修改中间件部分）
func InitRouter(log *slog.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// 关键修改：Writer → Output（适配 Gin v1.9.0+）
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Output: &slogWriter{log: log}, // 字段名改为 Output
	}), gin.Recovery())

	registerQRCodeRoutes(r, log)
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"title":   "页面不存在",
			"message": "您访问的路径不存在，请检查 URL 是否正确",
		})
	})
	return r
}

// 以下 slogWriter 结构体和 Write 方法不变，无需修改
type slogWriter struct {
	log *slog.Logger
}

func (w *slogWriter) Write(p []byte) (n int, err error) {
	w.log.Info("gin 请求日志", "detail", string(p))
	return len(p), nil
}

// registerQRCodeRoutes 注册二维码相关路由（路由分组）
func registerQRCodeRoutes(r *gin.Engine, log *slog.Logger) {
	// 路由前缀：/api/qrcode
	qrGroup := r.Group("/api/qrcode")
	{
		qrGroup.GET("/generate", handleGenerateQR(log)) // 生成二维码
		qrGroup.POST("/decode", handleDecodeQR(log))    // 识别二维码
	}
}

// handleGenerateQR 生成二维码处理器（原main.go迁移）
func handleGenerateQR(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 解析参数
		param := utils.QRGenerateParam{
			Content: c.Query("content"),
			Level:   c.Query("level"),
			Type:    c.Query("type"),
		}

		// 2. 参数校验
		if param.Content == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "需要参数：content"})
			return
		}
		// 解析尺寸
		size, err := utils.ParseQRSize(c.Query("size"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": -2, "message": err.Error()})
			return
		}
		param.Size = size
		// 校验格式
		if param.Type == "" {
			param.Type = "png"
		}
		if !utils.CheckQRType(param.Type) {
			supportTypes := utils.GetQRSupportTypes()
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    -3,
				"message": "不支持的格式, 只支持: " + supportTypes,
			})
			return
		}

		// 3. 生成+响应图片
		img, err := utils.GenerateQR(param)
		if err != nil {
			log.Error("生成QRCode失败", "content", param.Content, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": -4, "message": "生成失败"})
			return
		}
		c.Header("Content-Type", "image/"+param.Type)
		if err := utils.EncodeQR(img, param.Type, c.Writer); err != nil {
			log.Error("编码二维码失败", "type", param.Type, "error", err)
		}
	}
}

// router/router.go 中的 handleDecodeQR 函数（确保只接收 2 个返回值）
func handleDecodeQR(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var img image.Image
		var err error

		// 1. 优先处理 URL 参数（支持通过 URL 下载图片识别）
		imgURL := c.Query("url")
		if imgURL != "" {
			// 1.1 下载 URL 对应的图片
			resp, err := http.Get(imgURL)
			if err != nil {
				log.Error("从URL下载图像失败", "url", imgURL, "error", err)
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    -6,
					"message": "无法从URL下载图像（检查URL有效性）",
				})
				return
			}
			defer resp.Body.Close()

			// 1.2 校验 HTTP 响应状态（仅处理 200 OK 的图片）
			if resp.StatusCode != http.StatusOK {
				log.Error("图像URL的无效HTTP状态", "url", imgURL, "status", resp.Status)
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    -7,
					"message": "图像URL返回无效状态: " + resp.Status,
				})
				return
			}

			// 1.3 解码下载的图片（与上传文件的解码逻辑一致）
			img, _, err = image.Decode(resp.Body)
			if err != nil {
				log.Error("从URL解码图像失败", "url", imgURL, "error", err)
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    -8,
					"message": "URL不是有效的图像（支持PNG/JPEG）",
				})
				return
			}
		} else {
			// 2. 无 URL 参数，走原有的“上传文件”逻辑
			fileHeader, err := c.FormFile("file")
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    -9,
					"message": "需要“文件”（上传）或“ URL”（图像链接）",
				})
				return
			}

			// 2.1 打开上传的文件流
			file, err := fileHeader.Open()
			if err != nil {
				log.Error("打开上传的文件失败", "error", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    -10,
					"message": "无法打开上传的文件",
				})
				return
			}
			defer file.Close()

			// 2.2 解码上传的图片
			img, _, err = image.Decode(file)
			if err != nil {
				log.Error("解码上传的图像失败", "error", err)
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    -11,
					"message": "上传的文件不是有效的映像（支持PNG/JPEG）",
				})
				return
			}
		}

		// 3. 统一识别二维码（无论图片来自 URL 还是上传）
		content, err := utils.DecodeQR(img)
		if err != nil {
			log.Error("解码二维码失败，可能不存在二维码", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "图像中未找到二维码",
			})
			return
		}

		// 4. 返回识别结果
		c.JSON(http.StatusOK, gin.H{
			"code":    1,
			"message": content,
			"data":    gin.H{"content": content},
		})
	}
}
