package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Bean struct {
	gorm.Model
	ID              string `gorm:"type:varchar(255);primary_key"`
	BeanName        string `gorm:"type:varchar(100);null" json:"bean_name"`
	DescriptionBean string `gorm:"type:varchar(255);null" json:"description_name"`
	PricePerUnit    string `gorm:"type:varchar(255);null" json:"price_per_unit"`
}

type DailyBean struct {
	gorm.Model
	ID        string `gorm:"type:varchar(255);primary_key"`
	BeanID    string `gorm:"type:varchar(255);null" json:"bean_id"`
	SalePrice int    `gorm:"type:int;null" json:"sale_price"`
}

type Distributor struct {
	gorm.Model
	ID              string `gorm:"type:varchar(255);primary_key"`
	DistributorName string `gorm:"type:varchar(255);null" json:"distributor_name"`
	City            string `gorm:"type:varchar(100);null" json:"city"`
	State           string `gorm:"type:varchar(100);null" json:"state"`
	Country         string `gorm:"type:varchar(100);null" json:"country"`
	Phone           string `gorm:"type:varchar(15);null" json:"phone"`
	Email           string `gorm:"type:varchar(100);null" json:"email"`
}

type Document struct {
	gorm.Model
	ID           string `gorm:"type:varchar(255);primary_key"`
	Title        string `gorm:"type:varchar(100);null" json:"title"`
	DocumentFile string `gorm:"type:varchar(255);null" json:"document_file"`
	Author       string `gorm:"type:varchar(50);null" json:"author"`
}

type Users struct {
	gorm.Model
	ID       int    `gorm:"type:int;primary_key"`
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
	result := db.Table("beans").
		Select("beans.bean_name, beans.description_bean, daily_beans.sale_price").
		Joins("INNER JOIN daily_beans ON beans.id = daily_beans.bean_id").
		Where("daily_beans.sale_price >= ?", 0).
		Scan(&beans)

	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "error fetching beans")
	}

	if len(beans) == 0 {
		return c.JSON(http.StatusNoContent, nil) // NoContent when no beans are found
	}

	return c.JSON(http.StatusOK, beans)
}

func getCatalogs(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	var beans []Bean
	db.Find(&beans)
	return c.JSON(http.StatusOK, beans)
}

func createCatalog(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	beans := new(Bean)
	if err := c.Bind(beans); err != nil {
		return err
	}
	beans.ID = uuid.New().String()
	db.Create(&beans)
	return c.JSON(http.StatusCreated, beans)
}

func createDailyBeans(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	bean := new(DailyBean)
	if err := c.Bind(bean); err != nil {
		return err
	}
	bean.ID = uuid.New().String()
	db.Create(&bean)
	return c.JSON(http.StatusCreated, bean)
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
	distributor.ID = uuid.New().String()
	db.Create(&distributor)
	return c.JSON(http.StatusCreated, distributor)
}

func updateDistributor(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	id := c.Param("id")
	var distributor Distributor

	if err := db.First(&distributor, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Distributor not found")
		}
		return err
	}

	if err := c.Bind(&distributor); err != nil {
		return err
	}

	if err := db.Save(&distributor).Error; err != nil {
		return err
	}

	return c.JSON(http.StatusOK, distributor)
}

func uploadDocument(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	document := new(Document)
	if err := c.Bind(document); err != nil {
		return err
	}
	document.ID = uuid.New().String()
	db.Create(&document)
	return c.JSON(http.StatusCreated, document)
}

func deleteUser(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	id := c.Param("id")
	var user Bean
	db.First(&user, id)
	if user.ID == "" {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
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

func registerUsers(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	user := new(Users)
	if err := c.Bind(user); err != nil {
		return err
	}
	db.Create(&user)
	return c.JSON(http.StatusCreated, user)
}

func main() {
	dsn := "host=postgresql-140411-0.cloudclusters.net user=admin password=admin123 dbname=coffee-valley port=12539 sslmode=disable TimeZone=Asia/Jakarta"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
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
	e.POST("/catalog", createCatalog)
	e.POST("/daily-beans", createDailyBeans)
	e.GET("/distributors", getDistributors)
	e.POST("/distributor", createDistributor)
	e.PUT("/distributor/:id", updateDistributor)
	e.POST("/document", uploadDocument)
	e.DELETE("/users/:id", deleteUser)
	e.GET("/users", getUsers)
	e.POST("/users", registerUsers)

	e.Start(":8080")
}
