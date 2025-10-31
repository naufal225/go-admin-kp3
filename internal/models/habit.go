package models

import "time"

type Habit struct {
    ID          uint      `json:"id" gorm:"primary_key"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    UserID      uint      `json:"user_id"`
    XP          int       `json:"xp_reward"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type HabitLog struct {
    ID        uint      `json:"id" gorm:"primary_key"`
    HabitID   uint      `json:"habit_id"`
    Date      time.Time `json:"date"`
    Status    string    `json:"status"` // done, not_done
    UserID    uint      `json:"user_id"`
    SubmittedAt  *time.Time `json:"submitted_at"`
    CreatedAt    time.Time  `json:"created_at"`

    // Relasi
    Habit Habit `json:"habit" gorm:"foreignKey:HabitID"`
    User      User      `json:"user" gorm:"foreignKey:UserID"`
}