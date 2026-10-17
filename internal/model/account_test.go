package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccount_ToDTO(t *testing.T) {
	t.Run("Complete conversion", func(t *testing.T) {
		account := Account{
			ID:       DBID(42),
			Username: "testuser",
			Password: "secretpassword",
			Owner:    true,
			Config: UserConfig{
				ShowId:        true,
				ListMode:      false,
				HideThumbnail: true,
				HideExcerpt:   false,
				Theme:         "dark",
				KeepMetadata:  true,
				UseArchive:    false,
				CreateEbook:   true,
				MakePublic:    false,
			},
		}

		dto := account.ToDTO()

		assert.Equal(t, DBID(42), dto.ID)
		assert.Equal(t, "testuser", dto.Username)
		// Password should not be copied for security
		assert.Equal(t, "", dto.Password)
		assert.True(t, *dto.Owner)
		assert.Equal(t, "dark", (*dto.Config).Theme)
		assert.True(t, (*dto.Config).ShowId)
	})

	t.Run("Non-owner account conversion", func(t *testing.T) {
		account := Account{
			ID:       DBID(1),
			Username: "regularuser",
			Password: "password123",
			Owner:    false,
			Config: UserConfig{
				Theme: "light",
			},
		}

		dto := account.ToDTO()

		assert.Equal(t, DBID(1), dto.ID)
		assert.Equal(t, "regularuser", dto.Username)
		assert.False(t, *dto.Owner)
	})

	t.Run("Empty account conversion", func(t *testing.T) {
		account := Account{}

		dto := account.ToDTO()

		assert.Equal(t, DBID(0), dto.ID)
		assert.Equal(t, "", dto.Username)
		assert.Equal(t, "", dto.Password)
		assert.False(t, *dto.Owner)
		assert.NotNil(t, dto.Config)
	})
}

func TestAccountDTO_IsOwner(t *testing.T) {
	t.Run("Owner account returns true", func(t *testing.T) {
		owner := true
		dto := AccountDTO{
			Owner: &owner,
		}

		assert.True(t, dto.IsOwner())
	})

	t.Run("Non-owner account returns false", func(t *testing.T) {
		owner := false
		dto := AccountDTO{
			Owner: &owner,
		}

		assert.False(t, dto.IsOwner())
	})

	t.Run("Nil owner returns false", func(t *testing.T) {
		dto := AccountDTO{
			Owner: nil,
		}

		assert.False(t, dto.IsOwner())
	})
}

func TestAccountDTO_IsValidCreate(t *testing.T) {
	t.Run("Valid account", func(t *testing.T) {
		dto := AccountDTO{
			Username: "validuser",
			Password: "validpass",
		}

		err := dto.IsValidCreate()

		assert.NoError(t, err)
	})

	t.Run("Missing username", func(t *testing.T) {
		dto := AccountDTO{
			Password: "validpass",
		}

		err := dto.IsValidCreate()

		assert.Error(t, err)
		assert.IsType(t, ValidationError{}, err)
		validationErr := err.(ValidationError)
		assert.Equal(t, "username", validationErr.Field)
		assert.Equal(t, "username should not be empty", validationErr.Message)
	})

	t.Run("Empty username", func(t *testing.T) {
		dto := AccountDTO{
			Username: "",
			Password: "validpass",
		}

		err := dto.IsValidCreate()

		assert.Error(t, err)
		assert.IsType(t, ValidationError{}, err)
	})

	t.Run("Missing password", func(t *testing.T) {
		dto := AccountDTO{
			Username: "validuser",
		}

		err := dto.IsValidCreate()

		assert.Error(t, err)
		assert.IsType(t, ValidationError{}, err)
		validationErr := err.(ValidationError)
		assert.Equal(t, "password", validationErr.Field)
		assert.Equal(t, "password should not be empty", validationErr.Message)
	})

	t.Run("Both username and password missing", func(t *testing.T) {
		dto := AccountDTO{}

		err := dto.IsValidCreate()

		assert.Error(t, err)
		assert.IsType(t, ValidationError{}, err)
		// Should report first validation error (username)
		validationErr := err.(ValidationError)
		assert.Equal(t, "username", validationErr.Field)
	})
}

func TestAccountDTO_IsValidUpdate(t *testing.T) {
	t.Run("Valid update with username", func(t *testing.T) {
		dto := AccountDTO{
			Username: "newusername",
		}

		err := dto.IsValidUpdate()

		assert.NoError(t, err)
	})

	t.Run("Valid update with password", func(t *testing.T) {
		dto := AccountDTO{
			Password: "newpassword",
		}

		err := dto.IsValidUpdate()

		assert.NoError(t, err)
	})

	t.Run("Valid update with owner status", func(t *testing.T) {
		owner := true
		dto := AccountDTO{
			Owner: &owner,
		}

		err := dto.IsValidUpdate()

		assert.NoError(t, err)
	})

	t.Run("Valid update with config", func(t *testing.T) {
		config := UserConfig{
			Theme: "dark",
		}
		dto := AccountDTO{
			Config: &config,
		}

		err := dto.IsValidUpdate()

		assert.NoError(t, err)
	})

	t.Run("Valid update with multiple fields", func(t *testing.T) {
		owner := false
		config := UserConfig{
			Theme: "light",
		}
		dto := AccountDTO{
			Username: "updateduser",
			Password: "updatedpass",
			Owner:    &owner,
			Config:   &config,
		}

		err := dto.IsValidUpdate()

		assert.NoError(t, err)
	})

	t.Run("Invalid update with no fields", func(t *testing.T) {
		dto := AccountDTO{}

		err := dto.IsValidUpdate()

		assert.Error(t, err)
		assert.IsType(t, ValidationError{}, err)
		validationErr := err.(ValidationError)
		assert.Equal(t, "account", validationErr.Field)
		assert.Equal(t, "no fields to update", validationErr.Message)
	})

	t.Run("Invalid update with empty strings", func(t *testing.T) {
		dto := AccountDTO{
			Username: "",
			Password: "",
		}

		err := dto.IsValidUpdate()

		assert.Error(t, err)
		assert.IsType(t, ValidationError{}, err)
	})
}

func TestUserConfig_Serialization(t *testing.T) {
	t.Run("Marshal to JSON", func(t *testing.T) {
		config := UserConfig{
			ShowId:        true,
			ListMode:      false,
			HideThumbnail: true,
			HideExcerpt:   false,
			Theme:         "dark",
			KeepMetadata:  true,
			UseArchive:    false,
			CreateEbook:   true,
			MakePublic:    false,
		}

		data, err := config.Value()

		require.NoError(t, err)

		var unmarshaled map[string]interface{}
		err = json.Unmarshal(data.([]byte), &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, "dark", unmarshaled["Theme"])
		assert.Equal(t, true, unmarshaled["ShowId"])
	})

	t.Run("Scan from JSON bytes", func(t *testing.T) {
		jsonData := `{"Theme":"light","ShowId":false}`

		var config UserConfig
		err := config.Scan([]byte(jsonData))

		require.NoError(t, err)
		assert.Equal(t, "light", config.Theme)
		assert.Equal(t, false, config.ShowId)
	})

	t.Run("Scan from JSON string", func(t *testing.T) {
		jsonData := `{"Theme":"dark","KeepMetadata":true}`

		var config UserConfig
		err := config.Scan(jsonData)

		require.NoError(t, err)
		assert.Equal(t, "dark", config.Theme)
		assert.Equal(t, true, config.KeepMetadata)
	})

	t.Run("Scan from unsupported type", func(t *testing.T) {
		var config UserConfig
		err := config.Scan(123)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported type")
	})

	t.Run("Default config values", func(t *testing.T) {
		var config UserConfig

		assert.Equal(t, "", config.Theme)
		assert.False(t, config.ShowId)
		assert.False(t, config.ListMode)
		assert.False(t, config.HideThumbnail)
		assert.False(t, config.HideExcerpt)
		assert.False(t, config.KeepMetadata)
		assert.False(t, config.UseArchive)
		assert.False(t, config.CreateEbook)
		assert.False(t, config.MakePublic)
	})

	t.Run("NewUserConfig returns config with defaults", func(t *testing.T) {
		config := NewUserConfig()

		assert.Equal(t, "system", config.Theme)
		assert.False(t, config.ShowId)
		assert.False(t, config.ListMode)
		assert.False(t, config.HideThumbnail)
		assert.False(t, config.HideExcerpt)
		assert.False(t, config.KeepMetadata)
		assert.False(t, config.UseArchive)
		assert.False(t, config.CreateEbook)
		assert.False(t, config.MakePublic)
	})

	t.Run("Defaults method sets theme to system when empty", func(t *testing.T) {
		config := UserConfig{
			ShowId:        true,
			ListMode:      true,
			HideThumbnail: true,
			// Theme is empty
		}

		config.Defaults()

		assert.Equal(t, "system", config.Theme)
		assert.True(t, config.ShowId)
		assert.True(t, config.ListMode)
		assert.True(t, config.HideThumbnail)
	})

	t.Run("Defaults method preserves existing theme", func(t *testing.T) {
		config := UserConfig{
			Theme: "dark",
		}

		config.Defaults()

		assert.Equal(t, "dark", config.Theme)
	})

	t.Run("Scan applies defaults automatically", func(t *testing.T) {
		jsonData := `{"ShowId":true,"ListMode":false}`
		// Note: Theme is not in the JSON, so it should be empty and get defaulted

		var config UserConfig
		err := config.Scan([]byte(jsonData))

		assert.NoError(t, err)
		assert.Equal(t, "system", config.Theme) // Should be defaulted
		assert.True(t, config.ShowId)
		assert.False(t, config.ListMode)
	})
}
