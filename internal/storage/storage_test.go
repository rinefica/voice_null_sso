package storage

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rinefica/voice_null_sso/internal/lib/sl"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

var testDbInstance *pgxpool.Pool

func TestMain(m *testing.M) {
	testDB := SetupTestDatabase()
	testDbInstance = testDB.DbInstance
	defer testDB.TearDown()
	os.Exit(m.Run())
}

func TestCreateUser(t *testing.T) {
	ds := Storage{
		sl.SetupLogger(""),
		testDbInstance,
	}

	email := "test@mail.co"
	userID, err := ds.SaveUser(context.Background(), email, []byte("jkjkjk"))

	log.Println(userID)

	assert.NotNil(t, userID)
	assert.NoError(t, err)

	id := int64(1)
	assert.Equal(t, id, userID)

	user, err := ds.User(context.Background(), email)
	assert.NoError(t, err)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, id, user.ID)

	user, err = ds.User(context.Background(), email+"123")
	assert.Error(t, err)
	assert.Nil(t, user)
}
