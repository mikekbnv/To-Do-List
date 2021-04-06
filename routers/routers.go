package routers

import (
	"To-Do-List/model"
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

var Collections *mongo.Collection
var con context.Context

func init() {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://admin:admin@cluster0.0alta.mongodb.net/tasks?retryWrites=true&w=majority"))
	if err != nil {
		fmt.Println(err)
	}
	ctx, cencel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cencel()
	con = ctx
	err = client.Connect(ctx)
	if err != nil {
		fmt.Println(err)
	}
	db := client.Database("tasks")
	//collection := db.Collection("list")
	Collections = db.Collection("list")

}
func Register(e *echo.Echo) {
	e.Add("GET", "/", alltasks, middleware.Logger())
	e.Add("POST", "/", createtask, middleware.Logger())
	e.Add("POST", "/delete", deletetask, middleware.Logger())
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

func createtask(c echo.Context) error {
	submit := c.Request().PostFormValue("name")
	if submit != "" {
		_, err := Collections.InsertOne(context.Background(), model.Task{ID: primitive.NewObjectID(), Name: submit})
		if err != nil {
			fmt.Println(err)
		}
	}
	return c.Redirect(http.StatusFound, "/")
}

func getAll() ([]*model.Task, error) {
	var tasks []*model.Task
	cur, err := Collections.Find(context.Background(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	for cur.Next(con) {
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
	cur.Close(con)

	if len(tasks) == 0 {
		return tasks, mongo.ErrNoDocuments
	}

	return tasks, nil

}

func deletetask(c echo.Context) error {
	id := c.Request().PostFormValue("name")
	start := strings.Index(id, "\"")
	tmp := id[start+1 : len(id)-2]
	idPrimitive, err := primitive.ObjectIDFromHex(tmp)
	if err != nil {
		log.Fatal("primitive.ObjectIDFromHex ERROR:", err)
	}
	_, err = Collections.DeleteOne(context.Background(), bson.M{"_id": idPrimitive})
	if err != nil {
		log.Fatal("DeleteOne() ERROR:", err)
	}
	return c.Redirect(http.StatusFound, "/")
}
