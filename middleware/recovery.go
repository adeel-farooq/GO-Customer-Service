package middleware

import (
	"log"
	"net"
	"net/rpc"
	"runtime/debug"
)

// ServeConnWithRecovery wraps rpcServer.ServeConn(conn) with a panic recovery.
// A panic in any RPC handler would otherwise crash the whole process.
func ServeConnWithRecovery(rpcServer *rpc.Server, conn net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			remoteAddr := "<nil>"
			localAddr := "<nil>"
			if conn != nil {
				if conn.RemoteAddr() != nil {
					remoteAddr = conn.RemoteAddr().String()
				}
				if conn.LocalAddr() != nil {
					localAddr = conn.LocalAddr().String()
				}
			}

			log.Printf("PANIC recovered in RPC connection handler (remote=%s local=%s): %v\n%s", remoteAddr, localAddr, r, string(debug.Stack()))

			// Best-effort cleanup; ServeConn might not get a chance to close it.
			if conn != nil {
				_ = conn.Close()
			}
		}
	}()

	// Keep existing behavior (ServeConn closes the connection when done).
	rpcServer.ServeConn(conn)
}
