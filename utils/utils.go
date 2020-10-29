package utils

import (
	"encoding/binary"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"time"
)

type IPv4Int uint32

var info *log.Logger

func init() {
	file, err := os.OpenFile("./log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open error log file:", err)
	}
	info = log.New(io.MultiWriter(file, os.Stderr), "INFO: ", log.Ldate|log.Ltime)
}

func Log(v ...interface{}) {
	info.Println(v)
}

func RandomIpv4Int() uint32 {
	return rand.New(rand.NewSource(time.Now().UnixNano())).Uint32()
}

func IsExist(f string) bool {
	_, err := os.Stat(f)
	return err == nil || os.IsExist(err)
}

func GetFileSize(filename string) int64 {
	var result int64
	filepath.Walk(filename, func(path string, f os.FileInfo, err error) error {
		result = f.Size()
		return nil
	})
	return result
}

func RandInt64(min, max int64) int64 {
	if min >= max || min == 0 || max == 0 {
		return max
	}
	return rand.Int63n(max-min) + min
}

func (i IPv4Int) ip() net.IP {
	ip := make(net.IP, net.IPv6len)
	copy(ip, net.IPv4zero)
	binary.BigEndian.PutUint32(ip.To4(), uint32(i))
	return ip.To16()
}
