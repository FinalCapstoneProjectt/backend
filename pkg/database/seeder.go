package database

import (
	"backend/internal/domain"
	"backend/pkg/enums"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedDatabase seeds the database with initial data
func SeedDatabase(db *gorm.DB) error {
	log.Println("Checking for seed data...")

	// Check if university already exists
	var universityCount int64
	db.Model(&domain.University{}).Count(&universityCount)
	if universityCount > 0 {
		log.Println("Database already seeded, skipping...")
		return nil
	}

	log.Println("Seeding database with initial data...")

	// 1. Create default university
	university := &domain.University{
		ID:               1,
		Name:             "Adama Science and Technology University",
		AcademicYear:     "2025/2026",
		ProjectPeriod:    "Semester 2",
		VisibilityRule:   "private",
		AICheckerEnabled: true,
	}
	if err := db.Create(university).Error; err != nil {
		log.Printf("Failed to create university: %v", err)
		return err
	}
	log.Println("✓ Created university: ASTU")

	// 2. Create departments
	departments := []domain.Department{
		{Name: "Computer Science", Code: "CS", UniversityID: 1},
		{Name: "Software Engineering", Code: "SE", UniversityID: 1},
		{Name: "Information Technology", Code: "IT", UniversityID: 1},
		{Name: "Electrical Engineering", Code: "EE", UniversityID: 1},
		{Name: "Mechanical Engineering", Code: "ME", UniversityID: 1},
	}

	for _, dept := range departments {
		if err := db.Create(&dept).Error; err != nil {
			log.Printf("Failed to create department %s: %v", dept.Code, err)
			return err
		}
		log.Printf("✓ Created department: %s (%s)", dept.Name, dept.Code)
	}

	// Password Helper
	hash := func(pwd string) string {
		h, _ := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
		return string(h)
	}

	// 3. Create default admin user
	admin := &domain.User{
		Name:          "Head of CS Department", // Changed name to reflect role
		Email:         "head_cs@astu.edu.et",   // Changed email
		Password:      hash("Admin@123"),
		Role:          enums.RoleAdmin,       // Acts as Dept Head
		UniversityID:  university.ID,
		DepartmentID:  1,                     // ⚠️ CRITICAL: Must belong to CS (ID: 1)
		IsActive:      true,
		EmailVerified: true,
	}

	if err := db.Create(admin).Error; err != nil {
		return err
	}
	log.Println("✓ Created admin user")

	// 4. Create sample advisor (Teacher)
	teacherDeptID := uint(1) // CS
	teacher := &domain.User{
		Name:          "Dr. John Doe",
		Email:         "teacher@astu.edu.et",
		Password:      hash("Teacher@123"),
		Role:          enums.RoleAdvisor, // Changed to match your Enums
		UniversityID:  university.ID,
		DepartmentID:  teacherDeptID,
		IsActive:      true,
		EmailVerified: true,
	}

	if err := db.Create(teacher).Error; err != nil {
		return err
	}
	log.Println("✓ Created advisor user")

	// 5. Create sample student (NEW!)
	student := &domain.User{
		Name:          "Jaefer Student",
		Email:         "student@astu.edu.et",
		Password:      hash("Student@123"),
		Role:          enums.RoleStudent,
		UniversityID:  university.ID,
		DepartmentID:  uint(2), // SE
		StudentID:     "ETS1234/14",
		IsActive:      true,
		EmailVerified: true,
	}

	if err := db.Create(student).Error; err != nil {
		return err
	}
	log.Println("✓ Created student user")

	log.Println("✓ Database seeded successfully!")
	log.Println("\nTest Credentials:")
	log.Println("─────────────────────────────────────────")
	log.Println("Admin:   head_cs@astu.edu.et / Admin@123")
	log.Println("Advisor: teacher@astu.edu.et / Teacher@123")
	log.Println("Student: student@astu.edu.et / Student@123")
	log.Println("─────────────────────────────────────────")

	return nil
}