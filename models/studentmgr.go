package models

import (
	"github.com/jinzhu/gorm"
)

type Student struct {
	gorm.Model
	UID           int      `json:"uid"`
	Username      string   `json:"username"`
	Password      string   `json:"password"`
	FullName      string   `json:"full_name"`
	Email         string   `json:"email"`
	HomeDirectory string   `json:"home_directory"`
	Courses       []Course `json:"courses" gorm:"many2many:user_courses;"`
}

type Course struct {
	gorm.Model
	Name            string `json:"name"`
	Description     string `json:"description"`
	Instructor      string `json:"instructor"`
	InstructorEmail string `json:"instructor_email"`
}
