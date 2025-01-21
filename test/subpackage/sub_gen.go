package subpackage

import "github.com/gin-gonic/gin"

func AutoAPI_handler_Add(c *gin.Context) {
	var r1 any
	r1 = Add(n1, n2)
	if _, ok := r1.(error); ok {
		if r1 == nil {
			c.JSON(200, gin.H{"ok": true})
		} else {
			c.JSON(200, gin.H{"ok": false, "error": r1})
		}
	} else {
		c.JSON(200, gin.H{"ok": true, "data": r1})
	}
}
