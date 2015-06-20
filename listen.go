// Package golisten allows a user to user http.ListenAndServe with
// any port as root and effectively accept the incomming connection as an other
// un-privileged user.
package golisten

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"
)

var listenFD = flag.Int("listen-fd", 0, "open fd for listener")

func lookupUser(username string) (uid, gid int, err error) {
	u, err := user.Lookup(username)
	if err != nil {
		return -1, -1, err
	}
	uid, err = strconv.Atoi(u.Uid)
	if err != nil {
		return -1, -1, err
	}
	gid, err = strconv.Atoi(u.Gid)
	if err != nil {
		return -1, -1, err
	}
	return uid, gid, nil
}

// ListenAndServe wraps `http.ListenAndServe`. Listen as root and accept as `targetUser`.
func ListenAndServe(targetUser, addr string, handler http.Handler) error {
	if !flag.Parsed() {
		flag.Parse()
	}
	u, err := user.Current()
	if err != nil {
		return err
	}
	if u.Uid != "0" && *listenFD == 0 {
		// we are not root and we have no listen fd. Error.
		return fmt.Errorf("need to run as root. Running as %s (%s)", u.Name, u.Uid)
	} else if u.Uid == "0" && *listenFD == 0 {
		// we are root and we have no listen fd. Do the listen.
		l, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("Listen error: %s", err)

		}
		f, err := l.(*net.TCPListener).File()
		if err != nil {
			return err
		}

		uid, gid, err := lookupUser(targetUser)
		if err != nil {
			return err
		}
		// First extra file: fd == 3
		cmd := exec.Command(os.Args[0])
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.ExtraFiles = append(cmd.ExtraFiles, f)
		cmd.Args = append(cmd.Args, []string{"-listen-fd", fmt.Sprint(2 + len(cmd.ExtraFiles))}...)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Credential: &syscall.Credential{
				Uid: uint32(uid),
				Gid: uint32(gid),
			},
		}
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("cmd.Run error: %s", err)
		}
		return nil
	} else if u.Uid != "0" && *listenFD != 0 {
		// We are not root and we have a listen fd. Do the accept.
		l, err := net.FileListener(os.NewFile(uintptr(*listenFD), "net"))
		if err != nil {
			return err
		}

		if err := http.Serve(l, handler); err != nil {
			return err
		}
	}
	return fmt.Errorf("setuid fail: %s, %d", u.Uid, *listenFD)
}

// Listen listens to the requested port and de-escalate the privilege to `targetUser`.
func Listen(targetUser, network, addr string) (net.Listener, error) {
	if !flag.Parsed() {
		flag.Parse()
	}
	u, err := user.Current()
	if err != nil {
		return nil, err
	}
	if u.Uid != "0" && *listenFD == 0 {
		// we are not root and we have no listen fd. Error.
		return nil, fmt.Errorf("need to run as root. Running as %s (%s)", u.Name, u.Uid)
	} else if u.Uid == "0" && *listenFD == 0 {
		// we are root and we have no listen fd. Do the listen.
		ln, err := net.Listen(network, addr)
		if err != nil {
			return nil, fmt.Errorf("Listen error: %s", err)

		}
		var f *os.File
		switch ln := ln.(type) {
		case *net.TCPListener:
			f, err = ln.File()
		case *net.UnixListener:
			f, err = ln.File()
		default:
			err = fmt.Errorf("unsupported network: %T", ln)
		}
		if err != nil {
			return nil, err
		}

		uid, gid, err := lookupUser(targetUser)
		if err != nil {
			return nil, err
		}
		// First extra file: fd == 3
		cmd := exec.Command(os.Args[0])
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.ExtraFiles = append(cmd.ExtraFiles, f)
		cmd.Args = append(cmd.Args, []string{"-listen-fd", fmt.Sprint(2 + len(cmd.ExtraFiles))}...)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Credential: &syscall.Credential{
				Uid: uint32(uid),
				Gid: uint32(gid),
			},
		}
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("cmd.Run error: %s", err)
		}
		return nil, nil
	} else if u.Uid != "0" && *listenFD != 0 {
		// We are not root and we have a listen fd. Do the accept.
		ln, err := net.FileListener(os.NewFile(uintptr(*listenFD), "net"))
		if err != nil {
			return nil, err
		}
		return ln, nil
	}
	return nil, fmt.Errorf("setuid fail: %s, %d", u.Uid, *listenFD)
}
