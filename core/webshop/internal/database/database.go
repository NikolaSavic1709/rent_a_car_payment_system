package database

import (
	"context"
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
	"webshop/internal/model"

	"github.com/google/uuid"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
)

// Service represents a service that interacts with a database.
type Service interface {
	Health() map[string]string
	Close() error

	GetUserByUsernameAndPassword(username, password string) (*model.User, error)
	GetUserByUsername(username string) (*model.User, error)
	CreateUser(fullname, email, username, password, role string) error

	GetAllVehicles() ([]model.Vehicle, error)
	GetVehicleByID(id int) (*model.Vehicle, error)
	GetActiveVehiclesByUser(userID int) ([]model.Payment, error)
	CreateVehicle(vehicle model.Vehicle) error
	CreatePayment(userID int, payment model.Payment, vehicletId int, pspToken string) error
	GetUserByID(userID int) (*model.User, error)

	InsertPurchaseStatus(status model.PurchaseStatus) error
	GetPurchaseStatusByMerchantOrderId(merchantOrderId uuid.UUID) (*model.PurchaseStatus, error)
}

type service struct {
	db *sql.DB
}

var (
	database   = os.Getenv("DB_DATABASE")
	password   = os.Getenv("DB_PASSWORD")
	username   = os.Getenv("DB_USERNAME")
	port       = os.Getenv("DB_HOST_PORT")
	host       = os.Getenv("DB_HOST")
	schema     = os.Getenv("DB_SCHEMA")
	dbInstance *service
)

func New() Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s", username, password, host, port, database, schema)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize data
	err = initializeData(db)
	if err != nil {
		log.Fatal(err)
	}

	dbInstance = &service{
		db: db,
	}
	return dbInstance
}

func checkAndCreateUsersTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(255) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
	    fullname VARCHAR(255) NOT NULL,
	    email VARCHAR(255) NOT NULL,
	    role VARCHAR(20) NOT NULL
	)`
	_, err := db.Exec(query)
	return err
}

func checkAndCreatePurchaseStatusTable(db *sql.DB) error {
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS purchase_status (
		id SERIAL PRIMARY KEY,
		url TEXT NOT NULL,
		merchant_order_id TEXT NOT NULL
	);`

	_, err := db.Exec(createTableQuery)
	return err
}

func checkAndCreateVehiclesTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS vehicles (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		description TEXT NOT NULL,
		price FLOAT NOT NULL,
		category VARCHAR(20) NOT NULL
	)`
	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	paymentTableQuery := `
	CREATE TABLE IF NOT EXISTS payments (
		id SERIAL PRIMARY KEY,
		user_id INT NOT NULL,
		vehicle_id INT NOT NULL REFERENCES vehicles(id),
		deadline DATE NOT NULL,
		cost FLOAT NOT NULL,
	    psp_token VARCHAR(255)
	)`
	_, err = db.Exec(paymentTableQuery)
	return err
}

func initializeData(db *sql.DB) error {
	// Check if users table exists and create it if not
	err := checkAndCreateUsersTable(db)
	if err != nil {
		return err
	}

	err = checkAndCreatePurchaseStatusTable(db)
	if err != nil {
		return err
	}

	err = checkAndCreateVehiclesTable(db)
	if err != nil {
		return err
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		// Users already exist, no need to insert
		return nil
	}

	// Users to insert
	users := []struct {
		Id       int
		Fullname string
		Email    string
		Username string
		Password string
		Role     string
	}{
		{1, "Miki Milan", "mm@gmail.com", "admin", "password", "admin"},
		{2, "Zoki Zoran", "zz@gmail.com", "customer", "password", "customer"},
	}

	for _, u := range users {
		hashedPassword := sha512.Sum512([]byte(u.Password))
		hashedPasswordHex := hex.EncodeToString(hashedPassword[:])

		_, err := db.Exec("INSERT INTO users (id, fullname, email, username, password, role) VALUES ($1, $2, $3, $4, $5, $6)", u.Id, u.Fullname, u.Email, u.Username, hashedPasswordHex, u.Role)

		if err != nil {
			return err
		}
	}

	var vehicleCount int
	err = db.QueryRow("SELECT COUNT(*) FROM vehicles").Scan(&vehicleCount)
	if err != nil {
		return err
	}

	vc_count := fmt.Sprintf("vehicle cound:%d", vehicleCount)
	fmt.Println(vc_count)

	if vehicleCount == 0 {
		vehicles := []struct {
			Id          int
			Category    string
			Name        string
			Description string
			Price       float64
		}{
			{1, "SUV", "Toyota RAV4", "Pouzdan SUV sa hibridnim pogonom, idealan za grad i duža putovanja.", 199.99},
			{2, "Sedan", "EBMW 320d", "Sportski sedan sa dizel motorom, visokim performansama i premium enterijerom.", 299.99},
			{3, "Hatchback", "Volkswagen Golf", "Kompaktan i praktičan automobil, odličan balans između udobnosti i potrošnje.", 119.99},
		}

		for _, v := range vehicles {
			_, err := db.Exec("INSERT INTO vehicles (id, category, name, description, price) VALUES ($1, $2, $3, $4, $5)", v.Id, v.Category, v.Name, v.Description, v.Price)
			if err != nil {
				return err
			}
		}
	}

	// Test payments
	var paymentCount int
	err = db.QueryRow("SELECT COUNT(*) FROM payments").Scan(&paymentCount)
	if err != nil {
		return err
	}

	if paymentCount == 0 {
		payments := []struct {
			UserID    int
			VehicleID int
			Deadline  time.Time
			Cost      float64
			PspToken  string
		}{
			{1, 1, time.Now().AddDate(0, 1, 0), 799.99, "1903408230984209438"},
			{2, 2, time.Now().AddDate(0, 2, 0), 299.99, "4394729384720934792"},
		}

		for _, py := range payments {
			_, err := db.Exec("INSERT INTO payments (user_id, vehicle_id, deadline, cost, psp_token) VALUES ($1, $2, $3, $4, $5)", py.UserID, py.VehicleID, py.Deadline, py.Cost, py.PspToken)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	// Ping the database
	err := s.db.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf("db down: %v", err) // Log the error and terminate the program
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// Get database stats (like open connections, in use, idle, etc.)
	dbStats := s.db.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	// Evaluate stats to provide a health message
	if dbStats.OpenConnections > 40 { // Assuming 50 is the max for this example
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *service) Close() error {
	log.Printf("Disconnected from database: %s", database)
	return s.db.Close()
}

func (s *service) GetUserByUsernameAndPassword(username, password string) (*model.User, error) {
	query := `SELECT id, username, password FROM users WHERE username = $1`
	row := s.db.QueryRow(query, username)
	fmt.Println(row)
	var user model.User
	err := row.Scan(&user.ID, &user.Username, &user.Password)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No user found with the given username
		}
		return nil, err
	}
	// Compute hash of the provided password with the salt
	hashedPassword := sha512.Sum512([]byte(password))
	hashedPasswordHex := hex.EncodeToString(hashedPassword[:])

	// Compare the stored password hash with the computed hash
	if user.Password != hashedPasswordHex {
		return nil, nil // Password does not match
	}

	return &user, nil
}

func (s *service) GetUserByUsername(username string) (*model.User, error) {
	query := `SELECT id, username, password, fullname, email, role FROM users WHERE username = $1`
	row := s.db.QueryRow(query, username)

	var user model.User
	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.Fullname, &user.Email, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (s *service) CreateUser(fullname, email, username, password, role string) error {
	hashedPassword := sha512.Sum512([]byte(password))
	hashedPasswordHex := hex.EncodeToString(hashedPassword[:])

	query := `INSERT INTO users (fullname, email, username, password, role) VALUES ($1, $2, $3, $4, $5)`
	_, err := s.db.Exec(query, fullname, email, username, hashedPasswordHex, role)
	if err != nil {
		return fmt.Errorf("could not insert user: %v", err)
	}

	return nil
}

//+++++++++++++++++++++++++++++++++++++

func (s *service) GetAllVehicles() ([]model.Vehicle, error) {
	query := `SELECT id, category, name, description, price FROM vehicles`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vehicles []model.Vehicle
	for rows.Next() {
		var vehicle model.Vehicle
		err := rows.Scan(&vehicle.ID, &vehicle.Category, &vehicle.Name, &vehicle.Description, &vehicle.Price)
		if err != nil {
			return nil, err
		}

		vehicles = append(vehicles, vehicle)
	}
	return vehicles, nil
}

func (s *service) GetVehicleByID(id int) (*model.Vehicle, error) {
	query := `SELECT id, category, name, description, price FROM vehicles WHERE id = $1`
	row := s.db.QueryRow(query, id)

	var vehicle model.Vehicle
	err := row.Scan(&vehicle.ID, &vehicle.Category, &vehicle.Name, &vehicle.Description, &vehicle.Price)
	if err != nil {
		return nil, err
	}

	return &vehicle, nil
}

func (s *service) GetActiveVehiclesByUser(userID int) ([]model.Payment, error) {
	query := `
		SELECT v.id, v.category, v.name, v.description, v.price, py.deadline 
		FROM payments py 
		JOIN vehicles v ON py.vehicle_id = v.id 
		WHERE py.user_id = ? AND py.deadline > DATE('now')
	`
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []model.Payment
	for rows.Next() {
		var payment model.Payment
		var vehicle model.Vehicle
		err := rows.Scan(&vehicle.ID, &vehicle.Category, &vehicle.Name, &vehicle.Description, &vehicle.Price, &payment.Deadline)
		if err != nil {
			return nil, err
		}
		payment.Vehicle = vehicle
		payments = append(payments, payment)
	}
	return payments, nil
}

func (s *service) CreateVehicle(vehicle model.Vehicle) error {
	fmt.Println("CV")
	query := `INSERT INTO vehicles (category, name, description, price) VALUES ($1, $2, $3, $4) RETURNING id`
	rows, err := s.db.Query(query, vehicle.Category, vehicle.Name, vehicle.Description, vehicle.Price)
	if err != nil {
		return err
	}
	fmt.Println("ECV")
	defer rows.Close()

	var vehicleID int
	if rows.Next() {
		err := rows.Scan(&vehicleID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *service) CreatePayment(userID int, payment model.Payment, vehicleId int, pspToken string) error {
	fmt.Println("CP")
	fmt.Println(payment.Deadline)
	fmt.Println()
	query := `INSERT INTO payments (user_id, vehicle_id, deadline, cost, psp_token) VALUES ($1, $2, $3, $4, $5)`
	_, err := s.db.Exec(query, userID, vehicleId, payment.Deadline, payment.Cost, pspToken)
	return err
}

func (s *service) GetUserByID(userID int) (*model.User, error) {
	userQuery := `SELECT id, fullname, email FROM users WHERE id = $1`
	row := s.db.QueryRow(userQuery, userID)

	var user model.User
	err := row.Scan(&user.ID, &user.Fullname, &user.Email)
	if err != nil {
		return nil, err
	}

	// Dohvati plaćanja korisnika
	paymentsQuery := `
		SELECT v.id, v.category, v.name, v.description, v.price, py.deadline 
		FROM payments py 
		JOIN vehicles v ON py.vehicle_id = v.id 
		WHERE py.user_id = $1
	`
	rows, err := s.db.Query(paymentsQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var payment model.Payment
		var vehicle model.Vehicle
		err := rows.Scan(&vehicle.ID, &vehicle.Category, &vehicle.Name, &vehicle.Description, &vehicle.Price, &payment.Deadline)
		if err != nil {
			return nil, err
		}
		payment.Vehicle = vehicle
		user.Payments = append(user.Payments, payment)
	}
	return &user, nil
}

func (s *service) InsertPurchaseStatus(status model.PurchaseStatus) error {
	insertQuery := `
	INSERT INTO purchase_status (url, merchant_order_id) 
	VALUES ($1, $2);`
	_, err := s.db.Exec(insertQuery, status.URL, status.MerchantOrderId.String())
	if err != nil {
		return err
	}
	return nil
}

func (s *service) GetPurchaseStatusByMerchantOrderId(merchantOrderId uuid.UUID) (*model.PurchaseStatus, error) {
	var purchaseStatus model.PurchaseStatus
	query := `SELECT id, url, merchant_order_id FROM purchase_status WHERE merchant_order_id = $1`
	row := s.db.QueryRow(query, merchantOrderId)
	err := row.Scan(&purchaseStatus.ID, &purchaseStatus.URL, &purchaseStatus.MerchantOrderId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No record found
		}
		return nil, err // Some other error
	}
	return &purchaseStatus, nil
}
