package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		t.Error("connection error")
	}
	store := NewParcelStore(db)
	parcel := getTestParcel()
	defer db.Close()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	if err != nil {
		t.Error("couldn't add parcel")
	}
	rows, err := db.Query("select number from parcel where number = :number", sql.Named("number", id))
	if err != nil {
		t.Error("no id")
	}
	rows.Close()

	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	rows, err = db.Query("select client, status, address, created_at from parcel where number = :number", sql.Named("number", id))
	if err != nil {
		t.Error("the package is missing")
	}

	var parc Parcel

	for rows.Next() {
		err = rows.Scan(&parc.Client, &parc.Status, &parc.Address, &parc.CreatedAt)
		if err != nil {
			t.Error("copy error")
		}
	}

	assert.Equal(t, parcel.Address, parc.Address)
	assert.Equal(t, parcel.Client, parc.Client)
	assert.Equal(t, parcel.Status, parc.Status)
	assert.Equal(t, parcel.CreatedAt, parc.CreatedAt)
	rows.Close()

	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	err = store.Delete(id)
	if err != nil {
		t.Error("delete error")
	}

	var bl bool

	err = db.QueryRow("select exists(select 1 from parcel where number = :number)", sql.Named("number", id)).Scan(&bl)
	if err != nil {
		t.Error(err)
	}
	if bl == true {
		t.Error("the record has not been deleted")
	}
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		t.Error("connection error")
	}
	store := NewParcelStore(db)
	parcel := getTestParcel()
	defer db.Close()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	if err != nil {
		t.Error("couldn't add parcel")
	}
	rows, err := db.Query("select number from parcel where number = :number", sql.Named("number", id))
	if err != nil {
		t.Error("no id")
	}
	rows.Close()

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	if err != nil {
		t.Error("error updating the address")
	}

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	rows, err = db.Query("select address from parcel where number = :number", sql.Named("number", id))
	if err != nil {
		t.Error("the package is missing")
	}
	defer rows.Close()

	var address string

	for rows.Next() {
		err = rows.Scan(&address)
		if err != nil {
			t.Error("copy error")
		}
	}

	assert.Equal(t, newAddress, address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		t.Error("connection error")
	}
	store := NewParcelStore(db)
	service := NewParcelService(store)
	parcel := getTestParcel()
	defer db.Close()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	if err != nil {
		t.Error("couldn't add parcel")
	}
	rows, err := db.Query("select number from parcel where number = :number", sql.Named("number", id))
	if err != nil {
		t.Error("no id")
	}
	rows.Close()

	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	err = service.NextStatus(id)
	if err != nil {
		t.Error("status update error")
	}

	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	rows, err = db.Query("select status from parcel where number = :number", sql.Named("number", id))
	if err != nil {
		t.Error("the package is missing")
	}

	var stat string

	for rows.Next() {
		err = rows.Scan(&stat)
		if err != nil {
			t.Error("copy error")
		}
	}

	assert.Equal(t, ParcelStatusSent, stat)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		t.Error("connection error")
	}
	store := NewParcelStore(db)
	defer db.Close()

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
		id, err := store.Add(parcels[i])
		if err != nil {
			t.Error("couldn't add parcel")
		}
		rows, err := db.Query("select number from parcel where number = :number", sql.Named("number", id))
		if err != nil {
			t.Error("no id")
		}
		rows.Close()

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client) // получите список посылок по идентификатору клиента, сохранённого в переменной client
	if err != nil {                                 // убедитесь в отсутствии ошибки
		t.Error("couldn't get the list")
	}

	// убедитесь, что количество полученных посылок совпадает с количеством добавленных
	assert.Equal(t, len(storedParcels), len(parcels))

	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		// убедитесь, что значения полей полученных посылок заполнены верно
		expected, exists := parcelMap[parcel.Number]
		if !exists {
			t.Errorf("There is no package with ID - %d", parcel.Number)
		}
		assert.Equal(t, expected.Address, parcel.Address)
		assert.Equal(t, expected.Client, parcel.Client)
		assert.Equal(t, expected.CreatedAt, parcel.CreatedAt)
		assert.Equal(t, expected.Status, parcel.Status)
	}
}
