package model

import "time"

// ConnectionMode represents Redis connection mode
type ConnectionMode string

const (
	ModeStandalone ConnectionMode = "standalone"
	ModeCluster    ConnectionMode = "cluster"
	ModeSentinel   ConnectionMode = "sentinel"
)

// SSHConfig represents SSH tunnel configuration
type SSHConfig struct {
	Enabled              bool   `json:"enabled"`
	Host                 string `json:"host"`
	Port                 int    `json:"port"`
	Username             string `json:"username"`
	AuthType             string `json:"authType"` // password or privateKey
	Password             string `json:"password"`
	PrivateKey           string `json:"privateKey"`
	PrivateKeyPassphrase string `json:"privateKeyPassphrase"`
}

// TLSConfig represents TLS configuration
type TLSConfig struct {
	Enabled    bool   `json:"enabled"`
	CACert     string `json:"caCert"`
	ClientCert string `json:"clientCert"`
	ClientKey  string `json:"clientKey"`
	SkipVerify bool   `json:"skipVerify"`
}

// ConnectionGroup represents a group of connections
type ConnectionGroup struct {
	ID        string    `json:"id" gorm:"primaryKey;size:36"`
	UserID    uint      `json:"userId" gorm:"index;not null;default:0"` // User isolation
	Name      string    `json:"name" gorm:"size:100;not null"`
	Icon      string    `json:"icon" gorm:"size:50"`
	Color     string    `json:"color" gorm:"size:20"`
	SortOrder int       `json:"sortOrder" gorm:"default:0"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ConnectionConfig represents Redis connection configuration
type ConnectionConfig struct {
	ID                 string         `json:"id" gorm:"primaryKey;size:36"`
	GroupID            string         `json:"groupId" gorm:"size:36;index"`
	Name               string         `json:"name" gorm:"size:100;not null"`
	Mode               ConnectionMode `json:"mode" gorm:"size:20;default:'standalone'"`
	Host               string         `json:"host" gorm:"size:255"`
	Port               int            `json:"port" gorm:"default:6379"`
	Username           string         `json:"username" gorm:"size:100"`
	Password           string         `json:"password" gorm:"size:255"`
	SentinelMasterName string         `json:"sentinelMasterName" gorm:"size:100"`
	SentinelNodesJSON  string         `json:"-" gorm:"column:sentinel_nodes;type:text"`
	ClusterNodesJSON   string         `json:"-" gorm:"column:cluster_nodes;type:text"`
	SSHJSON            string         `json:"-" gorm:"column:ssh_config;type:text"`
	TLSJSON            string         `json:"-" gorm:"column:tls_config;type:text"`
	SortOrder          int            `json:"sortOrder" gorm:"default:0"`
	CreatedAt          time.Time      `json:"createdAt"`
	UpdatedAt          time.Time      `json:"updatedAt"`

	// Non-database fields (for JSON serialization)
	SentinelNodes []string  `json:"sentinelNodes" gorm:"-"`
	ClusterNodes  []string  `json:"clusterNodes" gorm:"-"`
	SSH           SSHConfig `json:"ssh" gorm:"-"`
	TLS           TLSConfig `json:"tls" gorm:"-"`
}

// TableName specifies table name for ConnectionConfig
func (ConnectionConfig) TableName() string {
	return "connections"
}

// TableName specifies table name for ConnectionGroup
func (ConnectionGroup) TableName() string {
	return "connection_groups"
}

// DatabaseAlias represents database alias stored in MySQL
type DatabaseAlias struct {
	ID           uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	ConnectionID string    `json:"connectionId" gorm:"size:36;index"`
	DBIndex      int       `json:"dbIndex"`
	Alias        string    `json:"alias" gorm:"size:100"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// TableName specifies table name for DatabaseAlias
func (DatabaseAlias) TableName() string {
	return "database_aliases"
}

// DatabaseInfo represents database information
type DatabaseInfo struct {
	Index    int    `json:"index"`
	Alias    string `json:"alias"`
	KeyCount int64  `json:"keyCount"`
}

// KeyInfo represents basic key information
type KeyInfo struct {
	Key  string `json:"key"`
	Type string `json:"type"`
	TTL  int64  `json:"ttl"`
}

// KeyValue represents key with its value
type KeyValue struct {
	Key    string      `json:"key"`
	Type   string      `json:"type"`
	Value  interface{} `json:"value"`
	TTL    int64       `json:"ttl"`
	Size   int64       `json:"size,omitempty"`
	Format string      `json:"format,omitempty"`
}

// KeysRequest represents a request to list keys
type KeysRequest struct {
	DB      int    `json:"db"`
	Pattern string `json:"pattern"`
	Cursor  uint64 `json:"cursor"`
	Count   int64  `json:"count"`
}

// ScanRequest represents a scan request with regex support
type ScanRequest struct {
	DB           int    `json:"db"`
	SearchType   string `json:"searchType"` // keys, scan, regex
	Pattern      string `json:"pattern"`
	RegexPattern string `json:"regexPattern"`
	ScanCount    int64  `json:"scanCount"`
}

// CreateKeyRequest represents a request to create a key
type CreateKeyRequest struct {
	DB    int         `json:"db"`
	Key   string      `json:"key"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
	TTL   int64       `json:"ttl"`
}

// UpdateKeyRequest represents a request to update a key
type UpdateKeyRequest struct {
	DB    int         `json:"db"`
	Value interface{} `json:"value"`
}

// DeleteKeysRequest represents a request to delete keys
type DeleteKeysRequest struct {
	DB   int      `json:"db"`
	Keys []string `json:"keys"`
}

// TTLRequest represents a request to set TTL
type TTLRequest struct {
	DB  int   `json:"db"`
	TTL int64 `json:"ttl"`
}

// RenameKeyRequest represents a request to rename a key
type RenameKeyRequest struct {
	DB     int    `json:"db"`
	NewKey string `json:"newKey"`
}

// ListPushRequest represents a request to push to list
type ListPushRequest struct {
	DB        int      `json:"db"`
	Values    []string `json:"values"`
	Direction string   `json:"direction"` // left or right
}

// ListSetRequest represents a request to set list element
type ListSetRequest struct {
	DB    int    `json:"db"`
	Value string `json:"value"`
}

// HashSetRequest represents a request to set hash field
type HashSetRequest struct {
	DB    int    `json:"db"`
	Value string `json:"value"`
}

// HashDeleteRequest represents a request to delete hash fields
type HashDeleteRequest struct {
	DB     int      `json:"db"`
	Fields []string `json:"fields"`
}

// SetMembersRequest represents a request to add/remove set members
type SetMembersRequest struct {
	DB      int      `json:"db"`
	Members []string `json:"members"`
}

// ZSetMember represents a zset member with score
type ZSetMember struct {
	Member string  `json:"member"`
	Score  float64 `json:"score"`
}

// ZSetAddRequest represents a request to add zset members
type ZSetAddRequest struct {
	DB      int          `json:"db"`
	Members []ZSetMember `json:"members"`
}

// ZSetRemoveRequest represents a request to remove zset members
type ZSetRemoveRequest struct {
	DB      int      `json:"db"`
	Members []string `json:"members"`
}

// GeoLocation represents a geographic location
type GeoLocation struct {
	Name      string  `json:"name"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

// GeoAddRequest represents a request to add geo locations
type GeoAddRequest struct {
	DB        int           `json:"db"`
	Locations []GeoLocation `json:"locations"`
}

// GeoRemoveRequest represents a request to remove geo members
type GeoRemoveRequest struct {
	DB      int      `json:"db"`
	Members []string `json:"members"`
}

// StreamAddRequest represents a request to add stream message
type StreamAddRequest struct {
	DB     int                    `json:"db"`
	Values map[string]interface{} `json:"values"`
}

// StreamDeleteRequest represents a request to delete stream messages
type StreamDeleteRequest struct {
	DB  int      `json:"db"`
	IDs []string `json:"ids"`
}

// HyperLogLogAddRequest represents a request to add elements to hyperloglog
type HyperLogLogAddRequest struct {
	DB       int      `json:"db"`
	Elements []string `json:"elements"`
}

// BitmapSetRequest represents a request to set bitmap bit
type BitmapSetRequest struct {
	DB     int   `json:"db"`
	Offset int64 `json:"offset"`
	Value  int   `json:"value"`
}

// DatabaseAliasRequest represents a request to update database alias
type DatabaseAliasRequest struct {
	Alias string `json:"alias"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// KeysResponse represents a response with keys list
type KeysResponse struct {
	Success bool      `json:"success"`
	Keys    []KeyInfo `json:"keys"`
	Cursor  uint64    `json:"cursor"`
	Error   string    `json:"error,omitempty"`
}

// DeleteResponse represents a delete operation response
type DeleteResponse struct {
	Success bool   `json:"success"`
	Deleted int64  `json:"deleted"`
	Error   string `json:"error,omitempty"`
}

// TTLResponse represents a TTL query response
type TTLResponse struct {
	Success bool   `json:"success"`
	TTL     int64  `json:"ttl"`
	Error   string `json:"error,omitempty"`
}

// CommandRequest represents a Redis command execution request
type CommandRequest struct {
	DB      int    `json:"db"`
	Command string `json:"command"`
}

// CommandResponse represents a Redis command execution response
type CommandResponse struct {
	Success bool   `json:"success"`
	Result  string `json:"result,omitempty"`
	Error   string `json:"error,omitempty"`
}
