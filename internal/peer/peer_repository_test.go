package peer

import (
	"math/rand"
	"testing"

	"github.com/cryptogarageinc/p2pderivatives-server/internal/database"
	"github.com/stretchr/testify/assert"
)

func getRandomBytes(seed int64) []byte {
	bytes := make([]byte, 4)
	rand.Seed(seed)
	rand.Read(bytes)
	return bytes

}

func createID() []byte {
	return getRandomBytes(7)
}

func createPeer() *PeerModel {
	return NewPeerModel(createID(), "TestPeer")
}

func createPeerRepository() PeerRepository {
	db, _ := database.NewGormSqlite(":memory:")
	gormRepository, _ := NewGormPeerRepository(db)
	return gormRepository
}

func closePeerRepository(repository PeerRepository) {
	gormRepository := repository.(*GormPeerRepository)
	gormRepository.db.Close()
}

func TestCreatePeerWithNewIdSucceeds(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	peerRepository := createPeerRepository()
	defer closePeerRepository(peerRepository)
	peer := createPeer()

	// Act
	peerRepository.Create(peer)
	peer2, err := peerRepository.Get(peer.ID)

	// Assert
	assert.NoError(err)
	assert.True(peer.Equal(peer2))
}

func TestGetPeerWithNonExistingIdReturnsNotFoundError(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	peerRepository := createPeerRepository()
	defer closePeerRepository(peerRepository)
	id := createID()

	// Act
	peer, err := peerRepository.Get(id)
	dbError, ok := err.(*database.DbError)

	// Assert
	assert.Nil(peer)
	assert.Error(err)
	assert.True(ok)
	assert.NotNil(dbError)
	assert.Equal(database.NotFound, dbError.Code())
}

func TestDeletePeerWithExistingPeerSucceeds(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	peerRepository := createPeerRepository()
	defer closePeerRepository(peerRepository)
	peer := createPeer()
	peerRepository.Create(peer)

	// Act
	err := peerRepository.Delete(peer.ID)
	peer2, err2 := peerRepository.Get(peer.ID)
	dbError, ok := err2.(*database.DbError)

	// Assert
	assert.NoError(err)
	assert.Nil(peer2)
	assert.Error(err2)
	assert.True(ok)
	assert.Equal(database.NotFound, dbError.Code())
}

func TestDeletePeerWithExistingPeerIsDeleted(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	peerRepository := createPeerRepository()
	defer closePeerRepository(peerRepository)
	peer := createPeer()
	peerRepository.Create(peer)
	peerRepository.Delete(peer.ID)

	// Act
	peer2, err := peerRepository.Get(peer.ID)
	dbError, ok := err.(*database.DbError)

	// Assert
	assert.Nil(peer2)
	assert.Error(err)
	assert.True(ok)
	assert.Equal(database.NotFound, dbError.Code())
}

func TestDeletePeerWithExistingPeerOtherPeersNotDeleted(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	peerRepository := createPeerRepository()
	defer closePeerRepository(peerRepository)
	peer := createPeer()
	peer2 := NewPeerModel(getRandomBytes(11), "OtherPeer")
	peerRepository.Create(peer)
	peerRepository.Create(peer2)
	peerRepository.Delete(peer.ID)

	// Act
	peer2Copy, err := peerRepository.Get(peer2.ID)

	// Assert
	assert.NoError(err)
	assert.True(peer2.Equal(peer2Copy))
}
