package models

import "time"

type Challenge struct {
    ID          uint      `json:"id" gorm:"primary_key"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Type        string    `json:"type"` // individual, group
    EndDate     time.Time `json:"end_date"`
    XP          int       `json:"xp_reward"`
    CreatedAt   time.Time `json:"created_at"`
}

type ChallengeParticipant struct {
    ID           uint      `json:"id" gorm:"primary_key"`
    ChallengeID  uint      `json:"challenge_id"`
    UserID       uint      `json:"user_id"`
    Status       string    `json:"status"` // completed, in_progress
    SubmittedAt  *time.Time `json:"submitted_at"`
    CreatedAt    time.Time  `json:"created_at"`

    // Relasi
    Challenge Challenge `json:"challenge" gorm:"foreignKey:ChallengeID"`
    User      User      `json:"user" gorm:"foreignKey:UserID"`
}