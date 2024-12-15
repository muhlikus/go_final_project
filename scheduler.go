package main

import (
	"database/sql"
	"errors"
	"strconv"
	"time"
)

type repeatSettings struct {
	key      string
	interval int
}

type scheduler struct {
	db *sql.DB
}

type Task struct {
	Id      string `json:"id"`
	Date    string `json:"date"`    // "20240201",
	Title   string `json:"title"`   // "Подвести итог",
	Comment string `json:"comment"` // "Мой комментарий",
	Repeat  string `json:"repeat"`  // "d 5"
}

func NewScheduler(db *sql.DB) scheduler {
	return scheduler{db}
}

func (s scheduler) addTask(t Task) (int64, error) {
	res, err := s.db.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (:date, :title, :comment, :repeat)",
		sql.Named("date", t.Date),
		sql.Named("title", t.Title),
		sql.Named("comment", t.Comment),
		sql.Named("repeat", t.Repeat),
	)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (s scheduler) updateTask(t Task) error {
	id, err := strconv.Atoi(t.Id)
	if err != nil {
		return err
	}

	res, err := s.db.Exec("UPDATE scheduler SET date=:date, title=:title, comment=:comment, repeat=:repeat WHERE id=:id",
		sql.Named("id", id),
		sql.Named("date", t.Date),
		sql.Named("title", t.Title),
		sql.Named("comment", t.Comment),
		sql.Named("repeat", t.Repeat),
	)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("задача не найдена")
	}

	return nil
}

func (s scheduler) getTasks() ([]Task, error) {

	var tasks []Task = make([]Task, 0, getTasksLimit)

	rows, err := s.db.Query("SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT :limit", sql.Named("limit", getTasksLimit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		task := Task{}

		err := rows.Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s scheduler) getTask(id int) (Task, error) {

	var task Task

	row := s.db.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = :id", sql.Named("id", id))
	err := row.Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		return task, err
	}

	return task, nil
}

func (s scheduler) doneTask(id int) error {

	var err error

	task, err := s.getTask(id)
	if err != nil {
		return err
	}

	// если не заполнено правило повторения то просто удалим запись
	// иначе получим новую дату и обновим запись
	if task.Repeat == "" {
		err = s.deleteTask(id)
		if err != nil {
			return err
		}
		return nil
	}

	task.Date, err = NextDate(time.Now(), task.Date, task.Repeat)
	if err != nil {
		return err
	}

	err = s.updateTask(task)
	if err != nil {
		return err
	}

	return nil
}

func (s scheduler) deleteTask(id int) error {

	res, err := s.db.Exec("DELETE FROM scheduler WHERE id = :id", sql.Named("id", id))
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("задача не найдена")
	}

	return nil
}
