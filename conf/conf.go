package conf

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type (
	Config struct {
		API     API
		Version uint64
	}
	API struct {
		ListenOnPort       uint64
		CORSAllowedOrigins []string
	}
)

func GetNewConfig() (cfg Config, err error) {
	version, _ := strconv.ParseUint(os.Getenv("VERSION"), 10, 64)
	port, _ := strconv.ParseInt(os.Getenv("LISTEN_PORT"), 10, 64)
	cors := strings.Split(os.Getenv("CORS_ALLOWED"), ",")
	fmt.Println(cors)
	return Config{
		API: API{
			ListenOnPort:       uint64(port),
			CORSAllowedOrigins: cors,
		},
		Version: version,
	}, nil
}
