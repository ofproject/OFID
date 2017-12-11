package gomqttv2

import (
	"bytes"
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"os"
	"time"
)

func newPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", addr) },
	}
}

var (
	DB        *DateBase = nil
	RedisAddr string    = ""
)

type dbMessage struct {
	From      string `json:"From"`
	To        string `json:"To"`
	MessageId uint16 `json:"MessageID"`
	Data      []byte `json:"Body"`
}

type DateBase struct {
	pool *redis.Pool
}

func NewDataBase() *DateBase {
	if DB == nil {
		if len(RedisAddr) <= 0 {
			Logger.Fatal("redisAddr is empty")
			os.Exit(1)
		}

		DB = &DateBase{
			pool: newPool(RedisAddr),
		}

	}
	return DB
}

func (d *DateBase) Close() {
	d.pool.Close()
	DB = nil
}

func (d *DateBase) get(clientid string) (*Publish, bool, error) {
	Logger.Debug("start")
	conn := DB.pool.Get()
	defer conn.Close()
	reply, err := conn.Do("LPOP", clientid)
	if err != nil {
		Logger.Errorf("get %s fail from redisDB", clientid)
		return nil, false, err
	}

	if reply == nil {
		return nil, true, nil
	}

	mb, err := redis.Bytes(reply, err)
	if err != nil {
		Logger.Errorf("reply to []byte failed")
		return nil, false, err
	}

	var m dbMessage
	err = json.Unmarshal(mb, &m)
	if err != nil {
		Logger.Errorf("Unmarshal %s fail", mb)
		return nil, false, err
	}

	pub := new(Publish)
	pub.TopicName = m.From
	pub.MessageId = m.MessageId
	pub.Payload = m.Data
	Logger.Debug("end")
	return pub, false, nil
}

func (d *DateBase) save(clientid string, pub *Publish) bool {
	Logger.Debug("start")
	dbmsg := dbMessage{
		From:      clientid,
		To:        pub.TopicName,
		MessageId: pub.MessageId,
		Data:      pub.Payload,
	}

	var buf bytes.Buffer
	buf.Write(pub.Payload)

	val, err := json.Marshal(dbmsg)
	if err != nil {
		Logger.Errorf("Marshal %T to json fail error: %T", dbmsg, err)
		return false
	}
	conn := d.pool.Get()
	defer conn.Close()
	Logger.Debug("marshal : ", string(val))
	reply, err := conn.Do("RPUSH", dbmsg.To, val)
	if err != nil {
		Logger.Errorf("RPUSH data {key: %s, data: %s} to redisdb failed error: %T", dbmsg.To, string(val), err)
		return false
	}

	_, err = redis.Int(reply, err)
	if err != nil {
		Logger.Error("save reply to int fail error: ", err)
		return false
	}

	Logger.Debug("end")

	return true
}
