package services

import (
	"strings"
	"time"

	"go-admin/internal/db"
	"go-admin/internal/models"
	"go-admin/pkg/utils"

	"github.com/jinzhu/gorm"
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

// Helper function to safely get timestamp
func safeGetTimestamp(submittedAt *time.Time) time.Time {
	if submittedAt != nil {
		return *submittedAt
	}
	return time.Now() // fallback to current time
}

// Helper function to safely get user info
func safeGetUser(user models.User) *User {
	if user.ID == 0 {
		return nil
	}
	return &User{
		ID:        user.ID,
		Name:      user.Name,
		XP:        user.XP,
		Level:     user.Level,
		AvatarURL: user.AvatarURL,
	}
}

type dateRange struct {
	start   time.Time
	end     time.Time
	enabled bool
}

// resolveDateRange normalizes the requested period and returns a time range.
// Default is the current week to maintain previous behavior.
func resolveDateRange(period string, now time.Time) dateRange {
	p := strings.ToLower(strings.TrimSpace(period))

	switch p {
	case "", "minggu ini", "minggu-ini", "minggu":
		start := utils.StartOfWeek(now)
		end := utils.EndOfWeek(now).AddDate(0, 0, 1).Add(-time.Nanosecond)
		return dateRange{start: start, end: end, enabled: true}
	case "bulan ini", "bulan-ini", "bulan":
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
		return dateRange{start: start, end: end, enabled: true}
	case "tahun ini", "tahun-ini", "tahun":
		start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		end := start.AddDate(1, 0, 0).Add(-time.Nanosecond)
		return dateRange{start: start, end: end, enabled: true}
	case "semua", "all", "semua data":
		return dateRange{enabled: false}
	default:
		start := utils.StartOfWeek(now)
		end := utils.EndOfWeek(now).AddDate(0, 0, 1).Add(-time.Nanosecond)
		return dateRange{start: start, end: end, enabled: true}
	}
}

func applyDateRange(query *gorm.DB, rng dateRange, column string) *gorm.DB {
	if !rng.enabled {
		return query
	}
	return query.Where(column+" BETWEEN ? AND ?", rng.start, rng.end)
}

func (s *DashboardService) GetDashboardStats(period string) *DashboardStats {
	database := db.DB
	now := time.Now()
	rng := resolveDateRange(period, now)

	// 1. Basic user stats
	var totalStudents, totalTeachers, totalParents, totalActiveUsers int
	database.Model(&models.User{}).Where("role = ?", "siswa").Count(&totalStudents)
	database.Model(&models.User{}).Where("role = ?", "guru").Count(&totalTeachers)
	database.Model(&models.User{}).Where("role = ?", "ortu").Count(&totalParents)
	database.Model(&models.User{}).Where("role != ?", "admin").Count(&totalActiveUsers)

	// 2. Challenge stats
	var activeIndividualChallenges, activeGroupChallenges int
	database.Model(&models.Challenge{}).Where("type = ? AND end_date >= ?", "individual", now).Count(&activeIndividualChallenges)
	database.Model(&models.Challenge{}).Where("type = ? AND end_date >= ?", "group", now).Count(&activeGroupChallenges)

	// 3. Habits and reflections
	var doneHabits, notDoneHabits, reflectionsToday int
	applyDateRange(
		database.Model(&models.HabitLog{}).Where("status = ?", "completed"),
		rng,
		"submitted_at",
	).Count(&doneHabits)

	applyDateRange(
		database.Model(&models.HabitLog{}).Where("status != ?", "completed"),
		rng,
		"submitted_at",
	).Count(&notDoneHabits)

	applyDateRange(
		database.Model(&models.Reflection{}),
		rng,
		"date",
	).Count(&reflectionsToday)

	// 4. Top students (by XP)
	var topStudents []models.User
	database.Where("role = ?", "siswa").Order("xp DESC").Limit(10).Find(&topStudents)

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

	// 5. Mood distribution (period-aware)
	var reflections []models.Reflection
	applyDateRange(
		database.Model(&models.Reflection{}),
		rng,
		"created_at",
	).Find(&reflections)

	moodDistribution := map[string]int{
		"happy":   0,
		"neutral": 0,
		"sad":     0,
		"angry":   0,
		"tired":   0,
	}
	for _, r := range reflections {
		if r.Mood != "" {
			moodDistribution[r.Mood]++
		}
	}

	// 6. Habit trends (last 5 weeks, clipped to selected range)
	referenceDate := now
	if rng.enabled {
		referenceDate = rng.end
	}
	var habitTrends []HabitTrend
	for i := 4; i >= 0; i-- {
		start := utils.StartOfWeek(referenceDate.AddDate(0, 0, -7*i))
		end := utils.EndOfWeek(referenceDate.AddDate(0, 0, -7*i)).AddDate(0, 0, 1).Add(-time.Nanosecond)

		// Clamp to requested range if needed
		if rng.enabled {
			if end.Before(rng.start) || start.After(rng.end) {
				continue
			}
			if start.Before(rng.start) {
				start = rng.start
			}
			if end.After(rng.end) {
				end = rng.end
			}
		}

		var done, notDone int

		// Habit selesai = completed
		applyDateRange(
			database.Model(&models.HabitLog{}).Where("status = ?", "completed"),
			dateRange{start: start, end: end, enabled: true},
			"date",
		).Count(&done)

		// Habit belum selesai = joined atau submitted
		applyDateRange(
			database.Model(&models.HabitLog{}).Where("status IN (?)", []string{"joined", "submitted"}),
			dateRange{start: start, end: end, enabled: true},
			"date",
		).Count(&notDone)

		habitTrends = append(habitTrends, HabitTrend{
			Week:    start.Format("Jan 2") + " - " + end.Format("Jan 2"),
			Done:    done,
			NotDone: notDone,
		})
	}

	// 7. Recent activities (period-aware) - FIXED SECTION
	var recentActivities []Activity

	// Challenge completions with safe pointer handling
	var challengeCompletions []models.ChallengeParticipant
	applyDateRange(
		database.Where("status = ?", "completed"),
		rng,
		"submitted_at",
	).
		Preload("Challenge").
		Preload("User").
		Order("submitted_at DESC").
		Limit(10).
		Find(&challengeCompletions)

	for _, c := range challengeCompletions {
		// Skip if essential data is missing
		if c.User.ID == 0 || c.Challenge.ID == 0 {
			continue
		}

		xp := c.Challenge.XP
		recentActivities = append(recentActivities, Activity{
			Type:      "challenge_completion",
			Message:   c.User.Name + " menyelesaikan Challenge " + c.Challenge.Title,
			Timestamp: safeGetTimestamp(c.SubmittedAt), // Safe timestamp access
			User:      safeGetUser(c.User),             // Safe user access
			XP:        &xp,
		})
	}

	// Habit completions with safe pointer handling
	var habitCompletions []models.HabitLog
	applyDateRange(
		database.Where("status = ?", "completed"),
		rng,
		"submitted_at",
	).
		Preload("Habit").
		Preload("User").
		Order("created_at DESC").
		Limit(10).
		Find(&habitCompletions)

	for _, h := range habitCompletions {
		// Skip if essential data is missing
		if h.User.ID == 0 || h.Habit.ID == 0 {
			continue
		}

		xp := h.Habit.XP
		recentActivities = append(recentActivities, Activity{
			Type:      "habit_completion", // Fixed typo from "habit_completions"
			Message:   h.User.Name + " menyelesaikan Habit " + h.Habit.Title,
			Timestamp: safeGetTimestamp(h.SubmittedAt), // Safe timestamp access
			User:      safeGetUser(h.User),             // Safe user access
			XP:        &xp,
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
