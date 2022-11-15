package dbrepo

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/PushAndRun/bookings/internal/models"
)

func (m *postgresDBRepo) AllUsers() bool {
	return true
}

//InsertReservation inserts a reservation into the database
func (m *postgresDBRepo) InsertReservation(res models.Reservation) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var newId int

	stmt := `insert into reservations (first_name, last_name, email, phone, start_date, end_date, room_id, created_at, updated_at)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id`

	err := m.DB.QueryRowContext(ctx, stmt,
		res.FirstName,
		res.LastName,
		res.Email,
		res.Phone,
		res.StartDate,
		res.EndDate,
		res.RoomID,
		time.Now(),
		time.Now(),
	).Scan(&newId)

	if err != nil {
		fmt.Sprintln("no row found by scan")
		return 0, err

	}

	return newId, nil
}

//Inserts a restriction into the database
func (m *postgresDBRepo) InsertRoomRestriction(r models.RoomRestriction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `insert into room_restrictions (start_date, end_date, room_id, reservation_id,
		created_at, updated_at, restriction_id)
		values ($1, $2, $3, $4, $5, $6, $7)`

	_, err := m.DB.ExecContext(ctx, stmt,
		r.StartDate,
		r.EndDate,
		r.RoomID,
		r.ReservationID,
		time.Now(),
		time.Now(),
		r.RestrictionID,
	)
	if err != nil {
		return err
	}

	return nil
}

//SearchAvailabilityByDates returns true if availability exists for the given room and dates
func (m *postgresDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var numRows int

	query := `
	
		select count(id)
		from room_restrictions
		where 
		room_id = $1  
		and $2 < end_date and $3 > start_date
		values ($1, $2, $3);`

	row := m.DB.QueryRowContext(ctx, query, start, end)
	err := row.Scan(&numRows)
	if err != nil {
		return false, err
	}

	if numRows == 0 {
		return true, nil
	}

	return false, nil

}

func (m *postgresDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rooms []models.Room

	query := `select r.id, r.room_name
	from rooms r
	where r.id not in
	(select room_id from room_restrictions rr where $1 < rr.end_date and $2 > rr.start_date)
	values ($1, $2, $3)`

	rows, err := m.DB.QueryContext(ctx, query, start, end)
	if err != nil {
		return rooms, err
	}

	for rows.Next() {
		var room models.Room
		rows.Scan(&room.ID,
			room.RoomName)
		rooms = append(rooms, room)
	}

	if err = rows.Err(); err != nil {
		log.Fatal("Error scanning rows", err)
		return rooms, err
	}

	return rooms, nil

}
