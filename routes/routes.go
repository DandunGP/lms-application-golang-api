package routes

import (
	"lms-application/controllers"

	"github.com/labstack/echo"
)

func New() *echo.Echo {
	e := echo.New()

	eRegister := e.Group("registration")
	eRegister.POST("/registration_step_one", controllers.RegistrationStepOne)
	eRegister.GET("/email_verification", controllers.RegistrationStepTwo)

	return e
}