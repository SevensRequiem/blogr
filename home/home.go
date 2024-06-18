package home

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"

	"moe-blogger/auth"
	"moe-blogger/blog"
)

func Home(c echo.Context) error {
	tmpl, err := template.ParseFiles("views/base.html", "views/home.html")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	data := map[string]interface{}{}
	data = globaldata(c)
	data["pagename"] = "Home"
	data["Blogs"] = blog.GetBlogs(12)

	err = tmpl.ExecuteTemplate(c.Response().Writer, "base.html", data)
	if err != nil {
		fmt.Println("Error executing template:", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func BlogHandler(c echo.Context) error {
	tmpl, err := template.ParseFiles("views/base.html", "views/blog.html")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	data := map[string]interface{}{}
	data = globaldata(c)
	data["pagename"] = "Blog"
	data["Blogs"] = blog.GetBlogs(12)

	err = tmpl.ExecuteTemplate(c.Response().Writer, "base.html", data)
	if err != nil {
		fmt.Println("Error executing template:", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func BlogPostHandler(c echo.Context) error {
	id := c.Param("id")
	tmpl, err := template.ParseFiles("views/base.html", "views/blogpost.html")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	data := map[string]interface{}{}
	data = globaldata(c)
	data["pagename"] = "Blog Post"
	data["Blog"] = blog.GetBlog(id)
	data["Comments"] = blog.GetComments(id)

	err = tmpl.ExecuteTemplate(c.Response().Writer, "base.html", data)
	if err != nil {
		fmt.Println("Error executing template:", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func globaldata(c echo.Context) map[string]interface{} {
	data := make(map[string]interface{})
	sess, err := session.Get("session", c)
	if err != nil {
		log.Fatal(err)
		return data
	}
	user, ok := sess.Values["user"].(auth.User)
	if ok {
		isAdmin := strings.Contains(user.Groups, "admin")
		if isAdmin {
			data["IsAdmin"] = true
		}

		data["selfUser"] = user.Username
		data["selfGroups"] = user.Groups
		data["selfDate"] = user.DateCreated
		data["selfExists"] = user.DoesExist
		data["selfID"] = user.ID
		data["selfIP"] = c.RealIP()
	} else {
		data["Exists"] = false
	}
	if auth.AdminCheck(c) {
		data["IsAdmin"] = true
	} else {
		data["IsAdmin"] = false
	}

	csrfToken := c.Get("csrf")
	if csrfToken == nil {
		log.Fatal("CSRF token not found")
	}

	data["csrf"] = csrfToken
	data["ip"] = c.RealIP()
	data["version"] = "1.4.8"
	return data
}
