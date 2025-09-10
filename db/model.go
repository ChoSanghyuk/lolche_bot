package db

import "gorm.io/gorm"

type main struct {
	ID   uint
	Name string
	gorm.Model
}

type pbe struct {
	ID   uint
	Name string
	gorm.Model
}

type mode struct {
	ID     uint
	IsMain bool
}
