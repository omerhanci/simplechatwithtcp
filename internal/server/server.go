package server

import (
	"fmt"
)

func (server *Server) ListClientIDs() []uint64 {
	fmt.Println("TODO: Return the IDs of the connected clients")
	return []uint64{}
}

func (server *Server) Stop() error {
	fmt.Println("TODO: Stop accepting connections and close the existing ones")
	return nil
}
