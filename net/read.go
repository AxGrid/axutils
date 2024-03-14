package net

import (
	"fmt"
	"io"
	"net"
	"time"
)

/*
 __    _           ___
|  |  |_|_____ ___|_  |
|  |__| |     | .'|  _|
|_____|_|_|_|_|__,|___|
zed (14.03.2024)
*/

func ReadNBytes(conn net.Conn, n int) ([]byte, error) {
	buff := make([]byte, n)
	nRead, err := io.ReadFull(conn, buff)
	if err != nil {
		return nil, err
	}
	if nRead != n {
		return buff, fmt.Errorf("failed to read packet in full")
	}

	return buff, nil
}

func ReadNBytesWithTimeout(conn net.Conn, n int, timeout time.Duration) ([]byte, error) {
	err := conn.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		return nil, err
	}
	buff := make([]byte, n)
	nRead, err := io.ReadFull(conn, buff)
	if err != nil {
		return nil, err
	}
	if nRead != n {
		return buff, fmt.Errorf("failed to read packet in full")
	}
	return buff, nil
}
