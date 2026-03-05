package v1

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
	"github.com/yuin/goldmark"

	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// ExportArticlePDF uses Chromedp to generate PDF from HTML
func (h *ArticleHandler) ExportArticlePDF(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// 1. Fetch Article
	article, err := h.repo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	// 2. Load Content
	var content string
	if article.ArticleURL != "" {
		content, err = h.loadArticleContent(article.ArticleURL)
		if err != nil {
			fmt.Printf("⚠️ Failed to load article content: %v\n", err)
			content = "Content loading failed."
		}
	} else {
		content = "No content available."
	}

	// 3. Convert Markdown to HTML
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
			html.WithUnsafe(),
		),
	)
	var buf bytes.Buffer
	if err := md.Convert([]byte(content), &buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Markdown conversion failed"})
		return
	}
	articleHTML := buf.String()

	// Clean title
	cleanTitle := strings.ReplaceAll(article.Title, "**", "")

	// Calculate metadata
	wordCount := len(strings.Fields(content))
	dateStr := time.Now().Format("Jan 02, 2006")
	if article.PublishDate != nil {
		dateStr = article.PublishDate.Format("Jan 02, 2006")
	}

	// 4. Prepare Data for Template
	data := gin.H{
		"Title":        cleanTitle,
		"Category":     "每日精读",
		"Content":      template.HTML(articleHTML),
		"Words":        article.Words,
		"Watermark":    "外刊精读微信小程序：音文书",
		"WatermarkURL": "voicepaper.top",
		"Year":         time.Now().Year(),
		"Date":         dateStr,
		"WordCount":    wordCount,
	}
	if article.Category != nil {
		data["Category"] = article.Category.Name
	}

	// 📝 不再使用 Base64 字体，改用系统字体
	fmt.Println("📝 Using system fonts for PDF generation")

	// HTML Template - 简洁版，参考截图设计
	tmplStr := `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <title>{{ .Title }}</title>
    <style>
        /* 使用系统中文字体 */
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: 'WenQuanYi Micro Hei', 'WenQuanYi Zen Hei', 'Microsoft YaHei', 'SimHei', sans-serif;
            font-size: 14px;
            line-height: 1.8;
            color: #333;
            background: #fff;
            padding: 40px;
        }

        /* 标题区域 */
        .header {
            margin-bottom: 30px;
            padding-bottom: 20px;
            border-bottom: 1px solid #e0e0e0;
        }

        .meta {
            font-size: 12px;
            color: #999;
            margin-bottom: 12px;
        }

        .category {
            display: inline-block;
            background: #FFA500;
            color: #fff;
            padding: 2px 10px;
            border-radius: 3px;
            font-size: 11px;
            margin-left: 8px;
        }

        h1.title {
            font-size: 24px;
            font-weight: 700;
            color: #000;
            line-height: 1.4;
            margin: 0;
        }

        /* 内容布局 */
        .container {
            display: flex;
            gap: 30px;
        }

        /* 左侧主内容 */
        .main-content {
            flex: 1;
            min-width: 0;
        }

        /* 文章内容 */
        .article-content {
            font-size: 14px;
            line-height: 2;
        }

        .article-content p {
            margin-bottom: 14px;
            text-align: justify;
        }

        .article-content h2 {
            font-size: 18px;
            font-weight: 700;
            margin: 25px 0 15px 0;
        }

        .article-content h3 {
            font-size: 16px;
            font-weight: 700;
            margin: 20px 0 12px 0;
        }

        .article-content ul, .article-content ol {
            margin-left: 20px;
            margin-bottom: 14px;
        }

        .article-content li {
            margin-bottom: 6px;
        }

        .article-content strong {
            color: #d32f2f;
        }

        /* 隐藏文章内容中的第一个标题（避免重复） */
        .article-content > h1:first-child {
            display: none;
        }

        /* 右侧词汇栏 */
        .vocab-sidebar {
            width: 220px;
            flex-shrink: 0;
        }

        .vocab-header {
            font-size: 13px;
            font-weight: 700;
            color: #FFA500;
            margin-bottom: 15px;
            display: flex;
            align-items: center;
            gap: 5px;
        }

        .vocab-card {
            background: #f0f8ff;
            padding: 12px;
            margin-bottom: 12px;
            border-radius: 6px;
            font-size: 12px;
            page-break-inside: avoid;
        }

        .word {
            font-weight: 700;
            color: #008080;
            font-size: 14px;
            margin-bottom: 4px;
        }

        .phonetic {
            color: #666;
            font-size: 11px;
            font-style: italic;
            margin-bottom: 4px;
        }

        .meaning {
            color: #333;
            line-height: 1.6;
        }

        /* 右下角固定水印 */
        .watermark {
            position: fixed;
            bottom: 15px;
            right: 15px;
            font-size: 10px;
            color: #999;
            background: rgba(255, 255, 255, 0.9);
            padding: 5px 10px;
            border-radius: 4px;
            border: 1px solid #e0e0e0;
            z-index: 1000;
        }

        /* 斜向平铺水印层 */
        .watermark-overlay {
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            pointer-events: none;
            z-index: 1;
            overflow: hidden;
        }

        .watermark-text {
            position: absolute;
            font-size: 48px;
            color: rgba(0, 0, 0, 0.03);
            font-weight: 700;
            white-space: nowrap;
            transform: rotate(-45deg);
            user-select: none;
        }

        @media print {
            body {
                -webkit-print-color-adjust: exact;
                print-color-adjust: exact;
            }
        }
    </style>
</head>
<body>
    <!-- 斜向平铺水印 -->
    <div class="watermark-overlay">
        <div class="watermark-text" style="top: 15%; left: 10%;">{{ .WatermarkURL }}</div>
        <div class="watermark-text" style="top: 15%; left: 60%;">{{ .WatermarkURL }}</div>
        <div class="watermark-text" style="top: 40%; left: 10%;">{{ .WatermarkURL }}</div>
        <div class="watermark-text" style="top: 40%; left: 60%;">{{ .WatermarkURL }}</div>
        <div class="watermark-text" style="top: 65%; left: 10%;">{{ .WatermarkURL }}</div>
        <div class="watermark-text" style="top: 65%; left: 60%;">{{ .WatermarkURL }}</div>
    </div>

    <!-- 右下角固定水印 -->
    <div class="watermark">{{ .Watermark }}</div>

    <!-- 标题区域 -->
    <div class="header">
        <div class="meta">
            {{ .Date }}
            <span class="category">{{ .Category }}</span>
            {{ .WordCount }} 词
        </div>
        <h1 class="title">{{ .Title }}</h1>
    </div>

    <!-- 主体内容 -->
    <div class="container">
        <!-- 左侧主内容 -->
        <div class="main-content">
            <!-- 文章正文 -->
            <div class="article-content">
                {{ .Content }}
            </div>
        </div>

        <!-- 右侧词汇栏 -->
        <div class="vocab-sidebar">
            {{ if .Words }}
            <div class="vocab-header">📌 重点词汇 {{ len .Words }}</div>
            <!-- 词汇循环显示多次，填充右侧空间 -->
            {{ range .Words }}
            <div class="vocab-card">
                <div class="word">{{ .Text }}</div>
                {{ if .Phonetic }}
                <div class="phonetic">/{{ .Phonetic }}/</div>
                {{ end }}
                <div class="meaning">{{ .Meaning }}</div>
            </div>
            {{ end }}
            <!-- 如果词汇少，再重复一遍 -->
            {{ range .Words }}
            <div class="vocab-card">
                <div class="word">{{ .Text }}</div>
                {{ if .Phonetic }}
                <div class="phonetic">/{{ .Phonetic }}/</div>
                {{ end }}
                <div class="meaning">{{ .Meaning }}</div>
            </div>
            {{ end }}
            {{ else }}
            <!-- 如果没有词汇，显示提示 -->
            <div style="color: #999; font-size: 12px; padding: 20px; text-align: center;">
                暂无重点词汇
            </div>
            {{ end }}
        </div>
    </div>
</body>
</html>
`

	// Render Template
	t, err := template.New("pdf").Parse(tmplStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Template parsing failed"})
		return
	}
	var htmlBuf bytes.Buffer
	if err := t.Execute(&htmlBuf, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Template execution failed"})
		return
	}

	finalHTML := htmlBuf.String()

	// Generate PDF via Chromedp
	fmt.Println("🚀 Starting Chromedp PDF generation...")

	// 设置环境变量
	os.Setenv("GOOGLE_API_KEY", "no")
	os.Setenv("GOOGLE_DEFAULT_CLIENT_ID", "no")
	os.Setenv("GOOGLE_DEFAULT_CLIENT_SECRET", "no")

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.NoSandbox,
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-background-networking", true),
		chromedp.Flag("disable-component-update", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-plugins", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("metrics-recording-only", true),
		chromedp.Flag("no-proxy-server", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.Flag("disable-features", "TranslateUI"),
		// 内存优化
		chromedp.Flag("single-process", true),
		chromedp.Flag("disable-software-rasterizer", true),
		chromedp.Flag("disable-dev-tools", true),
		chromedp.Flag("disable-logging", true),
		chromedp.Flag("disable-breakpad", true),
		chromedp.Flag("js-flags", "--max-old-space-size=256"),
	)

	// Check for chromium-browser or chromium
	execPath := ""
	paths := []string{
		"/usr/bin/chromium-browser",
		"/usr/bin/chromium",
		"/usr/bin/google-chrome",
		"/usr/bin/google-chrome-stable",
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			execPath = p
			break
		}
	}
	if execPath != "" {
		opts = append(opts, chromedp.ExecPath(execPath))
	}

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	fmt.Println("⏳ Navigating and generating PDF...")
	var pdfBuf []byte
	if err := chromedp.Run(ctx,
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("📄 Setting document content...")
			frameTree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				return err
			}
			if err := page.SetDocumentContent(frameTree.Frame.ID, finalHTML).Do(ctx); err != nil {
				return err
			}

			fmt.Println("⏳ Waiting for fonts to load...")
			time.Sleep(5 * time.Second) // 等待字体加载

			return nil
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("🖨️ Printing to PDF...")
			buf, _, err := page.PrintToPDF().
				WithPrintBackground(true).
				WithMarginTop(0.4).
				WithMarginBottom(0.4).
				WithMarginLeft(0.4).
				WithMarginRight(0.4).
				Do(ctx)
			if err != nil {
				return err
			}
			pdfBuf = buf
			return nil
		}),
	); err != nil {
		fmt.Printf("❌ Chromedp error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("PDF generation failed: %v", err)})
		return
	}
	fmt.Println("✅ PDF generated successfully")

	// Serve PDF
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.pdf\"", article.Title))
	c.Data(http.StatusOK, "application/pdf", pdfBuf)
}
