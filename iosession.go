package gcbase

import (
	"errors"
	"fmt"
	"net"
)

//Iosession
type Iosession struct {
	id        uint64
	serv      *ioserv
	conn      net.Conn
	closed    bool
	dataCh    chan interface{}
	userId    interface{}
	extraData map[string]interface{}
}

func (s *Iosession) Id() uint64 {
	return s.id
}

func (session *Iosession) SetUserId(id interface{}) {
	session.userId = id
}

func (session *Iosession) GetUserId() interface{} {
	return session.userId
}

func (session *Iosession) ExtraData(key string) (value interface{}, ok bool) {
	value, ok = session.extraData[key]
	return
}

func (session *Iosession) SetExtraData(key string, value interface{}) {
	session.extraData[key] = value
}

func (s *Iosession) Conn() net.Conn {
	return s.conn
}

func (s *Iosession) dealDataCh() {
	/*
		var msg interface{}
		for !s.closed {
			select {
			case msg = <-s.dataCh:
				//fmt.Println("收到消息")

			}
		}*/
}

func (session *Iosession) ReadBytes() ([]byte, int, error) {
	session.serv.wg.Add(1)
	defer session.serv.wg.Done()
	var n int
	var err error
	buffer := make([]byte, 512)
	if session.serv.runnable && !session.closed {
		n, err = session.conn.Read(buffer)
		if err != nil {
			session.Close()
			return nil, 0, errors.New("recv error!")
		}
	}
	return buffer, n, nil

}

func (session *Iosession) readData() {
	session.serv.wg.Add(1)
	ioBuffer := NewBuffer()
	buffer := make([]byte, 512)
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
		if !session.closed {
			session.Close()
		}
		session.serv.wg.Done()
	}()
	var n int
	var err error
	for session.serv.runnable && !session.closed {
		n, err = session.conn.Read(buffer)
		ioBuffer.PutBytes(buffer[:n])
		if err != nil {
			session.serv.filterChain.errorCaught(session, err)
			session.Close()
			return
		}
		err = session.serv.codecer.Decode(ioBuffer, session.dataCh)
		if err != nil {
			session.serv.filterChain.errorCaught(session, err)
		}
	}
}

func (session *Iosession) WriteBytes(msg []byte) error {

	if !session.closed {
		_, err := session.conn.Write(msg)
		if err != nil {
			fmt.Printf("Send error!")
			return err
		}
	}

	//fmt.Printf("msg: %v \n",msg)
	//session.conn.Write([]byte{'h','e','l'})

	return nil
}

func (session *Iosession) Write(message interface{}) error {
	if !session.closed {

		_, err := session.conn.Write(message.([]byte))
		if err != nil {
			fmt.Println("write err:", err)
			return err
		}

		return nil
	} else {
		err := errors.New("Iosession is closed")
		return err
	}
}

//close iosession
func (this *Iosession) Close() {
	if !this.closed {
		this.closed = true
	}
	this.conn.Close()
}
