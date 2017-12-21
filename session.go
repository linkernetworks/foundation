package config

type SessionConfig struct {
	Size     int    `json:"size"`
	Protocal string `json:"protocal"`
	RedisUrl string `json:"redisUrl"`
	Password string `json:"password"`
	Age      int    `json:"age"`
	KeyPair  string `json:"keyPair"`
}
