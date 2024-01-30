package taran

import (
	"sync"
	"time"

	"github.com/tarantool/go-tarantool"
	"go.uber.org/zap"
)

const (
	OCRReadyStatus   = "ocr_ready"
	OCRProcessStatus = "ocr_process"
	HighPriority     = "HIGH"
	LowPriority      = "LOW"
)

type Tarantool struct {
	Conn             *tarantool.Connection
	DCCache          string
	ConnectionString string
	Mutex            *sync.Mutex
	Logger           *zap.Logger
}

type TaskStruct struct {
	Id        uint64
	Status    string
	Ttl       int32
	Attempts  int32
	Number    string
	DeviceId  uint64
	Platform  string
	Carrier   string
	Custom    string
	Priority  string
	MessageId string
	Server    string
}

func NewTarantoolClient(connectionString string, login string, password string, logger *zap.Logger) *Tarantool {

	conn, err := tarantool.Connect(connectionString, tarantool.Opts{
		User: login,

		Pass: password,
	})

	tarantool := &Tarantool{
		Conn:             conn,
		ConnectionString: connectionString,
		Mutex:            &sync.Mutex{},
		Logger:           logger,
	}

	if err != nil {
		tarantool.Logger.Error("connection tarantool refused", zap.Error(err))
	} else {
		tarantool.Logger.Info("tarantool 2.0 connection")
	}

	return tarantool
}

func (t *Tarantool) PullTask(space string) (interface{}, error) {

	//Old usage: resp, err := t.Conn.Call("take", []interface{}{space, "s", "transfer"})
	resp, err := t.Conn.Call("takeImageInspector", []interface{}{"DCCache", OCRReadyStatus, OCRProcessStatus, HighPriority})

	// must to log params
	if err != nil {
		return nil, err
	}
	if len(resp.Data) > 0 {
		return resp.Data[0], nil
	}

	resp, err = t.Conn.Call("takeImageInspector", []interface{}{"DCCache", OCRReadyStatus, OCRProcessStatus, LowPriority})

	if err != nil {
		return nil, err
	}
	if len(resp.Data) > 0 {
		return resp.Data[0], nil
	}

	return nil, nil
}

func (t *Tarantool) ResendTask(id uint64) error {
	_, err := t.Conn.Update(t.DCCache, "id", []interface{}{id},
		[]interface{}{[]interface{}{"=", 2, time.Now().Unix()}, []interface{}{"=", 1, "resend"}})
	if err != nil {
		t.Logger.Info("error when resend task", zap.Any("id", id))
		return err
	}
	return nil
}

func (t *Tarantool) UpdateTask(id uint64, update []interface{}) error {
	_, err := t.Conn.Update(t.DCCache, "id", []interface{}{id},
		update)
	if err != nil {
		t.Logger.Info("error when update task", zap.Any("id", id))
		return err
	}
	return nil
}

func (t *Tarantool) DeleteTask(id uint64) bool {
	_, err := t.Conn.Delete(t.DCCache, "id", []interface{}{id})
	if err != nil {
		t.Logger.Info("error when delete task", zap.Any("id", id))
		return false
	}
	return true
}
