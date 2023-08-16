package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Bean struct {
	gorm.Model
	ID              int    `gorm:"type:int;primary_key"`
	BeanName        string `gorm:"type:varchar(100);null" json:"bean_name"`
	DescriptionBean string `gorm:"type:varchar(255);null" json:"description_name"`
	PricePerUnit    string `gorm:"type:varchar(255);null" json:"price_per_unit"`
}

type DailyBean struct {
	gorm.Model
	ID        int    `gorm:"type:int;primary_key"`
	BeanID    int    `gorm:"type:int;bean_id"`
	SalePrice string `gorm:"type:varchar(255);null" json:"sale_price"`
}

type Distributor struct {
	gorm.Model
	ID              int    `gorm:"type:int;primary_key"`
	DistributorName string `gorm:"type:varchar(255);null" json:"distributor_name"`
	City            string `gorm:"type:varchar(100);null" json:"city"`
	State           string `gorm:"type:varchar(100);null" json:"state"`
	Country         string `gorm:"type:varchar(100);null" json:"country"`
	Phone           string `gorm:"type:varchar(15);null" json:"phone"`
	Email           string `gorm:"type:varchar(100);null" json:"email"`
}

type Document struct {
	gorm.Model
	ID           int    `gorm:"type:int;primary_key"`
	Title        string `gorm:"type:varchar(100);null" json:"title"`
	DocumentFile string `gorm:"type:varchar(255);null" json:"document_file"`
	Author       string `gorm:"type:varchar(50);null" json:"author"`
}

type Users struct {
	gorm.Model
	ID       int    `gorm:"type:string;primary_key"`
	FullName string `gorm:"type:varchar(100);null" json:"fullname"`
	Email    string `gorm:"type:varchar(255);null" json:"email"`
	Password string `gorm:"type:varchar(50);null" json:"password"`
}

func getBean(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	var beans []struct {
		BeanName        string `json:"bean_name"`
		DescriptionBean string `json:"description_bean"`
		SalePrice       string `json:"sale_price"`
	}
	db.Table("beans").
		Select("beans.bean_name, beans.description_bean, daily_beans.sale_price").
		Joins("INNER JOIN daily_beans ON beans.id = daily_beans.bean_id").
		Where("daily_beans.sale_price >= ?", 0).
		Scan(&beans)

	if len(beans) == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "beans not found")
	}

	return c.JSON(http.StatusOK, beans)
}

func getCatalogs(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	var beans []Bean
	db.Find(&beans)
	return c.JSON(http.StatusOK, beans)
}

func getDistributors(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	var distributors []Distributor
	db.Find(&distributors)
	return c.JSON(http.StatusOK, distributors)
}

func createDistributor(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	distributor := new(Distributor)
	if err := c.Bind(distributor); err != nil {
		return err
	}
	db.Create(&distributor)
	return c.JSON(http.StatusCreated, distributor)
}

func updateDistributor(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	id := c.Param("id")
	var distributor Distributor
	db.First(&distributor, id)
	if distributor.ID == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "distributor not found")
	}
	if err := c.Bind(&distributor); err != nil {
		return err
	}
	db.Save(&distributor)
	return c.JSON(http.StatusOK, distributor)
}

func uploadDocument(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	document := new(Document)
	if err := c.Bind(document); err != nil {
		return err
	}
	db.Create(&document)
	return c.JSON(http.StatusCreated, document)
}

func deleteUser(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	id := c.Param("id")
	var user Bean
	db.First(&user, id)
	if user.ID == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}
	db.Delete(&user)
	return c.NoContent(http.StatusNoContent)
}

func getUsers(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	var user []Users
	db.Find(&user)
	return c.JSON(http.StatusOK, user)
}

func main() {
	dsn := "root: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@tcp(127.0.0.1:3306)/simple-app?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println(err)
		return
	}
	db.AutoMigrate(&Bean{}, &Distributor{}, &DailyBean{}, &Document{}, &Users{})

	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("db", db)
			return next(c)
		}
	})
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	e.GET("/bean", getBean)
	e.GET("/catalogs", getCatalogs)
	e.GET("/distributors", getDistributors)
	e.POST("/distributor", createDistributor)
	e.PUT("/distributor/:id", updateDistributor)
	e.POST("/document", uploadDocument)
	e.DELETE("/users/:id", deleteUser)
	e.GET("/users", getUsers)

	e.Start(":8080")
}
