package main

import (
	"database/sql"
	"errors"
)

// ParcelStore отвечает за работу с БД: добавление, обновление, получение, удаление посылок.
type ParcelStore struct {
	db *sql.DB
}

// NewParcelStore возвращает новый экземпляр ParcelStore с переданным подключением к БД.
func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

// Add добавляет новую посылку в таблицу parcel.
// Возвращает сгенерированный номер (ID) новой посылки.
func (s ParcelStore) Add(p Parcel) (int, error) {
	// Выполняем SQL-запрос на добавление новой строки.
	res, err := s.db.Exec(
		"INSERT INTO parcel (client, status, address, created_at) VALUES (?, ?, ?, ?)",
		p.Client, p.Status, p.Address, p.CreatedAt,
	)
	if err != nil {
		return 0, err
	}

	// Получаем автоинкрементный номер (ID) последней вставленной записи.
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Приводим к типу int и возвращаем.
	return int(id), nil
}

// Get возвращает посылку по её номеру (primary key).
func (s ParcelStore) Get(number int) (Parcel, error) {
	// Выполняем SQL-запрос, который вернёт одну строку.
	row := s.db.QueryRow(
		"SELECT number, client, status, address, created_at FROM parcel WHERE number = ?",
		number,
	)

	// Сканируем строку в структуру Parcel.
	var p Parcel
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		return Parcel{}, err
	}

	return p, nil
}

// GetByClient возвращает все посылки, принадлежащие конкретному клиенту.
func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// Выполняем SQL-запрос, возвращающий все строки по заданному client.
	rows, err := s.db.Query(
		"SELECT number, client, status, address, created_at FROM parcel WHERE client = ?",
		client,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Parcel

	// Сканируем каждую строку в структуру Parcel и добавляем в срез.
	for rows.Next() {
		var p Parcel
		err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		res = append(res, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

// SetStatus обновляет статус посылки по её номеру.
func (s ParcelStore) SetStatus(number int, status string) error {
	_, err := s.db.Exec(
		"UPDATE parcel SET status = ? WHERE number = ?",
		status, number,
	)
	return err
}

// SetAddress меняет адрес доставки посылки, если она ещё не отправлена (т.е. статус = "registered").
func (s ParcelStore) SetAddress(number int, address string) error {
	// Пытаемся обновить адрес только если статус "registered"
	res, err := s.db.Exec(
		"UPDATE parcel SET address = ? WHERE number = ? AND status = ?",
		address, number, ParcelStatusRegistered,
	)
	if err != nil {
		return err // ошибка при выполнении запроса
	}

	// Проверка обновления хотя бы одной строки
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err // ошибка при получении количества обновлённых строк
	}
	if rowsAffected == 0 {
		// либо не существует посылки с таким номером,
		// либо она уже не в статусе "registered"
		return errors.New("can change address only for registered parcels")
	}

	return nil
}

// Delete удаляет посылку, если она ещё не была отправлена (т.е. статус = "registered").
func (s ParcelStore) Delete(number int) error {
	// Пытаемся удалить посылку, только если она в статусе "registered"
	res, err := s.db.Exec(
		"DELETE FROM parcel WHERE number = ? AND status = ?",
		number, ParcelStatusRegistered,
	)
	if err != nil {
		return err // ошибка при выполнении запроса
	}

	// действительно ли что-то было удалено?
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err // ошибка при получении количества затронутых строк
	}
	if rowsAffected == 0 {
		// либо посылки не существует,
		// либо она не подлежит удалению
		return errors.New("can delete only registered parcels")
	}

	return nil
}
