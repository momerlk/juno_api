package internal 

import (
	"golang.org/x/sys/unix"
	"log"
	"net"
	"reflect"
	"sync"
	"syscall"
)

// represents a single websocket connection
type WSConnection struct {
	NetConn 			net.Conn 			// underlying net connection
	UserId 				string 				// user id of the client
}

// represents a websocket connection writer
type WSWriter struct {
	Conn 				net.Conn
	Lock 				*sync.Mutex
}

type Packet []byte

type Epoll struct {
	fd          int
	connections map[int]*WSConnection
	lock        *sync.RWMutex

	Writers 	map[string]WSWriter
	WriterLock	*sync.Mutex
}

func MkEpoll() (*Epoll, error) {
	fd, err := unix.EpollCreate1(0)
	if err != nil {
		return nil, err
	}
	return &Epoll{
		fd:          fd,
		lock:        &sync.RWMutex{},
		connections: make(map[int]*WSConnection),
		Writers : 	 make(map[string]WSWriter),
	}, nil
}

func (e *Epoll) Add(conn *WSConnection) error {
	// Extract file descriptor associated with the connection
	fd := websocketFD(conn.NetConn)
	err := unix.EpollCtl(e.fd, syscall.EPOLL_CTL_ADD, fd, &unix.EpollEvent{Events: unix.POLLIN | unix.POLLHUP, Fd: int32(fd)})
	if err != nil {
		return err
	}
	e.lock.Lock()
	defer e.lock.Unlock()
	e.connections[fd] = conn
	if true {
		log.Printf("Total number of connections: %v", len(e.connections))
	}
	e.Writers[conn.UserId] = WSWriter{
		Conn: conn.NetConn,
		Lock : &sync.Mutex{},
	}
	return nil
}

func (e *Epoll) Remove(conn *WSConnection) error {
	fd := websocketFD(conn.NetConn)
	err := unix.EpollCtl(e.fd, syscall.EPOLL_CTL_DEL, fd, nil)
	if err != nil {
		return err
	}
	e.lock.Lock()
	defer e.lock.Unlock()
	delete(e.connections, fd)
	if len(e.connections)%100 == 0 {
		log.Printf("Total number of connections: %v", len(e.connections))
	}
	delete(e.Writers , conn.UserId)
	return nil
}

func (e *Epoll) Wait() ([]*WSConnection, error) {
	events := make([]unix.EpollEvent, 100)
	n, err := unix.EpollWait(e.fd, events, 100)
	if err != nil {
		return nil, err
	}
	e.lock.RLock()
	defer e.lock.RUnlock()
	var connections []*WSConnection
	for i := 0; i < n; i++ {
		conn := e.connections[int(events[i].Fd)]
		connections = append(connections, conn)
	}
	return connections, nil
}

func websocketFD(conn net.Conn) int {
	tcpConn := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn")
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")

	return int(pfdVal.FieldByName("Sysfd").Int())
}