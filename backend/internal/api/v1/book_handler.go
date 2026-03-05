package v1

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"
	"voicepaper/config"
	"voicepaper/internal/model"
	"voicepaper/internal/repository"
	"voicepaper/internal/storage"

	"github.com/gin-gonic/gin"
)

type BookHandler struct {
	repo    *repository.BookRepository
	storage storage.Storage
	isOSS   bool
}

func NewBookHandler() *BookHandler {
	cfg := config.GetConfig()
	var st storage.Storage
	var err error
	isOSS := false

	// 根据配置创建存储实例
	if cfg.Storage.Type == "oss" {
		st, err = storage.NewOSSStorage(cfg)
		if err != nil {
			log.Printf("❌ OSS存储初始化失败: %v", err)
			st = storage.NewLocalStorage(cfg)
			log.Printf("⚠️  已降级使用本地存储")
		} else {
			log.Printf("✅ OSS存储初始化成功")
			isOSS = true
		}
	} else {
		st = storage.NewLocalStorage(cfg)
		log.Printf("✅ 使用本地存储")
	}

	return &BookHandler{
		repo:    repository.NewBookRepository(),
		storage: st,
		isOSS:   isOSS,
	}
}

// GetBooks 获取书籍列表
// GET /api/v1/books
func (h *BookHandler) GetBooks(c *gin.Context) {
	// 获取查询参数
	bookType := c.Query("type")   // 书籍类型：0-默认 1-小说 2-传记
	keyword := c.Query("keyword") // 搜索关键词

	var books []interface{}
	var err error

	// 根据不同查询条件获取数据
	if keyword != "" {
		// 搜索书籍
		booksData, err := h.repo.SearchBooks(keyword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for _, book := range booksData {
			books = append(books, h.buildBookResponse(&book))
		}
	} else if bookType != "" {
		// 按类型获取
		bt, err := strconv.Atoi(bookType)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book type"})
			return
		}
		booksData, err := h.repo.GetBooksByType(int8(bt))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for _, book := range booksData {
			books = append(books, h.buildBookResponse(&book))
		}
	} else {
		// 获取所有书籍
		booksData, err := h.repo.GetAllBooks()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for _, book := range booksData {
			books = append(books, h.buildBookResponse(&book))
		}
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("⚡ GetBooks: 总书籍数=%d", len(books))

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    books,
	})
}

// GetBook 获取书籍详情
// GET /api/v1/books/:id
func (h *BookHandler) GetBook(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	book, err := h.repo.GetBookByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    h.buildBookResponse(book),
	})
}

// GetBookByBookID 根据book_id获取书籍详情
// GET /api/v1/books/detail/:book_id
func (h *BookHandler) GetBookByBookID(c *gin.Context) {
	bookIDStr := c.Param("book_id")
	bookID, err := strconv.Atoi(bookIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}

	book, err := h.repo.GetBookByBookID(bookID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    h.buildBookResponse(book),
	})
}

// GetBookPoints 获取书籍章节列表
// GET /api/v1/books/:book_id/points
func (h *BookHandler) GetBookPoints(c *gin.Context) {
	bookIDStr := c.Param("book_id")
	bookID, err := strconv.Atoi(bookIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}

	// 先验证书籍是否存在
	book, err := h.repo.GetBookByBookID(bookID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	// 获取章节列表
	points, err := h.repo.GetBookPoints(bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var pointsData []gin.H
	for _, point := range points {
		pointsData = append(pointsData, h.buildPointResponse(&point))
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"book":   h.buildBookResponse(book),
			"points": pointsData,
			"total":  len(pointsData),
		},
	})
}

// GetBookPoint 获取章节详情
// GET /api/v1/books/:book_id/points/:point_id
func (h *BookHandler) GetBookPoint(c *gin.Context) {
	bookIDStr := c.Param("book_id")
	pointIDStr := c.Param("point_id")

	bookID, err := strconv.Atoi(bookIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}

	pointID, err := strconv.Atoi(pointIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid point ID"})
		return
	}

	point, err := h.repo.GetBookPointByBookIDAndPointID(bookID, int8(pointID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Point not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    h.buildPointResponse(point),
	})
}

// buildBookResponse 构建书籍响应数据（处理OSS签名）
func (h *BookHandler) buildBookResponse(book *model.BookInfo) gin.H {
	cover := book.Cover
	smallPic := book.SmallPic

	// 如果是OSS URL，生成签名URL
	if h.storage != nil {
		ctx := context.Background()

		// 处理封面图
		if cover != "" && strings.Contains(cover, "oss-cn-") {
			parts := strings.Split(cover, ".aliyuncs.com/")
			if len(parts) > 1 {
				path := parts[1]
				signedURL, err := h.storage.GetSignedURL(ctx, path, 86400) // 24小时有效期
				if err == nil {
					cover = signedURL
				} else {
					log.Printf("⚠️  生成封面图签名URL失败: %v", err)
				}
			}
		}

		// 处理小图
		if smallPic != "" && strings.Contains(smallPic, "oss-cn-") {
			parts := strings.Split(smallPic, ".aliyuncs.com/")
			if len(parts) > 1 {
				path := parts[1]
				signedURL, err := h.storage.GetSignedURL(ctx, path, 86400)
				if err == nil {
					smallPic = signedURL
				} else {
					log.Printf("⚠️  生成小图签名URL失败: %v", err)
				}
			}
		}
	}

	return gin.H{
		"id":           book.ID,
		"book_id":      book.BookID,
		"name":         book.Name,
		"cover":        cover,
		"small_pic":    smallPic,
		"subtitle":     book.Subtitle,
		"author_name":  book.AuthorName,
		"value":        book.Value,
		"author_desc":  book.AuthorDesc,
		"what_inside":  book.WhatInside,
		"learn":        book.Learn,
		"duration":     book.Duration,
		"total_size":   book.TotalSize,
		"voice_name":   book.VoiceName,
		"voice_ids":    book.VoiceIDs,
		"voice_source": book.VoiceSource,
		"point_count":  book.PointCount,
		"key_points":   book.KeyPoints,
		"book_type":    book.BookType,
		"is_online":    book.IsOnline,
		"created_at":   book.CreatedAt,
		"updated_at":   book.UpdatedAt,
	}
}

// buildPointResponse 构建章节响应数据（处理OSS签名）
func (h *BookHandler) buildPointResponse(point *model.BookPoint) gin.H {
	audioURL := point.AudioURL

	// 如果是OSS URL，生成签名URL
	if h.storage != nil && audioURL != "" && strings.Contains(audioURL, "oss-cn-") {
		ctx := context.Background()
		parts := strings.Split(audioURL, ".aliyuncs.com/")
		if len(parts) > 1 {
			path := parts[1]
			signedURL, err := h.storage.GetSignedURL(ctx, path, 86400) // 24小时有效期
			if err == nil {
				audioURL = signedURL
				log.Printf("✅ 生成章节音频签名URL成功")
			} else {
				log.Printf("⚠️  生成章节音频签名URL失败: %v", err)
			}
		}
	}

	return gin.H{
		"id":             point.ID,
		"book_id":        point.BookID,
		"name":           point.Name,
		"subtitle":       point.Subtitle,
		"author_name":    point.AuthorName,
		"point_id":       point.PointID,
		"point_title":    point.PointTitle,
		"point_info":     point.PointInfo,
		"lrc_point_info": point.LrcPointInfo,
		"audio_url":      audioURL,
		"audio_times":    point.AudioTimes,
		"audio_size":     point.AudioSize,
		"audio_md5":      point.AudioMD5,
		"voice_name":     point.VoiceName,
		"voice_source":   point.VoiceSource,
		"book_type":      point.BookType,
		"status":         point.Status,
		"is_online":      point.IsOnline,
		"created_at":     point.CreatedAt,
		"updated_at":     point.UpdatedAt,
	}
}
