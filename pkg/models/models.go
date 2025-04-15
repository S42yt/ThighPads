package models

import (
	"time"
)

type Table struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"not null"`
	Author    string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	Entries   []Entry   `gorm:"-"`
}

type Entry struct {
	ID        uint      `gorm:"primaryKey"`
	TableID   uint      `gorm:"not null"`
	Title     string    `gorm:"not null"`
	Tags      string    `gorm:"not null"`
	Content   string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type Config struct {
	Username string `json:"username"`
}
