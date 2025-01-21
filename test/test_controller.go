package package_name

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

// 这是一个注释

// @AutoAPI
func TestController_int_none(a *int) {
	var r1, r2 error
	*a = 200
	c := gin.H{"ok": true}
	b := new(int)
	*b = *a
	r1, r2 := b, a
	r2 = r1
	r1 = r2
	*r1 = c["ok"].(int)
	fmt.Println("Call TestController_int_none")
}

type UserInfo struct {
	ID   uint
	Name string
}

func (x *UserInfo) Error() string {
	return x.Name
}

// @AutoAPI
func TestController_UserInfo_UserInfo(u *UserInfo) *UserInfo {
	u.Name = "Updated"
	return u
}

// TestController_UserInfo_UserInfo_2 @AutoAPI
func TestController_UserInfo_UserInfo_2(u *UserInfo, c *gin.Context) *UserInfo {
	u.Name = "Updated"
	return u
}

// @AutoAPI
func TestController_UserInfo_UserInfo_3(*gin.Context) (*UserInfo, error) {
	return &UserInfo{}, nil
}
