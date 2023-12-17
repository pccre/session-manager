package main

import (
	"log"
	"strings"
	"sync"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/earlydata"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/pccre/utils/Mut"
	"github.com/pccre/utils/c"
)

var wsConfig = websocket.Config{EnableCompression: true}
var json = c.JSON

var methodsList string
var pool = Mut.Map[string, *Mut.WS]{Mut: &sync.RWMutex{}, Map: map[string]*Mut.WS{}}
var subscribed = Mut.Array[*Mut.WS]{Mut: &sync.RWMutex{}}

type MethodHandler func(c *Mut.WS, content interface{})

func unsubscribe(c *Mut.WS) bool {
	for i, conn := range subscribed.Array {
		if conn == c {
			subscribed.Remove(i)
			return true
		}
	}
	return false
}

func disconnect(c *Mut.WS, isKicked bool) {
	if isKicked {
		c.WriteJSON(Response{Method: "StartSession", Content: "session_busy"})
		c.Close()
	}
	unsubscribe(c)
	removeFromPool(c)
}

func removeFromPool(c *Mut.WS) bool {
	for i, conn := range pool.Map {
		if conn == c {
			pool.Remove(i)
			return true
		}
	}
	return false
}

func BroadcastJSON(content interface{}) {
	if len(subscribed.Array) > 0 {
		data, _ := json.Marshal(content)
		subscribed.Mut.RLock()
		for _, c := range subscribed.Array {
			c.WriteRaw(data)
		}
		subscribed.Mut.RUnlock()
	}
}

func makeSessionCount() Response {
	return Response{Method: "sessionsCount", Content: map[string]int{"sessionCount": len(pool.Map)}}
}

var methods = map[string]MethodHandler{
	"startsession": func(c *Mut.WS, content interface{}) {
		msg, ok := content.(string)
		if !ok {
			c.WriteJSON(Response{Method: "StartSession", Content: "[ERROR] You need to provide a string as the argument!"})
			return
		}

		conn, ok := pool.GetValueAndState(msg)
		if ok {
			if conn == c {
				c.WriteJSON(Response{Method: "StartSession", Content: "session_already_started"})
				return
			} else {
				pool.Set(msg, c)
				disconnect(conn, true)
			}
		} else {
			pool.Set(msg, c)
			BroadcastJSON(makeSessionCount())
		}
		c.WriteJSON(Response{Method: "StartSession", Content: "session_started"})
	},
	"subscribe": func(c *Mut.WS, _ interface{}) {
		subscribed.Mut.RLock()
		for _, conn := range subscribed.Array {
			if c == conn {
				c.WriteJSON(Response{Method: "Subscribe", Content: "[ERROR] You are already subscribed to updates!"})
				return
			}
		}
		subscribed.Mut.RUnlock()

		subscribed.Append(c)
		c.WriteJSON(makeSessionCount())
	},
	"unsubscribe": func(c *Mut.WS, _ interface{}) {
		if !unsubscribe(c) {
			c.WriteJSON(Response{Method: "Unsubscribe", Content: "[ERROR] You are not subscribed to updates!"})
		}
	},
}

func main() {
	go func() {
		for method := range methods {
			methodsList += method + ", "
		}
		methodsList = methodsList[:len(methodsList)-2]
	}()
	http := fiber.New(fiber.Config{
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
		GETOnly:     true,
	})

	http.Use(recover.New())
	http.Use(earlydata.New())

	http.Use("/SessionManager", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	http.Get("/SessionManager", websocket.New(func(cn *websocket.Conn) {
		c := &Mut.WS{WS: cn, Mut: &sync.Mutex{}, Closer: make(chan int)}
		var (
			msg    []byte
			parsed map[string]interface{}
			method string
			err    error
			found  bool
		)

		go func() {
			for {
				if _, msg, err = c.WS.ReadMessage(); err != nil {
					log.Println("read err:", err)
					disconnect(c, false)
					BroadcastJSON(makeSessionCount())
					return
				}

				err = json.Unmarshal(msg, &parsed)
				if err != nil {
					goto invalidContent
				}

				method, found = parsed["method"].(string)
				if !found {
					goto invalidContent
				}

				found = false
				method = strings.ToLower(method)
				for meth, handler := range methods {
					if meth == method {
						go handler(c, parsed["args"])
						found = true
						break
					}
				}

				if !found {
					c.WriteJSON(Response{Method: "OnMessage", Content: "[ERROR] Invalid method! Method list: " + methodsList})
				}
				continue
			invalidContent:
				c.WriteJSON(Response{Method: "OnMessage", Content: `[ERROR] Invalid content! You must pass JSON like this: {"method": "methodName", "args": "arguments (optional)"}`})
			}
		}()

		if <-c.Closer == 1 { // https://github.com/gofiber/contrib/issues/698
			return
		}
	}, wsConfig))

	log.Fatal(http.Listen(":8082"))
}
