package main

import (
	"log"
	"time"

	"database/sql"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/django/v3"
	_ "github.com/mattn/go-sqlite3"
)

type RegisterForm struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Password2 string `json:"password2"`
}

var channels = make([]*websocket.Conn, 0)
var register = make(chan *websocket.Conn)
var unregister = make(chan *websocket.Conn)

func WebsocketHandler(c *websocket.Conn) {
	defer func() {
		log.Printf("closed Connection from %s\n", c.RemoteAddr().String())
		unregister <- c
		c.Close()
	}()

	log.Printf("new Websocket connection from %s\n", c.RemoteAddr().String())
	register <- c

	for {
		_, message, err := c.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("read error:", err)
			}

			return
		}

		log.Printf("received : %s\n", message)
	}
}

func SocketListen() {
	go func() {
		for {
			select {
			case c := <-register:
				channels = append(channels, c)
			case c := <-unregister:
				for i, con := range channels {
					if con == c {
						channels = append(channels[:i], channels[i+1:]...)
					}
				}
			}
		}
	}()

	for {
		for _, c := range channels {
			err := c.WriteMessage(websocket.TextMessage, []byte("hello"))

			if err != nil {
				log.Println(err)
				unregister <- c
			}
		}

		time.Sleep(time.Second)
	}
}

func main() {
	db, err := InitDb(true)

	app := InitApp()

	CheckError(err)

	go SocketListen()

	app.Use(func(c *fiber.Ctx) error {
		if err := db.Ping(); err != nil {
			panic(err)
		}

		c.Locals("db", db)
		return c.Next()
	})

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("WS-allowed", true)
			return c.Next()
		}

		return fiber.ErrUpgradeRequired
	})

	app.Get("", func(c *fiber.Ctx) error {
		return c.Redirect("/login")
		// return c.Render("index", fiber.Map{
		// 	"title": "Kafka example",
		// })
	})

	// app.Get("/ws", websocket.New(WebsocketHandler))

	app.Get("/login", func(c *fiber.Ctx) error {
		return c.Render("login", fiber.Map{})
	})

	app.Get("/register", func(c *fiber.Ctx) error {
		return c.Render("register", fiber.Map{})
	})

	app.Post("/register", PostRegister)

	err = app.Listen(":3000")

	log.Fatal(err)
}

func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func InitDb(drop bool) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "file:./tmp/db.sqlite3?cache=shared&mode=rwc")

	if err != nil {
		return nil, err
	}

	if drop {
		_, err = db.Exec("DROP TABLE IF EXISTS users;")

		if err != nil {
			return nil, err
		}
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username VARCHAR(255) NOT NULL UNIQUE,
        password TEXT NOT NULL
    );`)

	if err != nil {
		return nil, err
	}

	return db, nil
}

func InitApp() *fiber.App {
	templateEngine := django.New("./views", ".html")
    templateEngine.Reload(true)

	app := fiber.New(fiber.Config{
		ServerHeader:      "Abderrahmane",
		AppName:           "Kafka go example v0.0.1",
		CaseSensitive:     true,
		Views:             templateEngine,
		PassLocalsToViews: true,
		ViewsLayout:       "layouts/base",
	})

	app.Static("/static", "./static/")

	app.Use(logger.New())

	return app
}

func CreateUser(db *sql.DB, data *RegisterForm) {
	stmt, err := db.Prepare("INSERT INTO users (username, password) VALUES (?, ?)")

	if err != nil {
		log.Printf("CreateUser Prepare failed: %s\n", err.Error())
		return
	}

	_, err = stmt.Exec(data.Username, data.Password)

	// log.Println(res.LastInsertId())

	if err != nil {
		log.Printf("CreateUser Exec failed: %s\n", err.Error())
		return
	}
}

func PostRegister(c *fiber.Ctx) error {
	db := c.Locals("db").(*sql.DB)

	data := RegisterForm{}

	err := c.BodyParser(&data)

	if err != nil {
		return err
	}

	CreateUser(db, &data)

	c.WriteString("hello")

	return nil
}
