package home

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"

	"blogr.moe/backend/blog"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

var data = make(map[string]interface{})

func Home(c echo.Context) error {
	// Get all partial templates
	partials, err := filepath.Glob("views/partials/*.html")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Add base and home templates to the list
	files := append([]string{"views/base.html", "views/home.html"}, partials...)

	// Parse all templates
	tmpl, err := template.ParseFiles(files...)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	data["PageName"] = "Home"

	GlobalData(c)

	err = tmpl.ExecuteTemplate(c.Response().Writer, "base.html", data)
	if err != nil {
		fmt.Println("Error executing template:", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func Login(c echo.Context) error {
	// Get all partial templates
	partials, err := filepath.Glob("views/partials/*.html")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Add base and home templates to the list
	files := append([]string{"views/base.html", "views/login.html"}, partials...)

	// Parse all templates
	tmpl, err := template.ParseFiles(files...)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	data["PageName"] = "Login"
	GlobalData(c)

	err = tmpl.ExecuteTemplate(c.Response().Writer, "base.html", data)
	if err != nil {
		fmt.Println("Error executing template:", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func Register(c echo.Context) error {
	// Get all partial templates
	partials, err := filepath.Glob("views/partials/*.html")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Add base and home templates to the list
	files := append([]string{"views/base.html", "views/register.html"}, partials...)

	// Parse all templates
	tmpl, err := template.ParseFiles(files...)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	data["PageName"] = "Register"
	GlobalData(c)

	err = tmpl.ExecuteTemplate(c.Response().Writer, "base.html", data)
	if err != nil {
		fmt.Println("Error executing template:", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func Dashboard(c echo.Context) error {
	// Get all partial templates
	partials, err := filepath.Glob("views/partials/*.html")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Add base and home templates to the list
	files := append([]string{"views/base.html", "views/user/dashboard.html"}, partials...)

	// Parse all templates
	tmpl, err := template.ParseFiles(files...)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	data["PageName"] = "Dashboard"
	GlobalData(c)

	err = tmpl.ExecuteTemplate(c.Response().Writer, "base.html", data)
	if err != nil {
		fmt.Println("Error executing template:", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func Premium(c echo.Context) error {
	// Get all partial templates
	partials, err := filepath.Glob("views/partials/*.html")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Add base and home templates to the list
	files := append([]string{"views/base.html", "views/premium.html"}, partials...)

	// Parse all templates
	tmpl, err := template.ParseFiles(files...)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	data["PageName"] = "Premium"
	GlobalData(c)

	err = tmpl.ExecuteTemplate(c.Response().Writer, "base.html", data)
	if err != nil {
		fmt.Println("Error executing template:", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func UserDashboard(c echo.Context) error {
	// Get all partial templates
	partials, err := filepath.Glob("views/partials/*.html")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Add base and home templates to the list
	files := append([]string{"views/base.html", "views/user/dashboard.html"}, partials...)

	// Parse all templates
	tmpl, err := template.ParseFiles(files...)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	data["PageName"] = "Dashboard"
	GlobalData(c)

	err = tmpl.ExecuteTemplate(c.Response().Writer, "base.html", data)
	if err != nil {
		fmt.Println("Error executing template:", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func SinglePost(c echo.Context, user string, id string) error {
	// Get all partial templates
	partials, err := filepath.Glob("views/partials/*.html")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Add base and home templates to the list
	files := append([]string{"views/base.html", "views/blog/single.html"}, partials...)

	// Parse all templates
	tmpl, err := template.ParseFiles(files...)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	data["PageName"] = "Post"
	post, err := blog.GetPost(c, user, id)
	data["Post"] = post
	GlobalData(c)

	err = tmpl.ExecuteTemplate(c.Response().Writer, "base.html", data)
	if err != nil {
		fmt.Println("Error executing template:", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func GlobalData(c echo.Context) map[string]interface{} {
	session, err := session.Get("session", c)
	if err != nil {
		// Handle error appropriately
		return nil
	}
	data["User"] = session.Values["user"]

	return data
}
