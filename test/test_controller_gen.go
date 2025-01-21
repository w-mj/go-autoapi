package package_name

import "github.com/gin-gonic/gin"

func AutoAPI_handler_TestController_int_none(c *gin.Context) {
	n1 := new(int)
	if e := c.BindJSON(n1); e != nil {
		c.JSON(200, gin.H{"ok": false, "error": e})
		return
	}
	TestController_int_none(n1)
	c.JSON(200, gin.H{"ok": true})
}
func AutoAPI_handler_TestController_UserInfo_UserInfo(c *gin.Context) {
	n1 := new(UserInfo)
	if e := c.BindJSON(n1); e != nil {
		c.JSON(200, gin.H{"ok": false, "error": e})
		return
	}
	var r1 any
	r1 = TestController_UserInfo_UserInfo(n1)
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
func AutoAPI_handler_TestController_UserInfo_UserInfo_2(c *gin.Context) {
	n1 := new(UserInfo)
	if e := c.BindJSON(n1); e != nil {
		c.JSON(200, gin.H{"ok": false, "error": e})
		return
	}
	n2 := c
	var r1 any
	r1 = TestController_UserInfo_UserInfo_2(n1, n2)
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
func AutoAPI_handler_TestController_UserInfo_UserInfo_3(c *gin.Context) {
	n1 := c
	var r1, r2 any
	r1, r2 = TestController_UserInfo_UserInfo_3(n1)
	if r2 == nil {
		c.JSON(200, gin.H{"ok": true, "data": r1})
	} else {
		c.JSON(200, gin.H{"ok": false, "error": r2})
	}
}
