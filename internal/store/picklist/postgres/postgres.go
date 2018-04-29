package postgres

import (
	"database/sql"

	"github.com/Pigmice2733/scouting-backend/internal/store"

	"github.com/Pigmice2733/scouting-backend/internal/store/picklist"
)

// Service is used for getting information about a picklist from a postgres database.
type Service struct {
	db *sql.DB
}

// New creates a new picklist service.
func New(db *sql.DB) picklist.Service {
	return &Service{db: db}
}

// Get retrieves a picklist from the postgresql database given an id.
func (s *Service) Get(id string) (p picklist.Picklist, err error) {
	err = s.db.QueryRow("SELECT id, eventKey, name, owner FROM picklists WHERE id = $1", id).Scan(
		&p.ID, &p.EventKey, &p.Name, &p.Owner)
	if err != nil {
		if err == sql.ErrNoRows {
			err = store.ErrNoResults
		}
		return p, err
	}

	rows, err := s.db.Query("SELECT team FROM picks WHERE picklistId = $1", id)
	if err != nil {
		return p, err
	}
	defer rows.Close()

	for rows.Next() {
		var team string
		if err := rows.Scan(&team); err != nil {
			return p, err
		}

		p.List = append(p.List, team)
	}

	return p, rows.Err()
}

// GetBasicPicklists gets all basic picklists that belong to a certain user from the postgresql database.
func (s *Service) GetBasicPicklists(username string) (bPicklists []picklist.BasicPicklist, err error) {
	rows, err := s.db.Query("SELECT id, eventKey, name FROM picklists WHERE owner = $1", username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var bPicklist picklist.BasicPicklist
		if err := rows.Scan(&bPicklist.ID, &bPicklist.EventKey, &bPicklist.Name); err != nil {
			return nil, err
		}

		bPicklists = append(bPicklists, bPicklist)
	}

	return bPicklists, rows.Err()
}

// Insert inserts a picklist into the postgresql database.
func (s *Service) Insert(p picklist.Picklist) (string, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return p.ID, err
	}

	err = tx.QueryRow(`
		INSERT 
			INTO
				picklists(eventKey, name, owner)
			VALUES ($1, $2, $3)
			RETURNING id
		`, p.EventKey, p.Name, p.Owner).Scan(&p.ID)
	if err != nil {
		tx.Rollback()
		return p.ID, err
	}

	stmt, err := tx.Prepare("INSERT INTO picks(picklistId, team) VALUES ($1, $2)")
	if err != nil {
		tx.Rollback()
		return p.ID, err
	}
	defer stmt.Close()

	for _, pick := range p.List {
		if _, err := stmt.Exec(p.ID, pick); err != nil {
			return p.ID, err
		}
	}

	return p.ID, tx.Commit()
}

// GetOwner retrieves the owner of a picklist in the postgresql database.
func (s *Service) GetOwner(id string) (username string, err error) {
	err = s.db.QueryRow("SELECT owner FROM picklists WHERE id = $1", id).Scan(&username)
	if err == sql.ErrNoRows {
		err = store.ErrNoResults
	}
	return
}

// Update updates a picklist in the postgresql database.
func (s *Service) Update(p picklist.Picklist) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		UPDATE picklists
			SET eventKey = $1, name = $2
			WHERE OWNER = $3 AND id = $4
		`, p.EventKey, p.Name, p.Owner, p.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	if _, err = tx.Exec("DELETE FROM picks WHERE picklistId = $1", p.ID); err != nil {
		tx.Rollback()
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO picks(picklistId, team) VALUES ($1, $2)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, pick := range p.List {
		if _, err := stmt.Exec(p.ID, pick); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Delete deletes a picklist from the postgresql database.
func (s *Service) Delete(id string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec("DELETE FROM picks WHERE picklistId = $1", id); err != nil {
		tx.Rollback()
		return err
	}

	if _, err := tx.Exec("DELETE FROM picklists WHERE id = $1", id); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// GetByEvent gets basic picklists from the postgresql database by username and eventKey.
func (s *Service) GetByEvent(username, eventKey string) (bPicklists []picklist.BasicPicklist, err error) {
	rows, err := s.db.Query("SELECT id, eventKey, name FROM picklists WHERE owner = $1 AND eventKey = $2", username, eventKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var bPicklist picklist.BasicPicklist
		if err := rows.Scan(&bPicklist.ID, &bPicklist.EventKey, &bPicklist.Name); err != nil {
			return nil, err
		}

		bPicklists = append(bPicklists, bPicklist)
	}

	return bPicklists, rows.Err()
}
