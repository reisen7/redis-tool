package service

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/ssh"
	"redis-web-manager/model"
)

// ConnectionManager manages Redis connections
type ConnectionManager struct {
	connections map[string]*RedisConnection
	mu          sync.RWMutex
}

// RedisConnection represents an active Redis connection
type RedisConnection struct {
	Config     model.ConnectionConfig
	Client     redis.UniversalClient
	SSHTunnel  *SSHTunnel
	CurrentDB  int
	DBInfo     map[int]*model.DatabaseInfo
	mu         sync.RWMutex
}

// SSHTunnel represents an SSH tunnel
type SSHTunnel struct {
	Client   *ssh.Client
	Listener net.Listener
	LocalAddr string
}

var manager = &ConnectionManager{
	connections: make(map[string]*RedisConnection),
}

// GetManager returns the connection manager instance
func GetManager() *ConnectionManager {
	return manager
}

// Connect creates a new Redis connection
func (m *ConnectionManager) Connect(config model.ConnectionConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Close existing connection if any
	if existing, ok := m.connections[config.ID]; ok {
		existing.Close()
	}

	conn := &RedisConnection{
		Config:    config,
		CurrentDB: 0,
		DBInfo:    make(map[int]*model.DatabaseInfo),
	}

	// Setup SSH tunnel if enabled
	var redisAddr string
	if config.SSH.Enabled {
		tunnel, err := createSSHTunnel(config)
		if err != nil {
			return fmt.Errorf("SSH tunnel error: %v", err)
		}
		conn.SSHTunnel = tunnel
		redisAddr = tunnel.LocalAddr
	} else {
		redisAddr = fmt.Sprintf("%s:%d", config.Host, config.Port)
	}

	// Create Redis client based on mode
	var client redis.UniversalClient
	var err error

	switch config.Mode {
	case model.ModeCluster:
		client, err = createClusterClient(config, redisAddr)
	case model.ModeSentinel:
		client, err = createSentinelClient(config)
	default:
		client, err = createStandaloneClient(config, redisAddr)
	}

	if err != nil {
		if conn.SSHTunnel != nil {
			conn.SSHTunnel.Close()
		}
		return err
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		if conn.SSHTunnel != nil {
			conn.SSHTunnel.Close()
		}
		return fmt.Errorf("ping failed: %v", err)
	}

	conn.Client = client
	m.connections[config.ID] = conn

	return nil
}

// createSSHTunnel creates an SSH tunnel
func createSSHTunnel(config model.ConnectionConfig) (*SSHTunnel, error) {
	sshConfig := &ssh.ClientConfig{
		User:            config.SSH.Username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	if config.SSH.AuthType == "privateKey" {
		var signer ssh.Signer
		var err error
		if config.SSH.PrivateKeyPassphrase != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(config.SSH.PrivateKey), []byte(config.SSH.PrivateKeyPassphrase))
		} else {
			signer, err = ssh.ParsePrivateKey([]byte(config.SSH.PrivateKey))
		}
		if err != nil {
			return nil, fmt.Errorf("parse private key: %v", err)
		}
		sshConfig.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	} else {
		sshConfig.Auth = []ssh.AuthMethod{ssh.Password(config.SSH.Password)}
	}

	sshAddr := fmt.Sprintf("%s:%d", config.SSH.Host, config.SSH.Port)
	sshClient, err := ssh.Dial("tcp", sshAddr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("SSH dial: %v", err)
	}

	// Create local listener
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		sshClient.Close()
		return nil, fmt.Errorf("local listener: %v", err)
	}

	tunnel := &SSHTunnel{
		Client:    sshClient,
		Listener:  listener,
		LocalAddr: listener.Addr().String(),
	}

	// Start forwarding
	remoteAddr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	go func() {
		for {
			localConn, err := listener.Accept()
			if err != nil {
				return
			}

			remoteConn, err := sshClient.Dial("tcp", remoteAddr)
			if err != nil {
				localConn.Close()
				continue
			}

			go func() {
				defer localConn.Close()
				defer remoteConn.Close()
				go io.Copy(localConn, remoteConn)
				io.Copy(remoteConn, localConn)
			}()
		}
	}()

	return tunnel, nil
}

// Close closes the SSH tunnel
func (t *SSHTunnel) Close() {
	if t.Listener != nil {
		t.Listener.Close()
	}
	if t.Client != nil {
		t.Client.Close()
	}
}

// createTLSConfig creates TLS configuration
func createTLSConfig(config model.ConnectionConfig) (*tls.Config, error) {
	if !config.TLS.Enabled {
		return nil, nil
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: config.TLS.SkipVerify,
	}

	if config.TLS.CACert != "" {
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM([]byte(config.TLS.CACert)) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}

	if config.TLS.ClientCert != "" && config.TLS.ClientKey != "" {
		cert, err := tls.X509KeyPair([]byte(config.TLS.ClientCert), []byte(config.TLS.ClientKey))
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %v", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}

// createStandaloneClient creates a standalone Redis client
func createStandaloneClient(config model.ConnectionConfig, addr string) (*redis.Client, error) {
	tlsConfig, err := createTLSConfig(config)
	if err != nil {
		return nil, err
	}

	opts := &redis.Options{
		Addr:         addr,
		Username:     config.Username,
		Password:     config.Password,
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		TLSConfig:    tlsConfig,
	}

	return redis.NewClient(opts), nil
}

// createClusterClient creates a cluster Redis client
func createClusterClient(config model.ConnectionConfig, _ string) (*redis.ClusterClient, error) {
	tlsConfig, err := createTLSConfig(config)
	if err != nil {
		return nil, err
	}

	addrs := config.ClusterNodes
	if len(addrs) == 0 {
		addrs = []string{fmt.Sprintf("%s:%d", config.Host, config.Port)}
	}

	opts := &redis.ClusterOptions{
		Addrs:        addrs,
		Username:     config.Username,
		Password:     config.Password,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		TLSConfig:    tlsConfig,
	}

	return redis.NewClusterClient(opts), nil
}

// createSentinelClient creates a sentinel Redis client
func createSentinelClient(config model.ConnectionConfig) (*redis.Client, error) {
	tlsConfig, err := createTLSConfig(config)
	if err != nil {
		return nil, err
	}

	addrs := config.SentinelNodes
	if len(addrs) == 0 {
		addrs = []string{fmt.Sprintf("%s:%d", config.Host, config.Port)}
	}

	opts := &redis.FailoverOptions{
		MasterName:       config.SentinelMasterName,
		SentinelAddrs:    addrs,
		Username:         config.Username,
		Password:         config.Password,
		SentinelUsername: config.Username,
		SentinelPassword: config.Password,
		DialTimeout:      5 * time.Second,
		ReadTimeout:      5 * time.Second,
		WriteTimeout:     5 * time.Second,
		TLSConfig:        tlsConfig,
	}

	return redis.NewFailoverClient(opts), nil
}

// GetConnection returns an existing connection
func (m *ConnectionManager) GetConnection(id string) (*RedisConnection, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	conn, ok := m.connections[id]
	if !ok {
		return nil, fmt.Errorf("connection not found")
	}
	return conn, nil
}

// Disconnect closes a connection
func (m *ConnectionManager) Disconnect(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if conn, ok := m.connections[id]; ok {
		conn.Close()
		delete(m.connections, id)
	}
}

// Close closes the Redis connection
func (c *RedisConnection) Close() {
	if c.Client != nil {
		c.Client.Close()
	}
	if c.SSHTunnel != nil {
		c.SSHTunnel.Close()
	}
}

// SelectDB selects a database
func (c *RedisConnection) SelectDB(ctx context.Context, db int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// For cluster mode, DB selection is not supported
	if c.Config.Mode == model.ModeCluster {
		return nil
	}

	if client, ok := c.Client.(*redis.Client); ok {
		pipe := client.Pipeline()
		pipe.Select(ctx, db)
		_, err := pipe.Exec(ctx)
		if err != nil {
			return err
		}
	}

	c.CurrentDB = db
	return nil
}

// GetDatabases returns database information
func (c *RedisConnection) GetDatabases(ctx context.Context) ([]model.DatabaseInfo, error) {
	// For cluster mode, only db0 is available
	if c.Config.Mode == model.ModeCluster {
		keyCount, _ := c.Client.DBSize(ctx).Result()
		return []model.DatabaseInfo{{Index: 0, KeyCount: keyCount}}, nil
	}

	databases := make([]model.DatabaseInfo, 16)
	
	// Get keyspace info
	info, err := c.Client.Info(ctx, "keyspace").Result()
	if err == nil {
		for i := 0; i < 16; i++ {
			dbKey := fmt.Sprintf("db%d", i)
			databases[i] = model.DatabaseInfo{Index: i, KeyCount: 0}
			
			// Check if we have stored alias
			c.mu.RLock()
			if dbInfo, ok := c.DBInfo[i]; ok {
				databases[i].Alias = dbInfo.Alias
			}
			c.mu.RUnlock()
			
			// Parse keyspace info
			if strings.Contains(info, dbKey+":") {
				lines := strings.Split(info, "\n")
				for _, line := range lines {
					if strings.HasPrefix(line, dbKey+":") {
						parts := strings.Split(line, ",")
						for _, part := range parts {
							if strings.HasPrefix(part, "keys=") || strings.HasPrefix(strings.TrimPrefix(part, dbKey+":"), "keys=") {
								keyStr := strings.TrimPrefix(part, "keys=")
								keyStr = strings.TrimPrefix(keyStr, dbKey+":keys=")
								if count, err := strconv.ParseInt(keyStr, 10, 64); err == nil {
									databases[i].KeyCount = count
								}
							}
						}
					}
				}
			}
		}
	}

	return databases, nil
}

// SetDatabaseAlias sets the alias for a database
func (c *RedisConnection) SetDatabaseAlias(db int, alias string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.DBInfo[db] == nil {
		c.DBInfo[db] = &model.DatabaseInfo{Index: db}
	}
	c.DBInfo[db].Alias = alias
}

// ScanKeys scans keys with pattern
func (c *RedisConnection) ScanKeys(ctx context.Context, db int, pattern string, cursor uint64, count int64) ([]model.KeyInfo, uint64, error) {
	// Select database first
	if err := c.SelectDB(ctx, db); err != nil {
		return nil, 0, err
	}

	if pattern == "" {
		pattern = "*"
	}
	if count == 0 {
		count = 100
	}

	keys, newCursor, err := c.Client.Scan(ctx, cursor, pattern, count).Result()
	if err != nil {
		return nil, 0, err
	}

	keyInfos := make([]model.KeyInfo, 0, len(keys))
	for _, key := range keys {
		keyType, _ := c.Client.Type(ctx, key).Result()
		ttl, _ := c.Client.TTL(ctx, key).Result()
		keyInfos = append(keyInfos, model.KeyInfo{
			Key:  key,
			Type: keyType,
			TTL:  int64(ttl.Seconds()),
		})
	}

	return keyInfos, newCursor, nil
}

// ScanKeysWithRegex scans keys and filters with regex
func (c *RedisConnection) ScanKeysWithRegex(ctx context.Context, db int, regexPattern string, count int64) ([]model.KeyInfo, error) {
	if err := c.SelectDB(ctx, db); err != nil {
		return nil, err
	}

	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex: %v", err)
	}

	var cursor uint64
	keyInfos := make([]model.KeyInfo, 0)
	maxIterations := 100 // Prevent infinite loop

	for i := 0; i < maxIterations; i++ {
		keys, newCursor, err := c.Client.Scan(ctx, cursor, "*", count).Result()
		if err != nil {
			return nil, err
		}

		for _, key := range keys {
			if re.MatchString(key) {
				keyType, _ := c.Client.Type(ctx, key).Result()
				ttl, _ := c.Client.TTL(ctx, key).Result()
				keyInfos = append(keyInfos, model.KeyInfo{
					Key:  key,
					Type: keyType,
					TTL:  int64(ttl.Seconds()),
				})
			}
		}

		cursor = newCursor
		if cursor == 0 {
			break
		}
	}

	return keyInfos, nil
}

// GetKeyValue gets the value of a key
func (c *RedisConnection) GetKeyValue(ctx context.Context, db int, key string) (*model.KeyValue, error) {
	if err := c.SelectDB(ctx, db); err != nil {
		return nil, err
	}

	keyType, err := c.Client.Type(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	if keyType == "none" {
		return nil, fmt.Errorf("key not found")
	}

	ttl, _ := c.Client.TTL(ctx, key).Result()

	var value interface{}
	var size int64

	switch keyType {
	case "string":
		val, err := c.Client.Get(ctx, key).Result()
		if err != nil {
			return nil, err
		}
		value = val
		size = int64(len(val))

	case "list":
		val, err := c.Client.LRange(ctx, key, 0, -1).Result()
		if err != nil {
			return nil, err
		}
		value = val
		size = c.Client.LLen(ctx, key).Val()

	case "hash":
		val, err := c.Client.HGetAll(ctx, key).Result()
		if err != nil {
			return nil, err
		}
		value = val
		size = c.Client.HLen(ctx, key).Val()

	case "set":
		val, err := c.Client.SMembers(ctx, key).Result()
		if err != nil {
			return nil, err
		}
		value = val
		size = c.Client.SCard(ctx, key).Val()

	case "zset":
		zs, err := c.Client.ZRangeWithScores(ctx, key, 0, -1).Result()
		if err != nil {
			return nil, err
		}
		members := make([]map[string]interface{}, len(zs))
		for i, z := range zs {
			members[i] = map[string]interface{}{
				"member": z.Member,
				"score":  z.Score,
			}
		}
		value = members
		size = c.Client.ZCard(ctx, key).Val()

	case "stream":
		// 获取 Stream 消息
		msgs, err := c.Client.XRange(ctx, key, "-", "+").Result()
		if err != nil {
			return nil, err
		}
		streamData := make([]map[string]interface{}, len(msgs))
		for i, msg := range msgs {
			streamData[i] = map[string]interface{}{
				"id":     msg.ID,
				"values": msg.Values,
			}
		}
		value = streamData
		size = c.Client.XLen(ctx, key).Val()

	default:
		// 对于 bitmap, hyperloglog, geo 等特殊类型，Redis 返回的 type 是 string 或 zset
		// 需要特殊处理
		return nil, fmt.Errorf("unsupported key type: %s", keyType)
	}

	return &model.KeyValue{
		Key:   key,
		Type:  keyType,
		Value: value,
		TTL:   int64(ttl.Seconds()),
		Size:  size,
	}, nil
}

// SetKey creates or updates a key
func (c *RedisConnection) SetKey(ctx context.Context, db int, key string, keyType string, value interface{}, ttl int64) error {
	if err := c.SelectDB(ctx, db); err != nil {
		return err
	}

	// Delete existing key first
	c.Client.Del(ctx, key)

	var err error
	switch keyType {
	case "string":
		strVal, ok := value.(string)
		if !ok {
			return fmt.Errorf("invalid value type for string")
		}
		if ttl > 0 {
			err = c.Client.Set(ctx, key, strVal, time.Duration(ttl)*time.Second).Err()
		} else {
			err = c.Client.Set(ctx, key, strVal, 0).Err()
		}

	case "list":
		listVal, ok := value.([]interface{})
		if !ok {
			return fmt.Errorf("invalid value type for list")
		}
		if len(listVal) > 0 {
			err = c.Client.RPush(ctx, key, listVal...).Err()
		}

	case "hash":
		hashVal, ok := value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid value type for hash")
		}
		if len(hashVal) > 0 {
			err = c.Client.HSet(ctx, key, hashVal).Err()
		}

	case "set":
		setVal, ok := value.([]interface{})
		if !ok {
			return fmt.Errorf("invalid value type for set")
		}
		if len(setVal) > 0 {
			err = c.Client.SAdd(ctx, key, setVal...).Err()
		}

	case "zset":
		zsetVal, ok := value.([]interface{})
		if !ok {
			return fmt.Errorf("invalid value type for zset")
		}
		members := make([]redis.Z, 0)
		for _, item := range zsetVal {
			if m, ok := item.(map[string]interface{}); ok {
				score, _ := m["score"].(float64)
				member := m["member"]
				members = append(members, redis.Z{Score: score, Member: member})
			}
		}
		if len(members) > 0 {
			err = c.Client.ZAdd(ctx, key, members...).Err()
		}

	case "stream":
		// Stream: 使用 XADD 添加消息
		streamVal, ok := value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid value type for stream")
		}
		if len(streamVal) > 0 {
			args := &redis.XAddArgs{
				Stream: key,
				ID:     "*", // 自动生成 ID
				Values: streamVal,
			}
			err = c.Client.XAdd(ctx, args).Err()
		}

	case "geo":
		// Geo: 使用 GEOADD 添加地理位置
		geoVal, ok := value.([]interface{})
		if !ok {
			return fmt.Errorf("invalid value type for geo")
		}
		geoLocations := make([]*redis.GeoLocation, 0)
		for _, item := range geoVal {
			if m, ok := item.(map[string]interface{}); ok {
				name, _ := m["name"].(string)
				longitude, _ := m["longitude"].(float64)
				latitude, _ := m["latitude"].(float64)
				geoLocations = append(geoLocations, &redis.GeoLocation{
					Name:      name,
					Longitude: longitude,
					Latitude:  latitude,
				})
			}
		}
		if len(geoLocations) > 0 {
			err = c.Client.GeoAdd(ctx, key, geoLocations...).Err()
		}

	case "hyperloglog":
		// HyperLogLog: 使用 PFADD 添加元素
		hllVal, ok := value.([]interface{})
		if !ok {
			return fmt.Errorf("invalid value type for hyperloglog")
		}
		if len(hllVal) > 0 {
			err = c.Client.PFAdd(ctx, key, hllVal...).Err()
		}

	case "bitmap":
		// Bitmap: 使用 SETBIT 设置位
		bitmapVal, ok := value.([]interface{})
		if !ok {
			return fmt.Errorf("invalid value type for bitmap")
		}
		for _, item := range bitmapVal {
			if m, ok := item.(map[string]interface{}); ok {
				offset, _ := m["offset"].(float64)
				bitValue, _ := m["value"].(float64)
				if err = c.Client.SetBit(ctx, key, int64(offset), int(bitValue)).Err(); err != nil {
					return err
				}
			}
		}

	default:
		return fmt.Errorf("unsupported key type: %s", keyType)
	}

	if err != nil {
		return err
	}

	// Set TTL if specified
	if ttl > 0 && keyType != "string" {
		err = c.Client.Expire(ctx, key, time.Duration(ttl)*time.Second).Err()
	}

	return err
}

// DeleteKeys deletes keys
func (c *RedisConnection) DeleteKeys(ctx context.Context, db int, keys []string) (int64, error) {
	if err := c.SelectDB(ctx, db); err != nil {
		return 0, err
	}
	return c.Client.Del(ctx, keys...).Result()
}

// SetTTL sets TTL for a key
func (c *RedisConnection) SetTTL(ctx context.Context, db int, key string, ttl int64) error {
	if err := c.SelectDB(ctx, db); err != nil {
		return err
	}
	return c.Client.Expire(ctx, key, time.Duration(ttl)*time.Second).Err()
}

// RemoveTTL removes TTL from a key
func (c *RedisConnection) RemoveTTL(ctx context.Context, db int, key string) error {
	if err := c.SelectDB(ctx, db); err != nil {
		return err
	}
	return c.Client.Persist(ctx, key).Err()
}

// GetTTL gets TTL of a key
func (c *RedisConnection) GetTTL(ctx context.Context, db int, key string) (int64, error) {
	if err := c.SelectDB(ctx, db); err != nil {
		return 0, err
	}
	ttl, err := c.Client.TTL(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return int64(ttl.Seconds()), nil
}

// RenameKey renames a key
func (c *RedisConnection) RenameKey(ctx context.Context, db int, oldKey, newKey string) error {
	if err := c.SelectDB(ctx, db); err != nil {
		return err
	}
	return c.Client.Rename(ctx, oldKey, newKey).Err()
}

// UpdateStringValue updates a string value
func (c *RedisConnection) UpdateStringValue(ctx context.Context, db int, key string, value string) error {
	if err := c.SelectDB(ctx, db); err != nil {
		return err
	}
	
	// Get current TTL
	ttl, _ := c.Client.TTL(ctx, key).Result()
	
	if ttl > 0 {
		return c.Client.Set(ctx, key, value, ttl).Err()
	}
	return c.Client.Set(ctx, key, value, 0).Err()
}

// List operations
func (c *RedisConnection) ListPush(ctx context.Context, db int, key string, values []string, direction string) error {
	if err := c.SelectDB(ctx, db); err != nil {
		return err
	}
	
	args := make([]interface{}, len(values))
	for i, v := range values {
		args[i] = v
	}
	
	if direction == "left" {
		return c.Client.LPush(ctx, key, args...).Err()
	}
	return c.Client.RPush(ctx, key, args...).Err()
}

func (c *RedisConnection) ListSet(ctx context.Context, db int, key string, index int64, value string) error {
	if err := c.SelectDB(ctx, db); err != nil {
		return err
	}
	return c.Client.LSet(ctx, key, index, value).Err()
}

func (c *RedisConnection) ListRemove(ctx context.Context, db int, key string, index int64) error {
	if err := c.SelectDB(ctx, db); err != nil {
		return err
	}
	// Redis doesn't have direct index removal, use placeholder approach
	placeholder := "__DELETED__" + fmt.Sprintf("%d", time.Now().UnixNano())
	if err := c.Client.LSet(ctx, key, index, placeholder).Err(); err != nil {
		return err
	}
	return c.Client.LRem(ctx, key, 1, placeholder).Err()
}

// Hash operations
func (c *RedisConnection) HashSet(ctx context.Context, db int, key, field, value string) error {
	if err := c.SelectDB(ctx, db); err != nil {
		return err
	}
	return c.Client.HSet(ctx, key, field, value).Err()
}

func (c *RedisConnection) HashDelete(ctx context.Context, db int, key string, fields []string) error {
	if err := c.SelectDB(ctx, db); err != nil {
		return err
	}
	return c.Client.HDel(ctx, key, fields...).Err()
}

// Set operations
func (c *RedisConnection) SetAdd(ctx context.Context, db int, key string, members []string) error {
	if err := c.SelectDB(ctx, db); err != nil {
		return err
	}
	args := make([]interface{}, len(members))
	for i, m := range members {
		args[i] = m
	}
	return c.Client.SAdd(ctx, key, args...).Err()
}

func (c *RedisConnection) SetRemove(ctx context.Context, db int, key string, members []string) error {
	if err := c.SelectDB(ctx, db); err != nil {
		return err
	}
	args := make([]interface{}, len(members))
	for i, m := range members {
		args[i] = m
	}
	return c.Client.SRem(ctx, key, args...).Err()
}

// ZSet operations
func (c *RedisConnection) ZSetAdd(ctx context.Context, db int, key string, members []model.ZSetMember) error {
	if err := c.SelectDB(ctx, db); err != nil {
		return err
	}
	zMembers := make([]redis.Z, len(members))
	for i, m := range members {
		zMembers[i] = redis.Z{Score: m.Score, Member: m.Member}
	}
	return c.Client.ZAdd(ctx, key, zMembers...).Err()
}

func (c *RedisConnection) ZSetRemove(ctx context.Context, db int, key string, members []string) error {
	if err := c.SelectDB(ctx, db); err != nil {
		return err
	}
	args := make([]interface{}, len(members))
	for i, m := range members {
		args[i] = m
	}
	return c.Client.ZRem(ctx, key, args...).Err()
}

// GetGeoMembers gets all geo members with positions
func (c *RedisConnection) GetGeoMembers(ctx context.Context, db int, key string) ([]map[string]interface{}, error) {
	if err := c.SelectDB(ctx, db); err != nil {
		return nil, err
	}

	// Get all members from the underlying zset
	members, err := c.Client.ZRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	if len(members) == 0 {
		return []map[string]interface{}{}, nil
	}

	// Get positions for all members
	positions, err := c.Client.GeoPos(ctx, key, members...).Result()
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(members))
	for i, member := range members {
		if positions[i] != nil {
			result = append(result, map[string]interface{}{
				"name":      member,
				"longitude": positions[i].Longitude,
				"latitude":  positions[i].Latitude,
			})
		}
	}

	return result, nil
}

// GetHyperLogLogCount gets the cardinality estimate
func (c *RedisConnection) GetHyperLogLogCount(ctx context.Context, db int, key string) (int64, error) {
	if err := c.SelectDB(ctx, db); err != nil {
		return 0, err
	}
	return c.Client.PFCount(ctx, key).Result()
}

// GetBitmapInfo gets bitmap statistics
func (c *RedisConnection) GetBitmapInfo(ctx context.Context, db int, key string) (map[string]interface{}, error) {
	if err := c.SelectDB(ctx, db); err != nil {
		return nil, err
	}

	// Get bit count
	bitCount, err := c.Client.BitCount(ctx, key, nil).Result()
	if err != nil {
		return nil, err
	}

	// Get string length (bytes)
	length, err := c.Client.StrLen(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	// Get sample bits (first 64 bits)
	sample := make([]int, 0, 64)
	for i := int64(0); i < 64 && i < length*8; i++ {
		bit, _ := c.Client.GetBit(ctx, key, i).Result()
		sample = append(sample, int(bit))
	}

	return map[string]interface{}{
		"bitCount": bitCount,
		"length":   length,
		"sample":   sample,
	}, nil
}

// StreamAdd adds a message to stream
func (c *RedisConnection) StreamAdd(ctx context.Context, db int, key string, values map[string]interface{}) (string, error) {
	if err := c.SelectDB(ctx, db); err != nil {
		return "", err
	}
	
	args := &redis.XAddArgs{
		Stream: key,
		ID:     "*",
		Values: values,
	}
	return c.Client.XAdd(ctx, args).Result()
}

// StreamDelete deletes messages from stream
func (c *RedisConnection) StreamDelete(ctx context.Context, db int, key string, ids []string) (int64, error) {
	if err := c.SelectDB(ctx, db); err != nil {
		return 0, err
	}
	return c.Client.XDel(ctx, key, ids...).Result()
}

// GeoAdd adds geo locations
func (c *RedisConnection) GeoAdd(ctx context.Context, db int, key string, locations []model.GeoLocation) error {
	if err := c.SelectDB(ctx, db); err != nil {
		return err
	}
	
	geoLocations := make([]*redis.GeoLocation, len(locations))
	for i, loc := range locations {
		geoLocations[i] = &redis.GeoLocation{
			Name:      loc.Name,
			Longitude: loc.Longitude,
			Latitude:  loc.Latitude,
		}
	}
	return c.Client.GeoAdd(ctx, key, geoLocations...).Err()
}

// GeoRemove removes geo members (uses ZREM since geo is stored as zset)
func (c *RedisConnection) GeoRemove(ctx context.Context, db int, key string, members []string) error {
	if err := c.SelectDB(ctx, db); err != nil {
		return err
	}
	args := make([]interface{}, len(members))
	for i, m := range members {
		args[i] = m
	}
	return c.Client.ZRem(ctx, key, args...).Err()
}

// HyperLogLogAdd adds elements to hyperloglog
func (c *RedisConnection) HyperLogLogAdd(ctx context.Context, db int, key string, elements []interface{}) error {
	if err := c.SelectDB(ctx, db); err != nil {
		return err
	}
	return c.Client.PFAdd(ctx, key, elements...).Err()
}

// BitmapSetBit sets a bit
func (c *RedisConnection) BitmapSetBit(ctx context.Context, db int, key string, offset int64, value int) error {
	if err := c.SelectDB(ctx, db); err != nil {
		return err
	}
	return c.Client.SetBit(ctx, key, offset, value).Err()
}

// TestConnection tests a connection without storing it
func TestConnection(config model.ConnectionConfig) error {
	var redisAddr string
	var tunnel *SSHTunnel

	if config.SSH.Enabled {
		var err error
		tunnel, err = createSSHTunnel(config)
		if err != nil {
			return fmt.Errorf("SSH tunnel error: %v", err)
		}
		defer tunnel.Close()
		redisAddr = tunnel.LocalAddr
	} else {
		redisAddr = fmt.Sprintf("%s:%d", config.Host, config.Port)
	}

	var client redis.UniversalClient
	var err error

	switch config.Mode {
	case model.ModeCluster:
		client, err = createClusterClient(config, redisAddr)
	case model.ModeSentinel:
		client, err = createSentinelClient(config)
	default:
		client, err = createStandaloneClient(config, redisAddr)
	}

	if err != nil {
		return err
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return client.Ping(ctx).Err()
}
