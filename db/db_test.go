package db

import "testing"

func TestMigration(t *testing.T) {
	s, err := NewStorage()
	if err != nil {
		t.Error(err)
		return
	}
	s.db.AutoMigrate(&main{}, &pbe{}, &mode{})
}
