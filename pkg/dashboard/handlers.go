package dashboard

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func Error(err error)  {
	 if err!=nil{
	 	fmt.Println(err)
	 	panic(err)
	 }
}
func errorHandler() gin.HandlerFunc{
	gin.ErrorLogger()
	return func(c *gin.Context) {
		 defer func() {
		 	if e:=recover();e!=nil{
				c.AbortWithStatusJSON(400, gin.H{"error":e.(error).Error()})
			}
		 }()
		 c.Next()
	}
}

