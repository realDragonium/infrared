package api

import (
	"github.com/haveachin/infrared/api/service"
)

type Config struct {
	Address string
	MySQL   service.MySQLConfig
}
