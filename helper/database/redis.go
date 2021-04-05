package database

import (
	"fmt"
	"os"
	"time"

	"github.com/garyburd/redigo/redis"
)

// RedisPool redis 缓冲池
var RedisPool *redis.Pool

func init() {
	RedisPool = &redis.Pool{
		MaxIdle:     20,
		MaxActive:   1000,
		IdleTimeout: 180,
		Dial: func() (redis.Conn, error) {
			con, err := redis.Dial("tcp",
				"localhost:6379",
				redis.DialPassword("123456"),
				redis.DialConnectTimeout(5*time.Second),
				redis.DialReadTimeout(5*time.Second),
				redis.DialWriteTimeout(5*time.Second))
			if err != nil {
				fmt.Println(err)
				os.Exit(-1)
			}
			return con, nil
		},
	}
}

/*
func main() {
	c := RedisPool.Get()
	defer c.Close()

	_, err := c.Do("set", "abc", 100)
	if err != nil {
		fmt.Println("c.Do err ", err)
		return
	}

	r, err := redis.Int(c.Do("get", "abc"))
	if err != nil {
		fmt.Println("get abc failed", err)
		return
	}
	fmt.Println(r)
	fmt.Printf("r type is %T", r)
	pool.Close()
}
*/
