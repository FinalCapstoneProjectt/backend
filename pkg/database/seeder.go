package database

import (
	"backend/internal/domain"
	"backend/pkg/enums"
	"errors"
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
	if universityCount == 0 {
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

		// continue to create admin/teacher below
	} else {
		log.Println("Database already seeded, ensuring admin user exists...")
	}

	// 3. Ensure default admin user exists (create or update)
	// Use admin@gmail.com with password 'password123' as requested
	var uni domain.University
	if err := db.First(&uni).Error; err != nil {
		// if no university exists, create one minimal entry
		uni = domain.University{
			Name:             "Adama Science and Technology University",
			AcademicYear:     "2025/2026",
			ProjectPeriod:    "Semester 2",
			VisibilityRule:   "private",
			AICheckerEnabled: true,
		}
		if err := db.Create(&uni).Error; err != nil {
			log.Printf("Failed to ensure university for admin: %v", err)
			return err
		}
	}

	adminEmail := "admin@gmail.com"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash admin password: %v", err)
		return err
	}

	// Ensure we have a valid department to satisfy FK constraint
	var dept domain.Department
	if err := db.Where("university_id = ?", uni.ID).First(&dept).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// create a default department
			d := domain.Department{Name: "Computer Science", Code: "CS", UniversityID: uni.ID}
			if err := db.Create(&d).Error; err != nil {
				log.Printf("Failed to create default department: %v", err)
				return err
			}
			dept = d
		} else {
			log.Printf("Failed to query departments: %v", err)
			return err
		}
	}

	var existingAdmin domain.User
	err = db.Where("email = ?", adminEmail).First(&existingAdmin).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			admin := &domain.User{
				Name:          "System Administrator",
				Email:         adminEmail,
				Password:      string(hashedPassword),
				Role:          enums.RoleAdmin,
				UniversityID:  uni.ID,
				DepartmentID:  dept.ID,
				IsActive:      true,
				EmailVerified: true,
			}
			if err := db.Create(admin).Error; err != nil {
				log.Printf("Failed to create admin user: %v", err)
				return err
			}
			log.Printf("✓ Created admin user: %s (password: password123)", adminEmail)
		} else {
			log.Printf("Failed to query admin user: %v", err)
			return err
		}
	} else {
		// update existing admin password/flags
		existingAdmin.Password = string(hashedPassword)
		existingAdmin.Role = enums.RoleAdmin
		existingAdmin.IsActive = true
		existingAdmin.EmailVerified = true
		existingAdmin.UniversityID = uni.ID
		existingAdmin.DepartmentID = dept.ID
		if err := db.Save(&existingAdmin).Error; err != nil {
			log.Printf("Failed to update admin user: %v", err)
			return err
		}
		log.Printf("✓ Updated admin user: %s (password reset)", adminEmail)
	}

	// 4. Create sample teacher for testing (only if not exists)
	teacherEmail := "teacher@astu.edu.et"
	var teacher domain.User
	if err := db.Where("email = ?", teacherEmail).First(&teacher).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		teacherDeptID := uint(1) // CS department
		hashedTeacherPassword, err := bcrypt.GenerateFromPassword([]byte("Teacher@123"), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Failed to hash teacher password: %v", err)
			return err
		}

		teacher := &domain.User{
			Name:          "Dr. John Doe",
			Email:         teacherEmail,
			Password:      string(hashedTeacherPassword),
			Role:          enums.RoleTeacher,
			UniversityID:  uni.ID,
			DepartmentID:  teacherDeptID,
			IsActive:      true,
			EmailVerified: true,
		}

		if err := db.Create(teacher).Error; err != nil {
			log.Printf("Failed to create teacher user: %v", err)
			return err
		}
		log.Println("✓ Created teacher user: teacher@astu.edu.et (password: Teacher@123)")
	}

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
