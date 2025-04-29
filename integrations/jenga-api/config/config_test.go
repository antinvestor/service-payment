package config

import (
	"os"
	"testing"

	"github.com/pitabwire/frame"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name           string
		configPath     string
		createTestFile bool
		testFileData   string
		expectError    bool
	}{
		{
			name:           "Happy path - config file exists and is valid",
			configPath:     "test_config.yaml",
			createTestFile: true,
			testFileData: `
jengaPrivateKey: "/app/keys/privatekey.pem"
apiKey: "TestApiKey"
consumerSecret: "TestConsumerSecret"
merchantCode: "1234567890"
env: "https://test.finserve.africa"
profileServiceURI: "test:7005"
partitionServiceURI: "test:7003"
securelyRunService: false
redisHost: "testhost"
redisPort: "6379"
`,
			expectError: false,
		},
		{
			name:           "Error path - config file does not exist",
			configPath:     "non_existent_config.yaml",
			createTestFile: false,
			testFileData:   "",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test file if needed
			if tt.createTestFile {
				err := os.WriteFile(tt.configPath, []byte(tt.testFileData), 0644)
				assert.NoError(t, err, "Failed to create test config file")
				defer os.Remove(tt.configPath)
			}

			// Create a JengaConfig struct and load config using frame package
			config := &JengaConfig{}
			
			// Set config file path to the test file
			if tt.configPath != "" {
				os.Setenv("CONFIG_FILE", tt.configPath)
				defer os.Unsetenv("CONFIG_FILE")
			}
			
			err := frame.ConfigProcess("", config)

			// Check expectations
			if tt.expectError {
				assert.Error(t, err, "Expected an error but got none")
			} else {
				assert.NoError(t, err, "Expected no error but got one")
				
				// Verify that values were loaded correctly
				assert.Equal(t, "/app/keys/privatekey.pem", config.JengaPrivateKey)
				assert.Equal(t, "TestApiKey", config.ApiKey)
				assert.Equal(t, "TestConsumerSecret", config.ConsumerSecret)
				assert.Equal(t, "1234567890", config.MerchantCode)
				assert.Equal(t, "https://test.finserve.africa", config.Env)
				assert.Equal(t, "test:7005", config.ProfileServiceURI)
				assert.Equal(t, "test:7003", config.PartitionServiceURI)
				assert.Equal(t, false, config.SecurelyRunService)
				assert.Equal(t, "testhost", config.RedisHost)
				assert.Equal(t, "6379", config.RedisPort)
			}
		})
	}
}
