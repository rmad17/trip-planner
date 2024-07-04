package controllers

import (
	"net/http"
	"triplanner/core"
	"triplanner/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func CreateUser(c *gin.Context) {

	var authInput models.AuthInput

	if err := c.ShouldBindJSON(&authInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userFound models.User
	core.DB.Where("username=?", authInput.Username).Find(&userFound)

	// if userFound.if ) {
	//     c.JSON(http.StatusBadRequest, gin.H{"error": "username already used"})
	//     return
	// }

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(authInput.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := models.User{
		Username: authInput.Username,
		Password: string(passwordHash),
	}

	core.DB.Create(&user)

	c.JSON(http.StatusOK, gin.H{"data": user})

}
