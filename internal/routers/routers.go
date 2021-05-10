package routers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/mikekbnv/To-Do-List/database"
	"github.com/mikekbnv/To-Do-List/internal/model"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "users")

func Register(e *echo.Echo) {
	e.Add("GET", "/", alltasks, middleware.Logger())
	e.Add("POST", "/", createtask, middleware.Logger())
	e.Add("POST", "/delete", deletetask, middleware.Logger())

	e.Add("GET", "/login", login, middleware.Logger())
	//e.Add("POST", "/login", getdata, middleware.Logger())

	//e.Add("GET", "/signup", signup, middleware.Logger())
}

func alltasks(c echo.Context) error {
	tasks, err := getAll()
	if err != nil {
		fmt.Println(err)
	}
	return c.Render(http.StatusOK, "index", map[string]interface{}{
		"List": tasks,
		"User": "Mike",
	})
}
func getAll() ([]*model.Task, error) {
	var tasks []*model.Task

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	cur, err := userCollection.Find(context.Background(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	for cur.Next(ctx) {
		var t model.Task
		err := cur.Decode(&t)
		if err != nil {
			return tasks, err
		}
		tasks = append(tasks, &t)
	}

	if err := cur.Err(); err != nil {
		return tasks, err
	}
	cur.Close(ctx)

	if len(tasks) == 0 {
		return tasks, mongo.ErrNoDocuments
	}
	return tasks, nil
}

func createtask(c echo.Context) error {
	task := c.Request().PostFormValue("task")
	if task != "" {
		addtodb(task)
	}
	return c.Redirect(http.StatusFound, "/")
}

func addtodb(task string) error {
	_, err := userCollection.InsertOne(context.Background(), model.Task{ID: primitive.NewObjectID(), Name: task})
	if err != nil {
		log.Fatal("InsertOne() ERROR:", err)
	}
	return nil
}

func deletetask(c echo.Context) error {
	id := getid(c.Request().PostFormValue("id"))
	_, err := userCollection.DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		log.Fatal("DeleteOne() ERROR:", err)
	}
	return c.Redirect(http.StatusFound, "/")
}

func getid(object string) primitive.ObjectID {
	start := strings.Index(object, "\"")
	tmp := object[start+1 : len(object)-2]
	idPrimitive, err := primitive.ObjectIDFromHex(tmp)
	if err != nil {
		log.Fatal("primitive.ObjectIDFromHex ERROR:", err)
	}
	return idPrimitive
}

func login(c echo.Context) error {
	return c.Render(http.StatusOK, "login1", map[string]interface{}{})
}

// func signup(c echo.Context) error {
// 	return c.Render(http.StatusOK, "signup", map[string]interface{}{})
//}

// func getdata(c echo.Context) error {
// 	name := c.Request().PostFormValue("name")
 	// pass := c.Request().PostFormValue("pass")
// 	return c.Redirect(http.StatusFound, "/")
//}
