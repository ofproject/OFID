package whitelist

import (
	"github.com/garyburd/redigo/redis"
	"time"
)

func newPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     Maxidle,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", addr) },
	}
}

var (
	redisDB          *RedisDB = nil
	redisAddr        string   = ""
	redisQueueLength          = 1024
	Maxidle                   = 5
)

type RquestType int8
type ReplyCode int8

const (
	CHECKTYPE = RquestType(iota + 1)
	ADDTYPE
	NUMBERTYPE
	MODIFYTYPE
	DELETETYPE
)

const (
	REPLYOK = ReplyCode(iota)
	REPLYFAIL
)

type msgtype struct {
	msgType string
	addr    string
}

type resulttype struct {
	msg msgtype
	ok  bool
}

type job struct {
	msg []msgtype
	rt  RquestType
	rc  ReplyCode
	r   chan ReplyCode
}

type RedisDB struct {
	jobs chan job
	pool *redis.Pool
	done chan struct{}
}

func NewRedisDB() *RedisDB {
	Logger.Info("redisAddr :", redisAddr)
	if redisDB == nil {
		redisDB = new(RedisDB)
		redisDB.jobs = make(chan job, redisQueueLength)
		redisDB.pool = newPool(redisAddr)
		redisDB.done = make(chan struct{})
	}
	return redisDB
}

func (r *RedisDB) Close() {
	close(r.jobs)
	r.pool.Close()
	close(r.done)

}

func (r *RedisDB) Wait() {
	<-r.done
}

func (r *RedisDB) AddJob(job *job) {
	r.jobs <- *job
}

func (r *RedisDB) Run() {
	for i := 0; i < Maxidle; i++ {
		go r.worker()
	}
}

func (r *RedisDB) worker() {
	for {
		Logger.Info("worker start")
		job, ok := <-r.jobs
		if !ok {
			return
		}
		Logger.Info("worker recv msg: ", job.msg)
		result := make([]resulttype, len(job.msg))

		conn := r.pool.Get()
		for i, v := range job.msg {
			Logger.Info("job.msg.v = ", v.addr)
			switch job.rt {
			case CHECKTYPE:
				{
					result[i].msg = job.msg[i]
					val, err := conn.Do("HGET", "hwhitelist", v.addr)
					if err != nil {
						Logger.Error("worker: HGET command error: ", err)
						job.rc = REPLYFAIL
						break
					} else {
						if val == nil {
							job.rc = REPLYFAIL
						} else {
							result[i].ok = true
							job.rc = REPLYOK
						}
					}
				}
			case ADDTYPE:
				{
					result[i].msg = job.msg[i]
					var exist bool

					reply, err := conn.Do("HEXISTS", "hwhitelist", v.addr)
					if err != nil {
						Logger.Errorf("worker: HEXISTS: %s command error: %T", v.addr, err)
						job.rc = REPLYFAIL
						break
					}

					exist, err = redis.Bool(reply, err)
					if err != nil {
						Logger.Errorf("worker: repley to bool: %s command error: %T", v.addr, err)
						job.rc = REPLYFAIL
						break
					}

					if exist {
						result[i].ok = true
					} else {
						conn.Send("MULTI")
						conn.Send("HSET", "hwhitelist", v.addr, v.msgType)
						conn.Send("SADD", v.msgType, v.addr)
						reply, err := conn.Do("EXEC")
						if err != nil || reply == nil {
							Logger.Errorf("worker: EXEC: %s command error: %T", v.addr, err)
							job.rc = REPLYFAIL
							break
						}

						ok, err := redis.Ints(reply, err)
						if err != nil {
							Logger.Error("worker: EXEC reply to ints fail")
							job.rc = REPLYFAIL
						}

						Logger.Info("worker:Do EXEC reply: ", ok)

						result[i].ok = true
						job.rc = REPLYOK

						err = Inster(v.addr, v.msgType)
						if err != nil {
							Logger.Error("inster to mysql fail err: ", err)
						}
					}

				}
			case NUMBERTYPE:
				{

				}
			case MODIFYTYPE:
				{

				}
			case DELETETYPE:
			default:
				Logger.Error("invalid Requst type: ", job.rt)
			}
		}

		conn.Close()

		if job.r != nil {
			//close(job.r)
			job.r <- job.rc
		}
	}
}
