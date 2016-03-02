package models

import (
	db "github.com/TF2Stadium/Helen/database"
)

type StoredServer struct {
	ID           uint   `gorm:"primary_key" json:"id"`
	Name         string `json:"name"`
	Address      string `json:"-"`
	RCONPassword string `json:"-"`
	Used         bool   `sql:"default:false" json:"-"`
}

func GetAvailableServers() []*StoredServer {
	var servers []*StoredServer
	db.DB.Model(&StoredServer{}).Where("used = FALSE").Find(&servers)
	return servers
}

func GetStoredServer(id uint) (*StoredServer, error) {
	server := &StoredServer{}
	err := db.DB.Model(&StoredServer{}).Where("id = ?", id).First(server).Error
	return server, err
}
