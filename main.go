package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

const (
	webDir        = "./web/"
	dateFormat    = "20060102"
	maxDays       = 400
	getTasksLimit = 10
)

var (
	serverPort string = "7540"
	dbFileName string = "scheduler.db"
)

var schedulerService scheduler

func main() {

	var err error

	// проверим переменные окружения и перезапишем глобальные переменные
	if envPort, ok := os.LookupEnv("TODO_PORT"); ok {
		serverPort = envPort
	}

	if envDbFile, ok := os.LookupEnv("TODO_DBFILE"); ok {
		dbFileName = envDbFile
	}

	db, err := dbConnect(dbFileName)
	if err != nil {
		log.Fatal("unable to connect database: ", err)
	}
	defer db.Close()

	// создаем экземпляр Планировщика
	schedulerService = NewScheduler(db)

	// устанавливаем обработчики и запускаем сервер
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(webDir)))
	mux.HandleFunc("/api/nextdate", handleNextDate)
	mux.HandleFunc("/api/task", handleTask)
	mux.HandleFunc("/api/tasks", handleTasks)
	mux.HandleFunc("/api/task/done", handleTaskDone)

	log.Println("Starting HTTP Server...")
	err = http.ListenAndServe(fmt.Sprintf(":%s", serverPort), mux)
	if err != nil {
		panic(err)
	}

}
