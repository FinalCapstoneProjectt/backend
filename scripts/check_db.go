package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type University struct {
	ID   uint `gorm:"primaryKey"`
	Name string
}

type Department struct {
	ID   uint `gorm:"primaryKey"`
	Name string
	Code string
}

type User struct {
	ID    uint `gorm:"primaryKey"`
	Name  string
	Email string
	Role  string
}

func main() {
	// Use your Neon database credentials
	dsn := "host=ep-weathered-bar-a41tdfpo-pooler.us-east-1.aws.neon.tech user=neondb_owner password=npg_OJdoyXhQ1vn6 dbname=neondb port=5432 sslmode=require"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	fmt.Println("✓ Database connection successful!")
	fmt.Println("\n=== Checking Database Contents ===\n")

	// Check universities
	var universities []University
	db.Find(&universities)
	fmt.Printf("Universities: %d found\n", len(universities))
	for _, u := range universities {
		fmt.Printf("  - ID: %d, Name: %s\n", u.ID, u.Name)
	}

	// Check departments
	var departments []Department
	db.Find(&departments)
	fmt.Printf("\nDepartments: %d found\n", len(departments))
	for _, d := range departments {
		fmt.Printf("  - ID: %d, Code: %s, Name: %s\n", d.ID, d.Code, d.Name)
	}

	// Check users
	var users []User
	db.Find(&users)
	fmt.Printf("\nUsers: %d found\n", len(users))
	for _, u := range users {
		fmt.Printf("  - ID: %d, Email: %s, Role: %s\n", u.ID, u.Email, u.Role)
	}

	if len(universities) == 0 {
		fmt.Println("\n❌ Database is empty! Seeder needs to run.")
	} else {
		fmt.Println("\n✓ Database has data.")
	}
}
