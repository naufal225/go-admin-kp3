package services

import (
	"go-admin/internal/db"
	"go-admin/internal/models"
	"go-admin/pkg/utils"
	"time"
)

type DashboardService struct{}

type DashboardStats struct {
	TotalStudents              int                 `json:"total_students"`
	TotalTeachers              int                 `json:"total_teachers"`
	TotalParents               int                 `json:"total_parents"`
	TotalActiveUsers           int                 `json:"total_active_users"`
	ActiveIndividualChallenges int                 `json:"active_individual_challenges"`
	ActiveGroupChallenges      int                 `json:"active_group_challenges"`
	DoneHabits                 int                 `json:"done_habits"`
	NotDoneHabits              int                 `json:"not_done_habits"`
	ReflectionsToday           int                 `json:"reflections_today"`
	TopStudents                []User              `json:"top_students"`
	RecentActivities           []Activity          `json:"recent_activities"`
	MoodDistribution           map[string]int      `json:"mood_distribution"`
	HabitTrends                []HabitTrend        `json:"habit_trends"`
	ChallengeProgress          []ChallengeProgress `json:"challenge_progress"`
	TopClasses                 []ClassStat         `json:"top_classes"`
}

type User struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	XP        int    `json:"xp"`
	Level     int    `json:"level"`
	AvatarURL string `json:"avatar_url"`
}

type Activity struct {
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	User      *User     `json:"user,omitempty"`
	XP        *int      `json:"xp,omitempty"`
}

type HabitTrend struct {
	Week    string `json:"week"`
	Done    int    `json:"done"`
	NotDone int    `json:"not_done"`
}

type ChallengeProgress struct {
	Title                 string  `json:"title"`
	Type                  string  `json:"type"`
	TotalParticipants     int     `json:"total_participants"`
	CompletedParticipants int     `json:"completed_participants"`
	CompletionRate        float64 `json:"completion_rate"`
}

type ClassStat struct {
	ClassName    string `json:"class_name"`
	AvgXP        int    `json:"avg_xp"`
	StudentCount int    `json:"student_count"`
}

func (s *DashboardService) GetDashboardStats() *DashboardStats {
	db := db.DB

	// 1. Basic user stats
	var totalStudents, totalTeachers, totalParents, totalActiveUsers int
	db.Model(&models.User{}).Where("role = ?", "siswa").Count(&totalStudents)
	db.Model(&models.User{}).Where("role = ?", "guru").Count(&totalTeachers)
	db.Model(&models.User{}).Where("role = ?", "ortu").Count(&totalParents)
	db.Model(&models.User{}).Where("role != ?", "admin").Count(&totalActiveUsers)

	// 2. Challenge stats
	var activeIndividualChallenges, activeGroupChallenges int
	db.Model(&models.Challenge{}).Where("type = ? AND end_date >= ?", "individual", time.Now()).Count(&activeIndividualChallenges)
	db.Model(&models.Challenge{}).Where("type = ? AND end_date >= ?", "group", time.Now()).Count(&activeGroupChallenges)

	// 3. Habits and reflections
	startOfWeek := utils.StartOfWeek(time.Now())
	endOfWeek := utils.EndOfWeek(time.Now())

	var doneHabits, notDoneHabits, reflectionsToday int
	db.Model(&models.HabitLog{}).
		Where("status = ? AND submitted_at BETWEEN ? AND ?", "completed", startOfWeek, endOfWeek).
		Count(&doneHabits)

	db.Model(&models.HabitLog{}).
		Where("status != ? AND submitted_at BETWEEN ? AND ?", "completed", startOfWeek, endOfWeek).
		Count(&notDoneHabits)
	db.Model(&models.Reflection{}).Where("DATE(date) = ?", time.Now().Format("2006-01-02")).Count(&reflectionsToday)

	// 4. Top students (by XP)
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

	// 5. Mood distribution (last 7 days)
	moodStart := time.Now().AddDate(0, 0, -7)
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

	// 6. Habit trends (last 5 weeks)
	var habitTrends []HabitTrend
	for i := 4; i >= 0; i-- {
		start := utils.StartOfWeek(time.Now().AddDate(0, 0, -7*i))
		end := utils.EndOfWeek(time.Now().AddDate(0, 0, -7*i))

		var done, notDone int

		// Habit selesai = completed
		db.Model(&models.HabitLog{}).
			Where("status = ? AND date BETWEEN ? AND ?", "completed", start, end).
			Count(&done)

		// Habit belum selesai = joined atau submitted
		db.Model(&models.HabitLog{}).
			Where("status IN (?) AND date BETWEEN ? AND ?", []string{"joined", "submitted"}, start, end).
			Count(&notDone)

		habitTrends = append(habitTrends, HabitTrend{
			Week:    start.Format("Jan 2") + " - " + end.Format("Jan 2"),
			Done:    done,
			NotDone: notDone,
		})
	}

	// 7. Recent activities (last 7 days)
	var recentActivities []Activity
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

	var habitCompletions []models.HabitLog
	db.Where("status = ? AND created_at >= ?", "completed", time.Now().AddDate(0, 0, -7)).
		Preload("Habit").
		Preload("User").
		Order("created_at DESC").
		Limit(10).
		Find(&habitCompletions)

	for _, h := range habitCompletions {
		xp := h.Habit.XP
		recentActivities = append(recentActivities, Activity{
			Type:      "habit_completions",
			Message:   h.User.Name + " menyelesaikan Habit " + h.Habit.Title,
			Timestamp: *h.SubmittedAt,
			User: &User{
				ID:   h.User.ID,
				Name: h.User.Name,
			},
			XP: &xp,
		})
	}

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
		TopStudents:                users,
		MoodDistribution:           moodDistribution,
		HabitTrends:                habitTrends,
		RecentActivities:           recentActivities,
	}
}
