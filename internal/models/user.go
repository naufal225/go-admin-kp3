package models

type User struct {
    ID        uint   `json:"id" gorm:"primary_key"`
    Name      string `json:"name"`
    Email     string `json:"email"`
    Password  string `json:"password"`
    Role      string `json:"role"` // admin, guru, siswa, ortu
    XP        int    `json:"xp"`
    Level     int    `json:"level"`
    AvatarURL string `json:"avatar_url"`
    ParentID  *uint  `json:"parent_id"`

    // Relasi (opsional)
    Children []User `json:"children" gorm:"foreignKey:ParentID"`
}