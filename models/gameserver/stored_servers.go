package gameserver

import (
	"errors"
	"sync"

	db "github.com/TF2Stadium/Helen/database"
)

type StoredServer struct {
	ID   uint   `gorm:"primary_key" json:"id"`
	Name string `json:"name"`

	Address      string `json:"-" sql:"unique"`
	RCONPassword string `json:"-"`
	Used         bool   `sql:"default:false" json:"-"`
}

var (
	ErrServerUsed          = errors.New("server is being used")
	ErrServerAlreadyExists = errors.New("server already exists")
)

func NewStoredServer(name, address, passwd string) (*StoredServer, error) {
	var count int
	db.DB.Model(&StoredServer{}).Where("address = ?", address).Count(&count)
	if count != 0 {
		return nil, ErrServerAlreadyExists
	}

	server := &StoredServer{
		Name:         name,
		Address:      address,
		RCONPassword: passwd,
	}

	db.DB.Save(server)
	return server, nil
}

func RemoveStoredServer(addr string) {
	db.DB.Model(&StoredServer{}).Where("address = ?", addr).Delete(&StoredServer{})
}

func GetAvailableServers() []*StoredServer {
	var servers []*StoredServer
	db.DB.Model(&StoredServer{}).Where("used = FALSE").Find(&servers)
	return servers
}

var storeLock = new(sync.Mutex)

func GetStoredServer(id uint) (*StoredServer, error) {
	storeLock.Lock()
	defer storeLock.Unlock()

	server := &StoredServer{}
	err := db.DB.Model(&StoredServer{}).Where("id = ?", id).First(server).Error
	if server.Used {
		return nil, ErrServerUsed
	}

	db.DB.First(server).UpdateColumn("used", true)
	return server, err
}

func PutStoredServer(address string) {
	storeLock.Lock()
	db.DB.Model(&StoredServer{}).Where("address = ?", address).UpdateColumn("used", false)
	storeLock.Unlock()
}

func GetAllStoredServers() []*StoredServer {
	var servers []*StoredServer
	db.DB.Model(&StoredServer{}).Find(&servers)
	return servers
}
