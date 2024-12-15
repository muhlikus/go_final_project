package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

type respTasks struct {
	Tasks []Task `json:"tasks"`
}

type respOkErr struct {
	Id    string `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

// handlers
func handleNextDate(w http.ResponseWriter, r *http.Request) {

	reqNow := r.FormValue("now")
	reqDate := r.FormValue("date")
	reqRepeat := r.FormValue("repeat")

	now, err := parseDate(reqNow)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	nextDate, err := NextDate(now, reqDate, reqRepeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprint(w, nextDate)
}

func handleTask(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodPost:
		handleTaskPOST(w, r)
	case http.MethodGet:
		handleTaskGET(w, r)
	case http.MethodPut:
		handleTaskPUT(w, r)
	case http.MethodDelete:
		handleTaskDELETE(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleTaskPOST(w http.ResponseWriter, r *http.Request) {

	var respBody []byte
	var err error

	id, err := addTask(r.Body)
	if err != nil {
		writeBadRequest(w, err)
		return
	}

	respBody, err = json.Marshal(respOkErr{Id: strconv.Itoa(id)})
	if err != nil {
		writeErrJSONEncoding(w, err)
		return
	}

	writeResponse(w, respBody)
}

func handleTaskGET(w http.ResponseWriter, r *http.Request) {
	var respBody []byte
	var task Task
	var err error

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		writeBadRequest(w, err)
		return
	}

	task, err = schedulerService.getTask(id)
	if err != nil {
		writeBadRequest(w, err)
		return
	}

	respBody, err = json.Marshal(task)
	if err != nil {
		writeErrJSONEncoding(w, err)
		return
	}

	writeResponse(w, respBody)
}

func handleTaskPUT(w http.ResponseWriter, r *http.Request) {
	var respBody []byte
	var err error
	var resp respOkErr

	err = updateTask(r.Body)
	if err != nil {
		writeBadRequest(w, err)
		return
	}

	respBody, err = json.Marshal(resp)
	if err != nil {
		writeErrJSONEncoding(w, err)
		return
	}

	writeResponse(w, respBody)
}

func handleTaskDone(w http.ResponseWriter, r *http.Request) {

	var respBody []byte
	var err error

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		writeBadRequest(w, err)
		return
	}

	err = schedulerService.doneTask(id)
	if err != nil {
		writeBadRequest(w, err)
		return
	}

	respBody, err = json.Marshal(respOkErr{})
	if err != nil {
		writeErrJSONEncoding(w, err)
		return
	}

	writeResponse(w, respBody)
}

func handleTaskDELETE(w http.ResponseWriter, r *http.Request) {
	var respBody []byte
	var err error

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		writeBadRequest(w, err)
		return
	}

	err = schedulerService.deleteTask(id)
	if err != nil {
		writeBadRequest(w, err)
		return
	}

	respBody, err = json.Marshal(respOkErr{})
	if err != nil {
		writeErrJSONEncoding(w, err)
		return
	}

	writeResponse(w, respBody)
}

func handleTasks(w http.ResponseWriter, r *http.Request) {

	var respBody []byte
	var err error

	tasks, err := schedulerService.getTasks()
	if err != nil {
		writeBadRequest(w, err)
		return
	}

	respBody, err = json.Marshal(respTasks{tasks})
	if err != nil {
		writeErrJSONEncoding(w, err)
		return
	}

	writeResponse(w, respBody)
}

func writeResponse(w http.ResponseWriter, responseBody []byte) {
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write(responseBody)
	if err != nil {
		log.Println("ERROR sending body to clien", err)
	}
}

func writeBadRequest(w http.ResponseWriter, e error) {
	respBody, err := json.Marshal(respOkErr{Error: e.Error()})
	if err != nil {
		writeErrJSONEncoding(w, err)
		return
	}

	// в общем тут вот странная ситуация, в описании метода http.Error() пишут что
	//
	// The error message should be plain text.
	// и прям в начале реализации этого метода:
	//
	// func Error(w ResponseWriter, error string, code int) {
	// 		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	//
	// то есть предполагается что если мы возвращаем ошибку то она должна быть plain/text
	// у нас в задании вообще нигде не прописано что надо возвращать http код ошибки, а только:
	//
	// При ошибке возвращается JSON-объект с полем error.
	// {"error": "текст ошибки"}
	//
	// поэтому я предполагал что всегда должна быть 200 ОК

	// в общем вот такой код не работает, он всегда вохзвращает 400 Bad Request
	// но Content-Type: text/plain; charset=utf-8
	/*
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, string(respBody), http.StatusBadRequest)
	*/

	// поэтому я написал так, но мне кажется это может противоречить RFC
	// которому предполагаю и следует логика метода http.Error()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	writeResponse(w, respBody)
}

func writeErrJSONEncoding(w http.ResponseWriter, err error) {
	errDescription := "error occured while encoding to JSON"
	http.Error(w, errDescription, http.StatusInternalServerError)
	log.Printf("%s: %v", errDescription, err)

}

func addTask(body io.ReadCloser) (int, error) {

	var task Task
	var buf bytes.Buffer
	var err error // устал подбирать = или :=

	// читаем тело запроса
	_, err = buf.ReadFrom(body)
	if err != nil {
		return 0, err
	}

	// десериализуем JSON
	if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
		return 0, err
	}

	// проверка заполнения поля Title
	if task.Title == "" /* len(task.Title) == 0 */ {
		return 0, errors.New("title cannot be empty")
	}

	// проверка заполнения поля Date
	now := time.Now() // что бы далее в коде не вычислять несколько раз
	nowFormated := now.Format(dateFormat)
	if task.Date != "" {
		_, err = parseDate(task.Date)
		if err != nil {
			return 0, errors.New("date cannot be recognized")
		}
	} else {
		task.Date = nowFormated
	}

	// если дата меньше текущей то или берем текущую если нет правила повторений или вычисляем новую дату согласно правилу
	if task.Date < nowFormated {
		if task.Repeat == "" {
			task.Date = nowFormated
		} else {
			task.Date, err = NextDate(now, task.Date, task.Repeat)
			if err != nil {
				return 0, err
			}
		}
	}

	id, err := schedulerService.addTask(task)
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func updateTask(body io.ReadCloser) error {

	var task Task
	var buf bytes.Buffer
	var err error // устал подбирать = или :=

	// читаем тело запроса
	_, err = buf.ReadFrom(body)
	if err != nil {
		return err
	}

	// десериализуем JSON
	if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
		return err
	}

	// проверка заполнения поля Id
	// еще можно было сразу преобразовать в число и проверить,
	// но мы это сделаем в методе scheduler.updateTask()
	if task.Id == "" /* len(task.Title) == 0 */ {
		return errors.New("ID cannot be empty")
	}

	// проверка заполнения поля Title
	if task.Title == "" /* len(task.Title) == 0 */ {
		return errors.New("title cannot be empty")
	}

	// проверка заполнения поля Date
	now := time.Now() // что бы далее в коде не вычислять несколько раз
	nowFormated := now.Format(dateFormat)
	if task.Date != "" {
		_, err = parseDate(task.Date)
		if err != nil {
			return errors.New("date cannot be recognized")
		}
	} else {
		task.Date = nowFormated
	}

	// если дата меньше текущей то или берем текущую если нет правила повторений или вычисляем новую дату согласно правилу
	if task.Date < nowFormated {
		if task.Repeat == "" {
			task.Date = nowFormated
		} else {
			task.Date, err = NextDate(now, task.Date, task.Repeat)
			if err != nil {
				return err
			}
		}
	}

	err = schedulerService.updateTask(task)
	return err
}
