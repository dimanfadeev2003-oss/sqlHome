package main

import (
	"database/sql"
	"errors"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	res, err := s.db.Exec("insert into parcel (client, status, address, created_at) values (:client, :status, :address, :created_at)",
		sql.Named("client", p.Client),
		sql.Named("status", p.Status),
		sql.Named("address", p.Address),
		sql.Named("created_at", p.CreatedAt))
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	// реализуйте чтение строки по заданному number
	// здесь из таблицы должна вернуться только одна строка
	rows, err := s.db.Query("select status from parcel where number = :number", sql.Named("number", number))
	if err != nil {
		return Parcel{}, err
	}
	defer rows.Close()

	p := Parcel{}

	for rows.Next() {
		// заполните объект Parcel данными из таблицы
		err = rows.Scan(&p.Status)
		if err != nil {
			return Parcel{}, err
		}
	}
	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// реализуйте чтение строк из таблицы parcel по заданному client
	// здесь из таблицы может вернуться несколько строк
	rows, err := s.db.Query("select number, client, status, address, created_at from parcel where client = :client", sql.Named("client", client))
	if err != nil {
		return []Parcel{}, err
	}
	defer rows.Close()
	// заполните срез Parcel данными из таблицы
	var res []Parcel

	for rows.Next() {

		var parcel Parcel

		err = rows.Scan(&parcel.Number, &parcel.Client, &parcel.Status, &parcel.Address, &parcel.CreatedAt)
		if err != nil {
			return []Parcel{}, err
		}
		res = append(res, parcel)
	}
	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	// реализуйте обновление статуса в таблице parcel
	_, err := s.db.Exec("update parcel set status = :status where number = :number",
		sql.Named("status", status),
		sql.Named("number", number))
	if err != nil {
		return err
	}
	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	// реализуйте обновление адреса в таблице parcel
	// менять адрес можно только если значение статуса registered
	rows, err := s.db.Query("select status from parcel where number = :number", sql.Named("number", number))
	if err != nil {
		return err
	}
	defer rows.Close()

	var stat string

	for rows.Next() {
		err = rows.Scan(&stat)
		if err != nil {
			return err
		}
	}
	if stat != ParcelStatusRegistered {
		return errors.New("the value of the not registered status")
	}
	_, err = s.db.Exec("update parcel set address = :address where number = :number",
		sql.Named("address", address),
		sql.Named("number", number))
	if err != nil {
		return err
	}
	return nil
}

func (s ParcelStore) Delete(number int) error {
	// реализуйте удаление строки из таблицы parcel
	// удалять строку можно только если значение статуса registered
	rows, err := s.db.Query("select status from parcel where number = :number", sql.Named("number", number))
	if err != nil {
		return err
	}
	defer rows.Close()

	var stat string

	for rows.Next() {
		err = rows.Scan(&stat)
		if err != nil {
			return err
		}
	}
	if stat != ParcelStatusRegistered {
		return errors.New("the value of the not registered status")
	}

	_, err = s.db.Exec("delete from parcel where number = :number", sql.Named("number", number))
	if err != nil {
		return err
	}
	return nil
}
