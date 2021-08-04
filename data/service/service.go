package service

import (
	"github.com/go-redis/redis"
	"gopkg.in/mgo.v2"
)

type Service struct {
	Session *mgo.Session
}

type RedisService struct {
	RedisClient *redis.Client
}
