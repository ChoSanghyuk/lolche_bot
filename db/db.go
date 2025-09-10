package db

import (
	"database/sql"
	"fmt"
	"lolcheBot"

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

	if !db.Migrator().HasTable("modes") {
		err = db.AutoMigrate(&main{}, &pbe{}, &mode{})
		if err != nil {
			panic("failed to migrate database")
		}
		db.Model(&mode{}).Create(&mode{ // default 값은 메인모드.
			IsMain: true,
		})

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

func (s Storage) Save(mode lolcheBot.Mode, name string) error {
	if mode == lolcheBot.MainMode {
		return s.saveMain(name)
	} else {
		return s.savePbe(name)
	}
}

func (s Storage) saveMain(name string) error {

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

func (s Storage) savePbe(name string) error {

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

func (s Storage) DeleteAll(mode lolcheBot.Mode) error {

	if mode == lolcheBot.MainMode {
		return s.deleteAllMain()
	} else {
		return s.deleteAllPbe()
	}
}

func (s Storage) deleteAllMain() error {
	result := s.db.Unscoped().Where("1 = 1").Delete(&main{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s Storage) deleteAllPbe() error {
	result := s.db.Unscoped().Where("1 = 1").Delete(&pbe{}) // memo. Unscopred : deleted_at으로 관리되던 삭제 여부 무시하고 수행. (delete면 싹 다 삭제. select면 deleted_at 되어있어도 조회)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s Storage) DeleteByName(mode lolcheBot.Mode, name string) error {
	if mode == lolcheBot.MainMode {
		return s.deleteMainByName(name)
	} else {
		return s.deletePbeByName(name)
	}
}

func (s Storage) deleteMainByName(name string) error {
	result := s.db.Where("name = ?", name).Delete(&main{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s Storage) deletePbeByName(name string) error {
	result := s.db.Where("name = ?", name).Delete(&pbe{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s Storage) All(mode lolcheBot.Mode) ([]string, error) {
	if mode == lolcheBot.MainMode {
		return s.allMain()
	} else {
		return s.allPbe()
	}
}

func (s Storage) allMain() ([]string, error) {

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

func (s Storage) allPbe() ([]string, error) {

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

func (s Storage) Mode() lolcheBot.Mode {
	m := mode{}
	s.db.Model(&mode{}).Last(&m)
	return lolcheBot.Mode(m.IsMain) // default 값을 false로 하기 위해 main.go에서의 변수명과 반대로 저장
}

func (s Storage) SaveMode(currentMode lolcheBot.Mode) {

	m := mode{}
	s.db.Last(&m)
	m.IsMain = bool(currentMode)
	if m.ID == 0 {
		s.db.Model(&mode{}).Create(&m)
	} else {
		s.db.Select("*").Updates(&m)
	}

}
