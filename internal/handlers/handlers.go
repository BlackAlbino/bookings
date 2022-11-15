package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/PushAndRun/bookings/internal/config"
	"github.com/PushAndRun/bookings/internal/driver"
	"github.com/PushAndRun/bookings/internal/forms"
	"github.com/PushAndRun/bookings/internal/helpers"
	"github.com/PushAndRun/bookings/internal/models"
	"github.com/PushAndRun/bookings/internal/render"
	"github.com/PushAndRun/bookings/internal/repository"
	"github.com/PushAndRun/bookings/internal/repository/dbrepo"
)

// Repo used by the handlers
var Repo *Repository

//Repository is the repository type
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

//NewRepo creates a new repository
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(db.SQL, a),
	}
}

//NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// About is the handler for the about page
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	//stringMap := make(map[string]string)
	//stringMap["test"] = "Hello, again."

	//remoteIP := m.App.Session.GetString(r.Context(), "remote_ip")
	//stringMap["remote_ip"] = remoteIP

	render.Template(w, r, "about.page.templ", &models.TemplateData{
		//StringMap: stringMap,
	})

}

func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	remoteIP := r.RemoteAddr
	m.App.Session.Put(r.Context(), "remote_ip", remoteIP)
	render.Template(w, r, "home.page.templ", &models.TemplateData{})
}

func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "majors.page.templ", &models.TemplateData{})
}

func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "generals.page.templ", &models.TemplateData{})
}

func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	var emptyReservation models.Reservation
	data := make(map[string]interface{})
	data["reservation"] = emptyReservation

	render.Template(w, r, "make-reservation.page.templ", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
	})
}

func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	sd := r.Form.Get("start_date")
	ed := r.Form.Get("end_date")

	// 01/02 04:04:05PM '06 -007
	layout := "01-02-2006"

	startDate, err := time.Parse(layout, sd)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	endDate, err := time.Parse(layout, ed)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	roomID, err := strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Email:     r.Form.Get("e-mail"),
		Phone:     r.Form.Get("phone"),
		StartDate: startDate,
		EndDate:   endDate,
		RoomID:    roomID,
	}

	form := forms.New(r.PostForm)

	form.Required("first_name", "last_name", "e-mail")
	form.HasMinLength("last_name", 2)
	form.IsEmail("e-mail")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation

		render.Template(w, r, "make-reservation.page.templ", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return

	}

	fmt.Println("Start to insert reservation")
	newReservationID, err := m.DB.InsertReservation(reservation)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	fmt.Println(fmt.Sprintf("Finished reservation insertion, new ID is %d", newReservationID))

	restriction := models.RoomRestriction{
		StartDate:     startDate,
		EndDate:       endDate,
		RoomID:        roomID,
		ReservationID: newReservationID,
		RestrictionID: 1,
	}

	fmt.Println("Start to insert reservation restriction")

	err = m.DB.InsertRoomRestriction(restriction)
	if err != nil {
		helpers.ServerError(w, err)
		fmt.Sprintln("Failed to insert room restriction")
		return
	}

	m.App.Session.Put(r.Context(), "reservation", reservation)

	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)

	if !ok {
		m.App.ErrorLog.Println("Cannot get item from session!")
		m.App.Session.Put(r.Context(), "error", "Can't find reservation details. Please make a new reservation.")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	m.App.Session.Remove(r.Context(), "reservation")

	data := make(map[string]interface{})
	data["reservation"] = reservation

	render.Template(w, r, "reservation-summary.page.templ", &models.TemplateData{
		Data: data,
	})
}

func (m *Repository) SearchAvailability(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "search-availability.page.templ", &models.TemplateData{})
}

func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	start := r.Form.Get("arrival")
	end := r.Form.Get("departure")
	w.Write([]byte(fmt.Sprintf("Start date is %s and end date is %s", start, end)))
}

type jsonResponse struct {
	OK      bool   `json: "ok"`
	Message string `json: "message"`
}

func (m *Repository) PostAvailabilityJson(w http.ResponseWriter, r *http.Request) {

	resp := jsonResponse{
		OK:      true,
		Message: "available",
	}

	out, err := json.MarshalIndent(resp, "", "     ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	log.Print(string(out))
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}
