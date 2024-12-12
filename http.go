package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
		respBody, err = json.Marshal(respOkErr{Error: err.Error()})
	} else {
		respBody, err = json.Marshal(respOkErr{Id: strconv.Itoa(id)})
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(respBody)
}

func handleTaskGET(w http.ResponseWriter, r *http.Request) {
	var respBody []byte
	var task Task
	var err error

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		respBody, err = json.Marshal(respOkErr{Error: err.Error()})
	} else {
		task, err = schedulerService.getTask(id)
		if err != nil {
			respBody, err = json.Marshal(respOkErr{Error: err.Error()})
		} else {
			respBody, err = json.Marshal(task)
		}
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(respBody)
}

func handleTaskPUT(w http.ResponseWriter, r *http.Request) {
	var respBody []byte
	var err error
	var resp respOkErr

	err = updateTask(r.Body)
	if err != nil {
		resp.Error = err.Error()
	}

	respBody, err = json.Marshal(resp)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(respBody)
}

func handleTaskDone(w http.ResponseWriter, r *http.Request) {

	var respBody []byte
	var err error
	var resp respOkErr

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		resp.Error = err.Error()
	} else {
		err = schedulerService.doneTask(id)
		if err != nil {
			resp.Error = err.Error()
		}
	}

	respBody, err = json.Marshal(resp)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(respBody)
}

func handleTaskDELETE(w http.ResponseWriter, r *http.Request) {
	var respBody []byte
	var err error
	var resp respOkErr

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		resp.Error = err.Error()
	} else {
		err = schedulerService.deleteTask(id)
		if err != nil {
			resp.Error = err.Error()
		}
	}

	respBody, err = json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(respBody)
}

func handleTasks(w http.ResponseWriter, r *http.Request) {

	var respBody []byte
	var err error

	tasks, err := schedulerService.getTasks()
	if err != nil {
		respBody, err = json.Marshal(respOkErr{Error: err.Error()})
	} else {
		respBody, err = json.Marshal(respTasks{tasks})
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(respBody)
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
