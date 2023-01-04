package redis

import (
	"fmt"
	"log"

	"github.com/gomodule/redigo/redis"
)

func ConnectRedis(host string) redis.Conn {
	fmt.Println("Tentative de connexion à REDIS: " + host)
	conn, err := redis.Dial("tcp", host)
	if err != nil {
		log.Fatal(err)
	}
	return conn
}

func DisconnetRedis(conn redis.Conn) {
	conn.Close()
}
