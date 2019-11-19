package rds

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"log"
	"time"
)

type Redis interface {
	Action(fn func(redis.Conn) error) error
	Mutex(key string) (func(), error)
	Get(readonly...bool) redis.Conn
}

func NewClient(client *Client) Redis {
	ln := len(client.Addr)
	if ln == 1 {
		client := &redis.Pool{
			MaxIdle:   80,
			MaxActive: 12000,
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", client.Addr[0])
				if err != nil {
					panic(err.Error())
				}
				return c, err
			},
		}
		return &Client{
			readPool:  client,
			writePool: client,
		}
	}
	if ln >= 2 {
		i := 0
		mx := ln - 1
		return &Client{
			writePool: &redis.Pool{
				MaxIdle:   80,
				MaxActive: 12000,
				Dial: func() (redis.Conn, error) {
					c, err := redis.Dial("tcp", client.Addr[0])
					if err != nil {
						panic(err)
					}
					return c, err
				},
			},
			readPool: &redis.Pool{
				MaxIdle:   80,
				MaxActive: 12000,
				Dial: func() (redis.Conn, error) {
					if i > mx-1 {
						i = 0
					}
					c, err := redis.Dial("tcp", client.Addr[1+i])
					if err != nil {
						panic(err)
					}
					i++
					return c, err
				},
			},
		}
	}
	log.Fatal("redis configuration failed")
	return nil
}

type Client struct {
	Addr      []string
	readPool  *redis.Pool
	writePool *redis.Pool
}

func (p *Client) Get(readonly...bool) redis.Conn {
	if len(readonly) > 0 {
		return p.readPool.Get()
	}
	return p.writePool.Get()
}

func (p *Client) Action(fn func(redis.Conn) error) (err error) {
	c := p.Get()
	defer func() {
		if p := recover(); p != nil {
			c.Close()
			panic(p)
		} else {
			c.Close()
		}
	}()
	err = fn(c)
	return err
}

func (p *Client) Mutex(key string) (func(), error) {
	c := p.Get()
	data, _ := redis.Bool(c.Do("GET", key))
	if data {
		ch := make(chan bool)
		go func() {
			for i := 0; i < 10; i++ {
				time.Sleep(time.Millisecond * 100)
				data, _ := redis.Bool(c.Do("GET", key))
				if !data {
					ch <- true
					return
				}
			}
			ch <- false
		}()
		data := <-ch
		if !data {
			return nil, fmt.Errorf("failed to acquire lock")
		}
	}
	_ = c.Send("SET", key, true)
	return func() {
		_ = c.Send("SET", key, false)
	}, nil
}
