package db

import (
	"database/sql"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Storage struct {
	db *gorm.DB
}

func NewStorage(conf *StorageConfig) (*Storage, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", conf.user, conf.password, conf.ip, conf.port, conf.scheme)
	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn: sqlDB,
	}))
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&main{}, &pbe{}, &mode{})
	if err != nil {
		panic("failed to migrate database")
	}

	return &Storage{
		db: db,
	}, nil
}

type StorageConfig struct {
	user     string
	password string
	ip       string
	port     string
	scheme   string
}

func NewStorageConfig(user string, password string, ip string, port string, scheme string) *StorageConfig {
	return &StorageConfig{
		user:     user,
		password: password,
		ip:       ip,
		port:     port,
		scheme:   scheme,
	}
}

func (s Storage) SaveMain(name string) error {

	var cnt int64
	s.db.Model(&main{}).Where("name = ?", name).Count(&cnt)

	if cnt == 0 {
		dec := main{
			Name: name,
		}
		result := s.db.Create(&dec)
		if result.Error != nil {
			return result.Error
		}
	}

	return nil
}

func (s Storage) SavePbe(name string) error {

	var cnt int64
	s.db.Model(&pbe{}).Where("name = ?", name).Count(&cnt)

	if cnt == 0 {
		dec := pbe{
			Name: name,
		}
		result := s.db.Create(&dec)
		if result.Error != nil {
			return result.Error
		}
	}
	return nil
}

func (s Storage) DeleteAllMain() error {
	result := s.db.Unscoped().Where("1 = 1").Delete(&main{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s Storage) DeleteAllPbe() error {
	result := s.db.Unscoped().Where("1 = 1").Delete(&pbe{}) // memo. Unscopred : deleted_at으로 관리되던 삭제 여부 무시하고 수행. (delete면 싹 다 삭제. select면 deleted_at 되어있어도 조회)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s Storage) DeleteMain(name string) error {
	result := s.db.Where("name = ?", name).Delete(&main{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s Storage) DeletePbe(name string) error {
	result := s.db.Where("name = ?", name).Delete(&pbe{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s Storage) AllMain() ([]string, error) {

	var mains []main

	result := s.db.Model(&main{}).Select("name").Find(&mains)
	if result.Error != nil {
		return nil, result.Error
	}

	decs := make([]string, len(mains))
	for i := 0; i < len(mains); i++ {
		decs[i] = mains[i].Name
	}
	return decs, nil
}

func (s Storage) AllPbe() ([]string, error) {

	var pbes []pbe

	result := s.db.Model(&pbe{}).Select("name").Find(&pbes)
	if result.Error != nil {
		return nil, result.Error
	}

	decs := make([]string, len(pbes))
	for i := 0; i < len(pbes); i++ {
		decs[i] = pbes[i].Name
	}
	return decs, nil
}

func (s Storage) Mode() bool {
	m := mode{}
	s.db.Model(&mode{}).Last(&m)
	return !m.IsPbe // default 값을 false로 하기 위해 main.go에서의 변수명과 반대로 저장
}

func (s Storage) SaveMode(isMain bool) {
	m := mode{}
	s.db.Last(&m)
	m.IsPbe = !isMain
	if m.ID == 0 {
		s.db.Model(&mode{}).Create(&m)
	} else {
		s.db.Updates(m)
	}
}
