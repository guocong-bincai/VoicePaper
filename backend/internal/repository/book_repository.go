package repository

import (
	"voicepaper/internal/model"

	"gorm.io/gorm"
)

// BookRepository 书籍仓储
type BookRepository struct {
	db *gorm.DB
}

// NewBookRepository 创建书籍仓储实例
func NewBookRepository() *BookRepository {
	return &BookRepository{
		db: DB,
	}
}

// GetAllBooks 获取所有已上线的书籍列表
func (r *BookRepository) GetAllBooks() ([]model.BookInfo, error) {
	var books []model.BookInfo
	err := r.db.Where("is_online = ?", 1).
		Order("id DESC").
		Find(&books).Error
	return books, err
}

// GetBookByID 根据ID获取书籍详情
func (r *BookRepository) GetBookByID(id uint) (*model.BookInfo, error) {
	var book model.BookInfo
	err := r.db.Where("id = ?", id).First(&book).Error
	if err != nil {
		return nil, err
	}
	return &book, nil
}

// GetBookByBookID 根据book_id获取书籍详情
func (r *BookRepository) GetBookByBookID(bookID int) (*model.BookInfo, error) {
	var book model.BookInfo
	err := r.db.Where("book_id = ?", bookID).First(&book).Error
	if err != nil {
		return nil, err
	}
	return &book, nil
}

// GetBookPoints 获取书籍的所有章节
func (r *BookRepository) GetBookPoints(bookID int) ([]model.BookPoint, error) {
	var points []model.BookPoint
	err := r.db.Where("book_id = ? AND is_online = ?", bookID, 1).
		Order("point_id ASC").
		Find(&points).Error
	return points, err
}

// GetBookPointByID 根据ID获取单个章节详情
func (r *BookRepository) GetBookPointByID(id uint) (*model.BookPoint, error) {
	var point model.BookPoint
	err := r.db.Where("id = ?", id).First(&point).Error
	if err != nil {
		return nil, err
	}
	return &point, nil
}

// GetBookPointByBookIDAndPointID 根据book_id和point_id获取章节
func (r *BookRepository) GetBookPointByBookIDAndPointID(bookID int, pointID int8) (*model.BookPoint, error) {
	var point model.BookPoint
	err := r.db.Where("book_id = ? AND point_id = ?", bookID, pointID).First(&point).Error
	if err != nil {
		return nil, err
	}
	return &point, nil
}

// SearchBooks 搜索书籍（支持书名和作者名全文搜索）
func (r *BookRepository) SearchBooks(keyword string) ([]model.BookInfo, error) {
	var books []model.BookInfo
	err := r.db.Where("is_online = ? AND MATCH(name, author_name) AGAINST(? IN NATURAL LANGUAGE MODE)", 1, keyword).
		Order("id DESC").
		Find(&books).Error
	return books, err
}

// GetBooksByType 根据类型获取书籍列表
func (r *BookRepository) GetBooksByType(bookType int8) ([]model.BookInfo, error) {
	var books []model.BookInfo
	err := r.db.Where("is_online = ? AND book_type = ?", 1, bookType).
		Order("id DESC").
		Find(&books).Error
	return books, err
}
