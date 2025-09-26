package services

import (
    "go-admin/internal/db"
    "go-admin/internal/models"
    "go-admin/pkg/utils"
    "time"
)

type DashboardService struct{}

type DashboardStats struct {
    TotalStudents              int                    `json:"total_students"`
    TotalTeachers              int                    `json:"total_teachers"`
    TotalParents               int                    `json:"total_parents"`
    TotalActiveUsers           int                    `json:"total_active_users"`
    ActiveIndividualChallenges int                    `json:"active_individual_challenges"`
    ActiveGroupChallenges      int                    `json:"active_group_challenges"`
    DoneHabits                 int                    `json:"done_habits"`
    NotDoneHabits              int                    `json:"not_done_habits"`
    ReflectionsToday           int                    `json:"reflections_today"`
    ForumPostsThisWeek         int                    `json:"forum_posts_this_week"`
    ForumCommentsThisWeek      int                    `json:"forum_comments_this_week"`
    TopStudents                []User                 `json:"top_students"`
    PopularBadge               *Badge                 `json:"popular_badge"`
    RecentActivities           []Activity             `json:"recent_activities"`
    MoodDistribution           map[string]int         `json:"mood_distribution"`
    HabitTrends                []HabitTrend           `json:"habit_trends"`
    ChallengeProgress          []ChallengeProgress    `json:"challenge_progress"`
    TopClasses                 []ClassStat            `json:"top_classes"`
    ForumStats                 ForumStats             `json:"forum_stats"`
}

type User struct {
    ID        uint   `json:"id"`
    Name      string `json:"name"`
    XP        int    `json:"xp"`
    Level     int    `json:"level"`
    AvatarURL string `json:"avatar_url"`
}

type Badge struct {
    ID   uint   `json:"id"`
    Name string `json:"name"`
}

type Activity struct {
    Type      string    `json:"type"`
    Message   string    `json:"message"`
    Timestamp time.Time `json:"timestamp"`
    User      *User     `json:"user,omitempty"`
    Badge     *Badge    `json:"badge,omitempty"`
    XP        *int      `json:"xp,omitempty"`
}

type HabitTrend struct {
    Week    string `json:"week"`
    Done    int    `json:"done"`
    NotDone int    `json:"not_done"`
}

type ChallengeProgress struct {
    Title              string `json:"title"`
    Type               string `json:"type"`
    TotalParticipants  int    `json:"total_participants"`
    CompletedParticipants int `json:"completed_participants"`
    CompletionRate     float64 `json:"completion_rate"`
}

type ClassStat struct {
    ClassName   string `json:"class_name"`
    AvgXP       int    `json:"avg_xp"`
    StudentCount int   `json:"student_count"`
}

type ForumStats struct {
    WeeklyStats []WeeklyForumStat `json:"weekly_stats"`
    ActiveUsers []User            `json:"active_users"`
}

type WeeklyForumStat struct {
    Week   string `json:"week"`
    Posts  int    `json:"posts"`
    Comments int  `json:"comments"`
}

func (s *DashboardService) GetDashboardStats() *DashboardStats {
    db := db.DB

    // 1. Basic Stats
    var totalStudents int
    db.Model(&models.User{}).Where("role = ?", "siswa").Count(&totalStudents)

    var totalTeachers int
    db.Model(&models.User{}).Where("role = ?", "guru").Count(&totalTeachers)

    var totalParents int
    db.Model(&models.User{}).Where("role = ?", "ortu").Count(&totalParents)

    var totalActiveUsers int
    db.Model(&models.User{}).Where("role != ?", "admin").Count(&totalActiveUsers)

    var activeIndividualChallenges int
    db.Model(&models.Challenge{}).Where("type = ? AND end_date >= ?", "individual", time.Now()).Count(&activeIndividualChallenges)

    var activeGroupChallenges int
    db.Model(&models.Challenge{}).Where("type = ? AND end_date >= ?", "group", time.Now()).Count(&activeGroupChallenges)

    startOfWeek := utils.StartOfWeek(time.Now())
    endOfWeek := utils.EndOfWeek(time.Now())

    var doneHabits int
    db.Model(&models.HabitLog{}).Where("status = ? AND date BETWEEN ? AND ?", "done", startOfWeek, endOfWeek).Count(&doneHabits)

    var notDoneHabits int
    db.Model(&models.HabitLog{}).Where("status = ? AND date BETWEEN ? AND ?", "not_done", startOfWeek, endOfWeek).Count(&notDoneHabits)

    var reflectionsToday int
    db.Model(&models.Reflection{}).Where("DATE(date) = ?", time.Now().Format("2006-01-02")).Count(&reflectionsToday)

    // 2. Forum Stats
    var forumPostsThisWeek int
    db.Model(&models.ForumPost{}).Where("created_at BETWEEN ? AND ?", startOfWeek, endOfWeek).Count(&forumPostsThisWeek)

    var forumCommentsThisWeek int
    db.Model(&models.ForumComment{}).Where("created_at BETWEEN ? AND ?", startOfWeek, endOfWeek).Count(&forumCommentsThisWeek)

    // 3. Top Students (by XP)
    var topStudents []models.User
    db.Where("role = ?", "siswa").Order("xp DESC").Limit(10).Find(&topStudents)

    users := make([]User, len(topStudents))
    for i, u := range topStudents {
        users[i] = User{
            ID:        u.ID,
            Name:      u.Name,
            XP:        u.XP,
            Level:     u.Level,
            AvatarURL: u.AvatarURL,
        }
    }

    // 4. Mood Distribution
    moodStart := time.Now().AddDate(0, 0, -7) // Last 7 days
    var reflections []models.Reflection
    db.Where("created_at >= ?", moodStart).Find(&reflections)

    moodDistribution := map[string]int{
        "happy":   0,
        "neutral": 0,
        "sad":     0,
        "angry":   0,
        "tired":   0,
    }
    for _, r := range reflections {
        moodDistribution[r.Mood]++
    }

    // 5. Habit Trends (Last 5 weeks)
    habitTrends := []HabitTrend{}
    for i := 4; i >= 0; i-- {
        start := time.Now().AddDate(0, 0, -7*i).Truncate(24*time.Hour).AddDate(0, 0, -int(time.Now().Weekday())+1)
        end := start.AddDate(0, 0, 6)

        var done int
        db.Model(&models.HabitLog{}).Where("status = ? AND date BETWEEN ? AND ?", "done", start, end).Count(&done)

        var notDone int
        db.Model(&models.HabitLog{}).Where("status = ? AND date BETWEEN ? AND ?", "not_done", start, end).Count(&notDone)

        habitTrends = append(habitTrends, HabitTrend{
            Week:    start.Format("Jan 2") + " - " + end.Format("Jan 2"),
            Done:    done,
            NotDone: notDone,
        })
    }

    // 6. Recent Activities (Last 15)
    var recentActivities []Activity

    // Challenge completions
    var challengeCompletions []models.ChallengeParticipant
    db.Where("status = ? AND submitted_at >= ?", "completed", time.Now().AddDate(0, 0, -7)).
        Preload("Challenge").
        Preload("User").
        Order("submitted_at DESC").
        Limit(10).
        Find(&challengeCompletions)

    for _, c := range challengeCompletions {
        xp := c.Challenge.XP
        recentActivities = append(recentActivities, Activity{
            Type:      "challenge_completion",
            Message:   c.User.Name + " menyelesaikan Challenge " + c.Challenge.Title,
            Timestamp: *c.SubmittedAt,
            User: &User{
                ID:   c.User.ID,
                Name: c.User.Name,
            },
            XP: &xp,
        })
    }

    // Sort by timestamp descending
    // (You might want to implement custom sorting here)

    return &DashboardStats{
        TotalStudents:              totalStudents,
        TotalTeachers:              totalTeachers,
        TotalParents:               totalParents,
        TotalActiveUsers:           totalActiveUsers,
        ActiveIndividualChallenges: activeIndividualChallenges,
        ActiveGroupChallenges:      activeGroupChallenges,
        DoneHabits:                 doneHabits,
        NotDoneHabits:              notDoneHabits,
        ReflectionsToday:           reflectionsToday,
        ForumPostsThisWeek:         forumPostsThisWeek,
        ForumCommentsThisWeek:      forumCommentsThisWeek,
        TopStudents:                users,
        MoodDistribution:           moodDistribution,
        HabitTrends:                habitTrends,
        RecentActivities:           recentActivities,
    }
}