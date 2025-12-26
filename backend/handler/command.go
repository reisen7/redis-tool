package handler

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"redis-web-manager/model"
	"redis-web-manager/service"
)

// ExecuteCommand executes a Redis command
func ExecuteCommand(c *gin.Context) {
	id := c.Param("id")

	conn, err := service.GetManager().GetConnection(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.CommandResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	var req model.CommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.CommandResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Select database
	if err := conn.SelectDB(c.Request.Context(), req.DB); err != nil {
		c.JSON(http.StatusInternalServerError, model.CommandResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Parse command
	args := parseCommand(req.Command)
	if len(args) == 0 {
		c.JSON(http.StatusBadRequest, model.CommandResponse{
			Success: false,
			Error:   "empty command",
		})
		return
	}

	// Execute command with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// Convert to interface slice for Do command
	cmdArgs := make([]interface{}, len(args))
	for i, arg := range args {
		cmdArgs[i] = arg
	}

	result, err := conn.Client.Do(ctx, cmdArgs...).Result()
	if err != nil {
		c.JSON(http.StatusOK, model.CommandResponse{
			Success: true,
			Result:  formatError(err),
		})
		return
	}

	c.JSON(http.StatusOK, model.CommandResponse{
		Success: true,
		Result:  formatResult(result),
	})
}

// parseCommand parses a Redis command string into arguments
func parseCommand(cmd string) []string {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return nil
	}

	var args []string
	var current strings.Builder
	inQuote := false
	quoteChar := rune(0)

	for _, ch := range cmd {
		switch {
		case ch == '"' || ch == '\'':
			if inQuote && ch == quoteChar {
				inQuote = false
				quoteChar = 0
			} else if !inQuote {
				inQuote = true
				quoteChar = ch
			} else {
				current.WriteRune(ch)
			}
		case ch == ' ' && !inQuote:
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}

// formatResult formats the command result for display
func formatResult(result interface{}) string {
	switch v := result.(type) {
	case nil:
		return "(nil)"
	case string:
		return "\"" + v + "\""
	case int64:
		return formatInt(v)
	case []interface{}:
		return formatArray(v)
	case []byte:
		return "\"" + string(v) + "\""
	default:
		return formatDefault(v)
	}
}

func formatInt(v int64) string {
	return "(integer) " + strings.TrimSpace(strings.Replace(strings.Replace(strings.Replace(
		strings.Replace(strings.Replace(strings.Replace(
			strings.Replace(strings.Replace(strings.Replace(
				strings.Replace(
					"          "+string(rune('0'+v%10)),
					"          ", "", 1),
				"         ", "", 1),
			"        ", "", 1),
		"       ", "", 1),
	"      ", "", 1),
	"     ", "", 1),
	"    ", "", 1),
	"   ", "", 1),
	"  ", "", 1),
	" ", "", 1))
}

func formatArray(arr []interface{}) string {
	if len(arr) == 0 {
		return "(empty array)"
	}
	var sb strings.Builder
	for i, item := range arr {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(formatArrayItem(i+1, item))
	}
	return sb.String()
}

func formatArrayItem(index int, item interface{}) string {
	var sb strings.Builder
	sb.WriteString(strings.Repeat(" ", 0))
	sb.WriteString(string(rune('0' + index/10)))
	sb.WriteString(string(rune('0' + index%10)))
	sb.WriteString(") ")
	switch v := item.(type) {
	case nil:
		sb.WriteString("(nil)")
	case string:
		sb.WriteString("\"" + v + "\"")
	case int64:
		sb.WriteString("(integer) ")
		sb.WriteString(formatIntSimple(v))
	case []byte:
		sb.WriteString("\"" + string(v) + "\"")
	default:
		sb.WriteString(formatDefault(v))
	}
	return sb.String()
}

func formatIntSimple(v int64) string {
	if v == 0 {
		return "0"
	}
	if v < 0 {
		return "-" + formatIntSimple(-v)
	}
	var digits []byte
	for v > 0 {
		digits = append([]byte{byte('0' + v%10)}, digits...)
		v /= 10
	}
	return string(digits)
}

func formatDefault(v interface{}) string {
	switch val := v.(type) {
	case int64:
		return "(integer) " + formatIntSimple(val)
	case string:
		return "\"" + val + "\""
	case []byte:
		return "\"" + string(val) + "\""
	default:
		return "(unknown type)"
	}
}

func formatError(err error) string {
	return "(error) " + err.Error()
}
