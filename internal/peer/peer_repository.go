package peer

import (
	"bytes"

	"github.com/cryptogarageinc/p2pderivatives-server/internal/database"
	"github.com/jinzhu/gorm"
)

// PeerModel represents a peer in the system.
type PeerModel struct {
	ID   []byte
	Name string
}

// NewPeerModel creates a new PeerModel structure with the given parameters.
func NewPeerModel(id []byte, name string) *PeerModel {
	p := PeerModel{ID: id, Name: name}
	return &p
}

// Equal compares to PeerModel structures and return true if their values are
// equal, false otherwise.
func (peer *PeerModel) Equal(other *PeerModel) bool {
	return bytes.Equal(peer.ID, other.ID) && peer.Name == other.Name
}

// PeerRepository represents a repository abstracting storage operations of
// PeerModel structures.
type PeerRepository interface {
	Create(*PeerModel) error
	Get([]byte) (*PeerModel, error)
	Delete([]byte) error
}

// GormPeerRepository is an implementation of PeerRepository using the Gorm
// ORM library.
type GormPeerRepository struct {
	db *gorm.DB
}

// NewGormPeerRepository creates a new GormPeerRepository structure using
// the provided Gorm DB structure.
func NewGormPeerRepository(db *gorm.DB) (*GormPeerRepository, error) {
	err := db.AutoMigrate(&PeerModel{}).Error

	if err != nil {
		return nil, err
	}
	return &GormPeerRepository{db: db}, nil
}

// Create adds the given peer object to the database.
func (s *GormPeerRepository) Create(peer *PeerModel) error {
	err := s.db.Create(peer).Error

	if err != nil {
		return database.HandleGormError(err, "create", "peer", peer.ID)
	}

	return nil
}

// Get tries to retrieve the peer object with the given id from the database.
func (s *GormPeerRepository) Get(id []byte) (*PeerModel, error) {
	var peer PeerModel
	err := s.db.Where("ID = ?", id).First(&peer).Error

	if err != nil {
		err = database.HandleGormError(err, "get", "peer", id)

		return nil, err
	}

	return &peer, nil
}

// Delete deletes the peer with the given id from the database.
func (s *GormPeerRepository) Delete(id []byte) error {
	err := s.db.Delete(PeerModel{}, "id = ?", id).Error
	if err != nil {
		return database.HandleGormError(err, "delete", "peer", id)
	}

	return err
}
