package main

import (
	"context"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"helpdesk-backend/internal/config"
	"helpdesk-backend/internal/service"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type seedUser struct {
	Name  string
	Email string
	Role  string
}

type seedCategory struct {
	Name            string
	DefaultSLAHours int
}

type seededTicket struct {
	ID          int64
	Status      string
	RequesterID int64
	AssigneeID  int64
	CreatedAt   time.Time
}

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	if cfg.AppEnv == "production" {
		log.Fatal("seed command is disabled in production")
	}

	db, err := sqlx.Connect("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := seed(ctx, db); err != nil {
		log.Fatalf("seed failed: %v", err)
	}

	log.Println("seed complete")
}

func seed(ctx context.Context, db *sqlx.DB) error {
	rng := rand.New(rand.NewSource(20260831))
	hash, err := (&service.AuthService{}).HashPassword("password123")
	if err != nil {
		return err
	}

	users, err := seedUsers(ctx, db, hash)
	if err != nil {
		return err
	}

	categories, err := seedCategoriesAndSLA(ctx, db)
	if err != nil {
		return err
	}

	if err := clearDemoTickets(ctx, db); err != nil {
		return err
	}

	tickets, err := seedTickets(ctx, db, rng, users, categories)
	if err != nil {
		return err
	}

	if err := seedTicketLogs(ctx, db, rng, tickets); err != nil {
		return err
	}

	return nil
}

func seedUsers(ctx context.Context, db *sqlx.DB, passwordHash string) (map[string]int64, error) {
	rows := []seedUser{
		{Name: "Super Admin", Email: "superadmin@test.local", Role: "super_admin"},
		{Name: "Admin IT", Email: "admin@test.local", Role: "admin"},
		{Name: "Budi Santoso", Email: "budi@test.local", Role: "staff"},
		{Name: "Sari Wulandari", Email: "sari@test.local", Role: "staff"},
		{Name: "Andi Pratama", Email: "andi@test.local", Role: "staff"},
		{Name: "Rina Kusuma", Email: "rina@test.local", Role: "staff"},
		{Name: "Citra Lestari", Email: "citra@test.local", Role: "end_user"},
		{Name: "Dimas Nugroho", Email: "dimas@test.local", Role: "end_user"},
		{Name: "Eka Putri", Email: "eka@test.local", Role: "end_user"},
		{Name: "Fajar Maulana", Email: "fajar@test.local", Role: "end_user"},
		{Name: "Gita Amalia", Email: "gita@test.local", Role: "end_user"},
		{Name: "Hendra Wijaya", Email: "hendra@test.local", Role: "end_user"},
		{Name: "Intan Permata", Email: "intan@test.local", Role: "end_user"},
		{Name: "Joko Saputra", Email: "joko@test.local", Role: "end_user"},
		{Name: "Kartika Sari", Email: "kartika@test.local", Role: "end_user"},
		{Name: "Lukman Hakim", Email: "lukman@test.local", Role: "end_user"},
		{Name: "Maya Anggraini", Email: "maya@test.local", Role: "end_user"},
		{Name: "Nadia Safitri", Email: "nadia@test.local", Role: "end_user"},
	}

	ids := map[string]int64{}
	for _, row := range rows {
		var id int64
		err := db.QueryRowContext(ctx, `
			INSERT INTO users (name, email, password_hash, role, is_active)
			VALUES ($1,$2,$3,$4,true)
			ON CONFLICT (email) DO UPDATE
			SET name=EXCLUDED.name,
			    password_hash=EXCLUDED.password_hash,
			    role=EXCLUDED.role,
			    is_active=true,
			    updated_at=now()
			RETURNING id
		`, row.Name, row.Email, passwordHash, row.Role).Scan(&id)
		if err != nil {
			return nil, err
		}
		ids[row.Email] = id
	}

	return ids, nil
}

func seedCategoriesAndSLA(ctx context.Context, db *sqlx.DB) (map[string]int64, error) {
	rows := []seedCategory{
		{Name: "Hardware", DefaultSLAHours: 48},
		{Name: "Software", DefaultSLAHours: 24},
		{Name: "Network", DefaultSLAHours: 12},
		{Name: "Access Request", DefaultSLAHours: 8},
	}
	priorities := map[string]float64{
		"Low":      1,
		"Medium":   0.75,
		"High":     0.5,
		"Critical": 0.25,
	}

	ids := map[string]int64{}
	for _, row := range rows {
		var id int64
		err := db.QueryRowContext(ctx, `
			INSERT INTO categories (name, default_sla_hours)
			VALUES ($1,$2)
			ON CONFLICT (name) DO UPDATE
			SET default_sla_hours=EXCLUDED.default_sla_hours,
			    updated_at=now()
			RETURNING id
		`, row.Name, row.DefaultSLAHours).Scan(&id)
		if err != nil {
			return nil, err
		}
		ids[row.Name] = id

		for priority, multiplier := range priorities {
			hours := int(float64(row.DefaultSLAHours) * multiplier)
			if hours < 1 {
				hours = 1
			}
			if _, err := db.ExecContext(ctx, `
				INSERT INTO sla_rules (category_id, priority, resolution_hours)
				VALUES ($1,$2,$3)
				ON CONFLICT (category_id, priority)
				DO UPDATE SET resolution_hours=EXCLUDED.resolution_hours, updated_at=now()
			`, id, priority, hours); err != nil {
				return nil, err
			}
		}
	}

	return ids, nil
}

func clearDemoTickets(ctx context.Context, db *sqlx.DB) error {
	_, err := db.ExecContext(ctx, `
		DELETE FROM tickets
		WHERE title LIKE '[DEMO]%'
	`)
	return err
}

func seedTickets(ctx context.Context, db *sqlx.DB, rng *rand.Rand, users map[string]int64, categories map[string]int64) ([]seededTicket, error) {
	staffIDs := []int64{
		users["budi@test.local"],
		users["sari@test.local"],
		users["andi@test.local"],
		users["rina@test.local"],
	}
	requesterIDs := []int64{
		users["citra@test.local"],
		users["dimas@test.local"],
		users["eka@test.local"],
		users["fajar@test.local"],
		users["gita@test.local"],
		users["hendra@test.local"],
		users["intan@test.local"],
		users["joko@test.local"],
		users["kartika@test.local"],
		users["lukman@test.local"],
		users["maya@test.local"],
		users["nadia@test.local"],
	}
	categoryNames := []string{"Hardware", "Software", "Network", "Access Request"}
	statuses := []string{"Open", "In Progress", "Pending", "Resolved", "Closed"}
	priorities := []string{"Low", "Medium", "High", "Critical"}
	titleSamples := map[string][]string{
		"Hardware":       {"Laptop lambat saat booting", "Monitor tidak menyala", "Keyboard eksternal tidak terdeteksi"},
		"Software":       {"Aplikasi payroll error", "Office tidak bisa aktivasi", "Browser crash saat membuka sistem internal"},
		"Network":        {"WiFi kantor sering putus", "VPN tidak bisa connect", "Koneksi printer jaringan gagal"},
		"Access Request": {"Request akses folder finance", "Reset akses sistem HRIS", "Permintaan akses aplikasi inventory"},
	}

	tickets := make([]seededTicket, 0, 80)
	now := time.Now().UTC()
	for i := 1; i <= 80; i++ {
		categoryName := categoryNames[rng.Intn(len(categoryNames))]
		priority := priorities[rng.Intn(len(priorities))]
		status := statuses[rng.Intn(len(statuses))]
		requesterID := requesterIDs[rng.Intn(len(requesterIDs))]
		assigneeID := staffIDs[rng.Intn(len(staffIDs))]
		createdAt := now.Add(-time.Duration(rng.Intn(45*24)) * time.Hour)
		slaHours := slaHoursFor(categoryName, priority)
		deadline := createdAt.Add(time.Duration(slaHours) * time.Hour)
		var resolvedAt *time.Time
		if status == "Resolved" || status == "Closed" {
			resolved := createdAt.Add(time.Duration(2+rng.Intn(slaHours+24)) * time.Hour)
			resolvedAt = &resolved
		}

		title := fmtDemoTitle(i, titleSamples[categoryName][rng.Intn(len(titleSamples[categoryName]))])
		description := "Data demo untuk menguji workflow tiket, dashboard SLA, staff performance, dan export CSV."

		var ticketID int64
		err := db.QueryRowContext(ctx, `
			INSERT INTO tickets (
				title, description, requester_id, assignee_id, category_id,
				priority, status, sla_deadline, resolved_at, created_at, updated_at
			)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
			RETURNING id
		`, title, description, requesterID, assigneeID, categories[categoryName], priority, status, deadline, resolvedAt, createdAt, createdAt).Scan(&ticketID)
		if err != nil {
			return nil, err
		}

		tickets = append(tickets, seededTicket{
			ID:          ticketID,
			Status:      status,
			RequesterID: requesterID,
			AssigneeID:  assigneeID,
			CreatedAt:   createdAt,
		})
	}

	return tickets, nil
}

func seedTicketLogs(ctx context.Context, db *sqlx.DB, rng *rand.Rand, tickets []seededTicket) error {
	for _, ticket := range tickets {
		if _, err := db.ExecContext(ctx, `
			INSERT INTO ticket_status_log (ticket_id, old_status, new_status, changed_by, changed_at)
			VALUES ($1,NULL,'Open',$2,$3)
		`, ticket.ID, ticket.RequesterID, ticket.CreatedAt); err != nil {
			return err
		}

		if _, err := db.ExecContext(ctx, `
			INSERT INTO ticket_activity_log (ticket_id, actor_id, action, old_value, new_value, created_at)
			VALUES ($1,$2,'ticket_created',NULL,'Open',$3)
		`, ticket.ID, ticket.RequesterID, ticket.CreatedAt); err != nil {
			return err
		}

		if _, err := db.ExecContext(ctx, `
			INSERT INTO ticket_activity_log (ticket_id, actor_id, action, old_value, new_value, created_at)
			VALUES ($1,$2,'assigned',NULL,$3,$4)
		`, ticket.ID, ticket.AssigneeID, idString(ticket.AssigneeID), ticket.CreatedAt.Add(30*time.Minute)); err != nil {
			return err
		}

		if ticket.Status != "Open" {
			if _, err := db.ExecContext(ctx, `
				INSERT INTO ticket_status_log (ticket_id, old_status, new_status, changed_by, changed_at)
				VALUES ($1,'Open',$2,$3,$4)
			`, ticket.ID, ticket.Status, ticket.AssigneeID, ticket.CreatedAt.Add(time.Duration(1+rng.Intn(8))*time.Hour)); err != nil {
				return err
			}
		}

		if _, err := db.ExecContext(ctx, `
			INSERT INTO ticket_comments (ticket_id, user_id, message, created_at)
			VALUES ($1,$2,$3,$4), ($1,$5,$6,$7)
		`,
			ticket.ID,
			ticket.RequesterID,
			"Mohon dibantu dicek. Terima kasih.",
			ticket.CreatedAt.Add(15*time.Minute),
			ticket.AssigneeID,
			"Baik, ticket sudah kami terima dan sedang ditindaklanjuti.",
			ticket.CreatedAt.Add(time.Duration(2+rng.Intn(6))*time.Hour),
		); err != nil {
			return err
		}
	}

	return nil
}

func slaHoursFor(category, priority string) int {
	base := map[string]int{
		"Hardware":       48,
		"Software":       24,
		"Network":        12,
		"Access Request": 8,
	}[category]
	multiplier := map[string]float64{
		"Low":      1,
		"Medium":   0.75,
		"High":     0.5,
		"Critical": 0.25,
	}[priority]
	hours := int(float64(base) * multiplier)
	if hours < 1 {
		return 1
	}
	return hours
}

func fmtDemoTitle(number int, title string) string {
	return "[DEMO] " + strings.TrimSpace(title) + " #" + idString(int64(number))
}

func idString(id int64) string {
	return strconv.FormatInt(id, 10)
}
