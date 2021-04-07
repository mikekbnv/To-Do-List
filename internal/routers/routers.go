package routers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/mikekbnv/To-Do-List/internal/model"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

var Collections *mongo.Collection
var ctx context.Context

func init() {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://admin:admin@cluster0.0alta.mongodb.net/tasks?retryWrites=true&w=majority"))
	if err != nil {
		fmt.Println(err)
	}
	ctx, cencel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cencel()
	err = client.Connect(ctx)
	if err != nil {
		fmt.Println(err)
	}
	Collections = client.Database("tasks").Collection("list")
}
func Register(e *echo.Echo) {
	e.Add("GET", "/", alltasks, middleware.Logger())
	e.Add("POST", "/", createtask, middleware.Logger())
	e.Add("POST", "/delete", deletetask, middleware.Logger())

	e.Add("GET", "/login", login, middleware.Logger())
	e.Add("GET", "/signup", signup, middleware.Logger())
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
	cur, err := Collections.Find(context.Background(), bson.M{})
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
	_, err := Collections.InsertOne(context.Background(), model.Task{ID: primitive.NewObjectID(), Name: task})
	if err != nil {
		log.Fatal("InsertOne() ERROR:", err)
	}
	return nil
}

func deletetask(c echo.Context) error {
	id := getid(c.Request().PostFormValue("id"))
	_, err := Collections.DeleteOne(context.Background(), bson.M{"_id": id})
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
	return c.Render(http.StatusOK, "login", map[string]interface{}{})
}

func signup(c echo.Context) error {
	return c.Render(http.StatusOK, "signup", map[string]interface{}{})
}
