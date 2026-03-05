package main

import (
	"fmt"
	"log"
	"os"
	"time"
	"voicepaper/config"
	"voicepaper/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// 1. Load Config
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Connect to Database (Directly using gorm to avoid side effects of repository.InitDB if any)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Database,
		cfg.Database.Charset,
		cfg.Database.ParseTime,
		cfg.Database.Loc,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("✅ Database connected successfully")

	// 3. Find User
	var user model.User
	// Get email from command line args or environment variable
	email := os.Getenv("USER_EMAIL")
	if len(os.Args) > 1 {
		email = os.Args[1]
	}
	if email == "" {
		// Fallback to first user in database
		if err := db.First(&user, 1).Error; err != nil {
			log.Fatalf("Failed to find user. Usage: go run main.go <email> or set USER_EMAIL env var")
		}
	} else if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		log.Fatalf("Failed to find user with email %s: %v", email, err)
	}
	fmt.Printf("User found: ID=%d, Email=%s\n", user.ID, user.Email)

	// 4. Inspect Vocabulary
	var vocabs []model.Vocabulary
	if err := db.Where("user_id = ?", user.ID).Find(&vocabs).Error; err != nil {
		log.Fatalf("Failed to find vocabularies: %v", err)
	}

	fmt.Printf("Total vocabulary items: %d\n", len(vocabs))

	now := time.Now()
	fmt.Printf("Current time: %s\n", now.Format("2006-01-02 15:04:05"))

	countReview := 0
	for _, v := range vocabs {
		isDue := v.NextReviewAt == nil || v.NextReviewAt.Before(now)
		status := "Unknown"
		if v.NextReviewAt == nil {
			status = "New (Due)"
		} else if v.NextReviewAt.Before(now) {
			status = fmt.Sprintf("Due (was %s)", v.NextReviewAt.Format("15:04:05"))
		} else {
			status = fmt.Sprintf("Future (%s)", v.NextReviewAt.Format("15:04:05"))
		}

		if isDue {
			countReview++
			fmt.Printf("  [REVIEW] ID: %d, Content: %s, Meaning: '%s', Mastery: %d, Next: %v\n",
				v.ID, v.Content, v.Meaning, v.MasteryLevel, status)
		} else {
			// Uncomment to see future items
			fmt.Printf("  [FUTURE] ID: %d, Content: %s, Mastery: %d, Next: %v\n",
				v.ID, v.Content, v.MasteryLevel, status)
		}
	}

	fmt.Printf("Total items due for review according to logic: %d\n", countReview)

	// 5. Check GetTodayReviewList logic specifically
	var todayReviews []model.Vocabulary
	err = db.Where("user_id = ? AND (next_review_at IS NULL OR next_review_at <= ?)", user.ID, now).
		Find(&todayReviews).Error
	if err != nil {
		log.Printf("Query error: %v", err)
	}
	fmt.Printf("DB Query returned count: %d\n", len(todayReviews))
}
