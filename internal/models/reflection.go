package models

import "time"

type Reflection struct {
    ID        uint      `json:"id" gorm:"primary_key"`
    UserID    uint      `json:"user_id"`
    Date      time.Time `json:"date"`
    Mood      string    `json:"mood"` // happy, sad, neutral, etc
    Content   string    `json:"content"`
    CreatedAt time.Time `json:"created_at"`

    // Relasi (opsional)
    User User `json:"user" gorm:"foreignKey:UserID"`
}