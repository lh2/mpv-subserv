package main

/*
#cgo LDFLAGS: -Wl,--unresolved-symbols=ignore-all
#include "client.h"
*/
import "C"
import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	running      bool
	server       *http.Server
	listeners    []chan LineMsg
	mux          sync.Mutex
	filters      []*regexp.Regexp
	subs         []LineMsg
	currentSub   int
	currentDelay float64
	currentPos   float64
)

type LineMsg struct {
	Id       int     `json:"id"`
	Line     string  `json:"line"`
	SubStart float64 `json:"sub_start"`
	SubEnd   float64 `json:"sub_end"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func handleConn(conn *websocket.Conn) error {
	c := make(chan LineMsg)
	mux.Lock()
	listeners = append(listeners, c)
	mux.Unlock()
	for {
		var data []byte
		var err error
		msg, ok := <-c
		msgType := websocket.TextMessage
		if !ok {
			msgType = websocket.CloseMessage
		} else {
			data, err = json.Marshal(msg)
		}
		err = conn.WriteMessage(msgType, data)
		if err != nil {
			mux.Lock()
			close(c)
			ci := -1
			for i, c2 := range listeners {
				if c == c2 {
					ci = i
					break
				}
			}
			if ci > -1 {
				listeners[ci] = listeners[len(listeners)-1]
				listeners = listeners[:len(listeners)-1]
			}
			mux.Unlock()
			return err
		}
		if !ok {
			return nil
		}
	}
	return nil
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		w.Header().Add("Content-Type", "text/html")
		w.Write([]byte(INDEX))
	case "/style.css":
		w.Header().Add("Content-Type", "text/css")
		w.Write([]byte(STYLE))
	case "/script.js":
		w.Header().Add("Content-Type", "application/javascript")
		w.Write([]byte(SCRIPT))
	case "/socket":
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		err = handleConn(conn)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
		}
	default:
		w.WriteHeader(404)
		w.Write([]byte("404 file not found"))
	}
}

func startBrowser(url string) error {
	cmdn := os.Getenv("MPV_SUBSERV_BROWSER")
	if cmdn == "" {
		cmdn = os.Getenv("BROWSER")
	}
	if cmdn == "" {
		cmdn = "xdg-open"
	}
	cmd := exec.Command(cmdn, url)
	return cmd.Start()
}

func startServer() {
	listeners = make([]chan LineMsg, 0)
	http.HandleFunc("/", indexHandler)
	tcp, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	if err := startBrowser("http://" + tcp.Addr().String()); err != nil {
		fmt.Fprintf(os.Stderr, "mpv-subserv: could not start browser: %v\n", err)
	}
	server = &http.Server{}
	err = server.Serve(tcp)
	if err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func loadFilters(file string) {
	f, err := os.Open(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mpv-subserv: error opening filter: %v\n", err)
		return
	}
	defer f.Close()

	filters = make([]*regexp.Regexp, 0)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		rxp, err := regexp.Compile(line)
		if err != nil {
			fmt.Fprintf(
				os.Stderr,
				"mpv-subserv: error compiling filter regex '%s': %v\n",
				line,
				err)
			continue
		}
		filters = append(filters, rxp)
	}
	if err = scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "mpv-subserv: error scanning filter: %v\n", err)
	}
}

func parseSubtitleFile() error {
	file := os.Getenv("MPV_SUBSERV_SUBFILE")
	if file == "" {
		return nil
	}
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	var s []LineMsg
	switch strings.ToLower(filepath.Ext(file)) {
	case ".ass":
		s, err = parseAss(f)
	default:
		err = errors.New("unsupported subtitle format")
	}
	if err != nil {
		return err
	}
	subs = make([]LineMsg, 0)
	id := 0
	for _, sub := range s {
		m := false
		for _, filter := range filters {
			if filter.MatchString(sub.Line) {
				m = true
				fmt.Printf("sub %s filtered by regexp %s\n", sub.Line, filter)
				break
			}
		}
		if !m {
			sub.Id = id
			id++
			subs = append(subs, sub)
		}
	}
	return err
}

//export Start
func Start(title *C.char) {
	tpl, err := template.New("index").Parse(INDEX)
	if err != nil {
		log.Fatal(err)
	}
	buf := &bytes.Buffer{}
	err = tpl.Execute(buf, map[string]string{
		"title": C.GoString(title),
		"lang":  os.Getenv("MPV_SUBSERV_LANG"),
	})
	if err != nil {
		log.Fatal(err)
	}
	INDEX = buf.String()

	filterFile := os.Getenv("MPV_SUBSERV_FILTER")
	if filterFile != "" {
		loadFilters(filterFile)
	}

	err = parseSubtitleFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "mpv-subserv: failed parsing subtitle file: %v\n", err)
		return
	}

	go startServer()
	currentSub = -1
	running = true
	fmt.Println("mpv-subserv started")
}

func checkSub() {
	if !running {
		return
	}
	pos := currentPos - currentDelay
	latestStart := -1.0
	var msg LineMsg
	for _, s := range subs {
		if pos >= s.SubStart &&
			pos <= s.SubEnd &&
			s.SubStart > latestStart {
			msg = s
			latestStart = s.SubStart
		}
	}

	if latestStart > -1 && currentSub != msg.Id {
		currentSub = msg.Id
		mux.Lock()
		for _, c := range listeners {
			c <- msg
		}
		mux.Unlock()
	}
}

//export PosChanged
func PosChanged(pos float64) {
	currentPos = pos
	checkSub()
}

//export SubDelayChanged
func SubDelayChanged(delay float64) {
	currentDelay = delay
	checkSub()
}

//export Stop
func Stop() {
	for _, c := range listeners {
		close(c)
	}
	server.Close()
}

func main() {}
