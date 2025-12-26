package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"redis-web-manager/model"
	"redis-web-manager/service"
)

// ListKeys returns keys matching pattern
func ListKeys(c *gin.Context) {
	id := c.Param("id")

	conn, err := service.GetManager().GetConnection(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.KeysResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	db, _ := strconv.Atoi(c.DefaultQuery("db", "0"))
	pattern := c.DefaultQuery("pattern", "*")
	cursor, _ := strconv.ParseUint(c.DefaultQuery("cursor", "0"), 10, 64)
	count, _ := strconv.ParseInt(c.DefaultQuery("count", "100"), 10, 64)

	keys, newCursor, err := conn.ScanKeys(c.Request.Context(), db, pattern, cursor, count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.KeysResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.KeysResponse{
		Success: true,
		Keys:    keys,
		Cursor:  newCursor,
	})
}

// ScanKeys scans keys with advanced options
func ScanKeys(c *gin.Context) {
	id := c.Param("id")

	conn, err := service.GetManager().GetConnection(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.KeysResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	var req model.ScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.KeysResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if req.ScanCount == 0 {
		req.ScanCount = 100
	}

	var keys []model.KeyInfo
	var cursor uint64

	switch req.SearchType {
	case "regex":
		keys, err = conn.ScanKeysWithRegex(c.Request.Context(), req.DB, req.RegexPattern, req.ScanCount)
		cursor = 0
	case "keys":
		// Use KEYS command (not recommended for production)
		keys, cursor, err = conn.ScanKeys(c.Request.Context(), req.DB, req.Pattern, 0, 10000)
	default:
		keys, cursor, err = conn.ScanKeys(c.Request.Context(), req.DB, req.Pattern, 0, req.ScanCount)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, model.KeysResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.KeysResponse{
		Success: true,
		Keys:    keys,
		Cursor:  cursor,
	})
}

// GetKey returns key value
func GetKey(c *gin.Context) {
	id := c.Param("id")
	key := c.Param("key")

	conn, err := service.GetManager().GetConnection(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	db, _ := strconv.Atoi(c.DefaultQuery("db", "0"))

	keyValue, err := conn.GetKeyValue(c.Request.Context(), db, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    keyValue,
	})
}

// CreateKey creates a new key
func CreateKey(c *gin.Context) {
	id := c.Param("id")

	conn, err := service.GetManager().GetConnection(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	var req model.CreateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if err := conn.SetKey(c.Request.Context(), req.DB, req.Key, req.Type, req.Value, req.TTL); err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
	})
}

// UpdateKey updates a key value
func UpdateKey(c *gin.Context) {
	id := c.Param("id")
	key := c.Param("key")

	conn, err := service.GetManager().GetConnection(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	var req model.UpdateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Get current key type
	keyValue, err := conn.GetKeyValue(c.Request.Context(), req.DB, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Update based on type
	if keyValue.Type == "string" {
		strVal, ok := req.Value.(string)
		if !ok {
			c.JSON(http.StatusBadRequest, model.APIResponse{
				Success: false,
				Error:   "invalid value type",
			})
			return
		}
		if err := conn.UpdateStringValue(c.Request.Context(), req.DB, key, strVal); err != nil {
			c.JSON(http.StatusInternalServerError, model.APIResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}
	} else {
		// For other types, recreate the key
		ttl := keyValue.TTL
		if ttl < 0 {
			ttl = -1
		}
		if err := conn.SetKey(c.Request.Context(), req.DB, key, keyValue.Type, req.Value, ttl); err != nil {
			c.JSON(http.StatusInternalServerError, model.APIResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
	})
}

// DeleteKeys deletes keys
func DeleteKeys(c *gin.Context) {
	id := c.Param("id")

	conn, err := service.GetManager().GetConnection(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.DeleteResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	var req model.DeleteKeysRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.DeleteResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	deleted, err := conn.DeleteKeys(c.Request.Context(), req.DB, req.Keys)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.DeleteResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.DeleteResponse{
		Success: true,
		Deleted: deleted,
	})
}

// SetKeyTTL sets TTL for a key
func SetKeyTTL(c *gin.Context) {
	id := c.Param("id")
	key := c.Param("key")

	conn, err := service.GetManager().GetConnection(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	var req model.TTLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if err := conn.SetTTL(c.Request.Context(), req.DB, key, req.TTL); err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
	})
}

// RemoveKeyTTL removes TTL from a key
func RemoveKeyTTL(c *gin.Context) {
	id := c.Param("id")
	key := c.Param("key")

	conn, err := service.GetManager().GetConnection(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	var req struct {
		DB int `json:"db"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if err := conn.RemoveTTL(c.Request.Context(), req.DB, key); err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
	})
}

// GetKeyTTL gets TTL of a key
func GetKeyTTL(c *gin.Context) {
	id := c.Param("id")
	key := c.Param("key")

	conn, err := service.GetManager().GetConnection(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.TTLResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	db, _ := strconv.Atoi(c.DefaultQuery("db", "0"))

	ttl, err := conn.GetTTL(c.Request.Context(), db, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.TTLResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.TTLResponse{
		Success: true,
		TTL:     ttl,
	})
}

// RenameKey renames a key
func RenameKey(c *gin.Context) {
	id := c.Param("id")
	key := c.Param("key")

	conn, err := service.GetManager().GetConnection(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	var req model.RenameKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if err := conn.RenameKey(c.Request.Context(), req.DB, key, req.NewKey); err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
	})
}

// ListPush pushes values to list
func ListPush(c *gin.Context) {
	id := c.Param("id")
	key := c.Param("key")

	conn, err := service.GetManager().GetConnection(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	var req model.ListPushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if err := conn.ListPush(c.Request.Context(), req.DB, key, req.Values, req.Direction); err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
	})
}

// ListSet sets list element at index
func ListSet(c *gin.Context) {
	id := c.Param("id")
	key := c.Param("key")
	indexStr := c.Param("index")

	index, err := strconv.ParseInt(indexStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "invalid index",
		})
		return
	}

	conn, err := service.GetManager().GetConnection(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	var req model.ListSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if err := conn.ListSet(c.Request.Context(), req.DB, key, index, req.Value); err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
	})
}

// ListRemove removes list element at index
func ListRemove(c *gin.Context) {
	id := c.Param("id")
	key := c.Param("key")
	indexStr := c.Param("index")

	index, err := strconv.ParseInt(indexStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "invalid index",
		})
		return
	}

	conn, err := service.GetManager().GetConnection(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	var req struct {
		DB int `json:"db"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if err := conn.ListRemove(c.Request.Context(), req.DB, key, index); err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
	})
}

// HashSet sets hash field
func HashSet(c *gin.Context) {
	id := c.Param("id")
	key := c.Param("key")
	field := c.Param("field")

	conn, err := service.GetManager().GetConnection(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	var req model.HashSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if err := conn.HashSet(c.Request.Context(), req.DB, key, field, req.Value); err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
	})
}

// HashDelete deletes hash fields
func HashDelete(c *gin.Context) {
	id := c.Param("id")
	key := c.Param("key")

	conn, err := service.GetManager().GetConnection(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	var req model.HashDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if err := conn.HashDelete(c.Request.Context(), req.DB, key, req.Fields); err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
	})
}

// SetAdd adds members to set
func SetAdd(c *gin.Context) {
	id := c.Param("id")
	key := c.Param("key")

	conn, err := service.GetManager().GetConnection(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	var req model.SetMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if err := conn.SetAdd(c.Request.Context(), req.DB, key, req.Members); err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
	})
}

// SetRemove removes members from set
func SetRemove(c *gin.Context) {
	id := c.Param("id")
	key := c.Param("key")

	conn, err := service.GetManager().GetConnection(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	var req model.SetMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if err := conn.SetRemove(c.Request.Context(), req.DB, key, req.Members); err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
	})
}

// ZSetAdd adds members to zset
func ZSetAdd(c *gin.Context) {
	id := c.Param("id")
	key := c.Param("key")

	conn, err := service.GetManager().GetConnection(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	var req model.ZSetAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if err := conn.ZSetAdd(c.Request.Context(), req.DB, key, req.Members); err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
	})
}

// ZSetRemove removes members from zset
func ZSetRemove(c *gin.Context) {
	id := c.Param("id")
	key := c.Param("key")

	conn, err := service.GetManager().GetConnection(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	var req model.ZSetRemoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if err := conn.ZSetRemove(c.Request.Context(), req.DB, key, req.Members); err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
	})
}
