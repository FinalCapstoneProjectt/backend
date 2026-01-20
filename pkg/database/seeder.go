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

	// 3. Create default admin user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("Admin@123"), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash admin password: %v", err)
		return err
	}

	admin := &domain.User{
		Name:          "System Administrator",
		Email:         "admin@astu.edu.et",
		Password:      string(hashedPassword),
		Role:          enums.RoleAdmin,
		UniversityID:  university.ID,
		IsActive:      true,
		EmailVerified: true,
	}

	if err := db.Create(admin).Error; err != nil {
		log.Printf("Failed to create admin user: %v", err)
		return err
	}
	log.Println("✓ Created admin user: admin@astu.edu.et (password: Admin@123)")

	// 4. Create sample teacher for testing
	teacherDeptID := uint(1) // CS department
	hashedTeacherPassword, err := bcrypt.GenerateFromPassword([]byte("Teacher@123"), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash teacher password: %v", err)
		return err
	}

	teacher := &domain.User{
		Name:          "Dr. John Doe",
		Email:         "teacher@astu.edu.et",
		Password:      string(hashedTeacherPassword),
		Role:          enums.RoleTeacher,
		UniversityID:  university.ID,
		DepartmentID:  teacherDeptID,
		IsActive:      true,
		EmailVerified: true,
	}

	if err := db.Create(teacher).Error; err != nil {
		log.Printf("Failed to create teacher user: %v", err)
		return err
	}
	log.Println("✓ Created teacher user: teacher@astu.edu.et (password: Teacher@123)")

	log.Println("✓ Database seeded successfully!")
	log.Println("\nTest Credentials:")
	log.Println("─────────────────────────────────────────")
	log.Println("Admin:   admin@astu.edu.et   / Admin@123")
	log.Println("Teacher: teacher@astu.edu.et / Teacher@123")
	log.Println("─────────────────────────────────────────")
	log.Println("\nAvailable Departments:")
	log.Println("─────────────────────────────────────────")
	log.Println("1 - Computer Science (CS)")
	log.Println("2 - Software Engineering (SE)")
	log.Println("3 - Information Technology (IT)")
	log.Println("4 - Electrical Engineering (EE)")
	log.Println("5 - Mechanical Engineering (ME)")
	log.Println("─────────────────────────────────────────")

	return nil
}
