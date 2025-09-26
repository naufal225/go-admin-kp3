package models

import "time"

type ForumPost struct {
    ID        uint      `json:"id" gorm:"primary_key"`
    Title     string    `json:"title"`
    Content   string    `json:"content"`
    UserID    uint      `json:"user_id"`
    CreatedAt time.Time `json:"created_at"`

    // Relasi
    User User `json:"user" gorm:"foreignKey:UserID"`
}

type ForumComment struct {
    ID        uint      `json:"id" gorm:"primary_key"`
    PostID    uint      `json:"post_id"`
    Content   string    `json:"content"`
    UserID    uint      `json:"user_id"`
    CreatedAt time.Time `json:"created_at"`

    // Relasi
    Post ForumPost `json:"post" gorm:"foreignKey:PostID"`
    User User      `json:"user" gorm:"foreignKey:UserID"`
}