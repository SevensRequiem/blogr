package home

import (
	"fmt"
	"html/template"
	"net/http"

	"blogr.moe/auth"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

var data = make(map[string]interface{})

func Home(c echo.Context) error {
	tmpl, err := template.ParseFiles("views/base.html", "views/home.html")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	data["PageName"] = "Home"
	data["UserCount"], err = auth.TotalUserCount()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	GlobalData(c)

	err = tmpl.ExecuteTemplate(c.Response().Writer, "base.html", data)
	if err != nil {
		fmt.Println("Error executing template:", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func Profile(c echo.Context) error {
	tmpl, err := template.ParseFiles("views/base.html", "views/user/base.html", "views/user/profile.html")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	data["PageName"] = "Profile"
	GlobalData(c)

	err = tmpl.ExecuteTemplate(c.Response().Writer, "base.html", data)
	if err != nil {
		fmt.Println("Error executing template:", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func ProfilePosts(c echo.Context) error {
	tmpl, err := template.ParseFiles("views/base.html", "views/user/base.html", "views/user/posts.html")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	data["PageName"] = "Posts"
	GlobalData(c)

	err = tmpl.ExecuteTemplate(c.Response().Writer, "base.html", data)
	if err != nil {
		fmt.Println("Error executing template:", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func ProfileComments(c echo.Context) error {
	tmpl, err := template.ParseFiles("views/base.html", "views/user/base.html", "views/user/comments.html")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	data["PageName"] = "Comments"
	GlobalData(c)

	err = tmpl.ExecuteTemplate(c.Response().Writer, "base.html", data)
	if err != nil {
		fmt.Println("Error executing template:", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func ProfileSettings(c echo.Context) error {
	tmpl, err := template.ParseFiles("views/base.html", "views/user/base.html", "views/user/settings.html")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	data["PageName"] = "Settings"
	GlobalData(c)

	err = tmpl.ExecuteTemplate(c.Response().Writer, "base.html", data)
	if err != nil {
		fmt.Println("Error executing template:", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func ProfileSecurity(c echo.Context) error {
	tmpl, err := template.ParseFiles("views/base.html", "views/user/base.html", "views/user/security.html")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	data["PageName"] = "Security"
	GlobalData(c)

	err = tmpl.ExecuteTemplate(c.Response().Writer, "base.html", data)
	if err != nil {
		fmt.Println("Error executing template:", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func ProfileNotifications(c echo.Context) error {
	tmpl, err := template.ParseFiles("views/base.html", "views/user/base.html", "views/user/notifications.html")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	data["PageName"] = "Notifications"
	GlobalData(c)

	err = tmpl.ExecuteTemplate(c.Response().Writer, "base.html", data)
	if err != nil {
		fmt.Println("Error executing template:", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func ProfilePrivacy(c echo.Context) error {
	tmpl, err := template.ParseFiles("views/base.html", "views/user/base.html", "views/user/privacy.html")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	data["PageName"] = "Privacy"
	GlobalData(c)

	err = tmpl.ExecuteTemplate(c.Response().Writer, "base.html", data)
	if err != nil {
		fmt.Println("Error executing template:", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func ProfileDelete(c echo.Context) error {
	tmpl, err := template.ParseFiles("views/base.html", "views/user/base.html", "views/user/delete.html")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	data["PageName"] = "Delete Account"
	GlobalData(c)

	err = tmpl.ExecuteTemplate(c.Response().Writer, "base.html", data)
	if err != nil {
		fmt.Println("Error executing template:", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func Blog(c echo.Context) error {
	tmpl, err := template.ParseFiles("views/base.html", "views/blog.html")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	data["PageName"] = "Blog"
	GlobalData(c)

	err = tmpl.ExecuteTemplate(c.Response().Writer, "base.html", data)
	if err != nil {
		fmt.Println("Error executing template:", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func UserBlog(c echo.Context) error {
	tmpl, err := template.ParseFiles("views/base.html", "views/userblog.html")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	data["PageName"] = "User Blog"
	GlobalData(c)

	err = tmpl.ExecuteTemplate(c.Response().Writer, "base.html", data)
	if err != nil {
		fmt.Println("Error executing template:", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func BlogList(c echo.Context) error {
	tmpl, err := template.ParseFiles("views/base.html", "views/bloglist.html")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	data["PageName"] = "Blog List"
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
	userid := session.Values["uuid"]
	if userid == nil {
		data["User"] = nil
		return nil
	}
	// Get user from database
	user, err := auth.GetUserByID(userid.(string))
	if err != nil {
		return nil
	}
	data["User"] = user
	return nil
}
