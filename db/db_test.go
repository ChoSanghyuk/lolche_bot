package db

import (
	"os"
	"testing"
)

func TestMigration(t *testing.T) {

	user := os.Getenv("db_user")
	password := os.Getenv("db_password")

	s, err := NewStorage(&StorageConfig{
		user:     user,
		password: password,
		ip:       "127.0.0.1",
		port:     "3306",
		scheme:   "lolche",
	})
	if err != nil {
		t.Error(err)
		return
	}
	t.Run("mode", func(t *testing.T) {
		mode := s.Mode()
		t.Log(mode.Str())
		s.SaveMode(!mode)
		mode = s.Mode()
		t.Log(mode.Str())
	})
}
