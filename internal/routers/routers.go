package routers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/mikekbnv/To-Do-List/database"
	"github.com/mikekbnv/To-Do-List/helper"
	"github.com/mikekbnv/To-Do-List/internal/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2/bson"
)

type Message struct {
	Errors map[string]string
}

const Token_Cookie_Name = "token"
const Refresh_Token_Cookie_Name = "refresh_token"

var usersCollection *mongo.Collection = database.OpenCollection("users")
var tasksCollection *mongo.Collection = database.OpenCollection("tasks")

//GET tasks from Database
func Get_List(c echo.Context) error {
	clientname := c.Get("first_name")
	id := c.Get("uid")
	id_str := fmt.Sprint(id)
	list, err := get_All_Tasks(id_str, "valid")
	if err != nil {
		log.Println("error with getting tasks")
	}
	return c.Render(http.StatusOK, "index", map[string]interface{}{
		"User": clientname,
		"List": list,
	})
}
func get_All_Tasks(user_id interface{}, key string) ([]*model.Task, error) { //Helper for getting tasks
	var tasks []*model.Task

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	cur, err := tasksCollection.Find(context.Background(), bson.M{"user_id": user_id})
	if err != nil {
		log.Fatal(err)
	}
	for cur.Next(ctx) {
		var t model.Task
		_ = cur.Decode(&t)
		if err != nil {
			return tasks, err
		}
		if key == "all" && t.Status {
			tasks = append(tasks, &t)
		} else if key == "valid" && !t.Status {
			tasks = append([]*model.Task{&t}, tasks...)
		}
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

//Ending for getting tasks

//getting tasks for the whole period
func All_tasks(c echo.Context) error {
	clientname := c.Get("first_name")
	id := c.Get("uid")
	id_str := fmt.Sprint(id)
	list, err := get_All_Tasks(id_str, "all")
	if err != nil {
		log.Println("error with getting tasks")
	}
	return c.Render(http.StatusOK, "alltasks", map[string]interface{}{
		"User": clientname,
		"List": list,
	})
}

//Undo task if it was deleted
func Undo_task(c echo.Context) error {
	ID := getid(c.Request().FormValue("id"))
	//log.Println("TASK ID:", ID)

	upsert := true
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}
	update := bson.M{
		"$set": bson.M{"status": false},
	}
	_, err := tasksCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": ID},
		update,
		&opt,
	)
	if err != nil {
		log.Fatal("Undo error:", err)
	}

	return c.Redirect(http.StatusFound, "/alltasks")
}

//Adding and creating task to Database
func Createtask(c echo.Context) error {
	id := c.Get("uid")
	id_str := fmt.Sprint(id)
	task := c.Request().PostFormValue("task")
	if task != "" {
		add_To_Db(task, id_str)

	}
	return c.Redirect(http.StatusFound, "/list")
}

func add_To_Db(task, id string) { //Helper func for adding to Database
	_, err := tasksCollection.InsertOne(context.Background(), model.Task{ID: primitive.NewObjectID(), Name: task, User_Id: id})
	if err != nil {
		log.Fatal("InsertOne() ERROR:", err)
	}
}

//Ending for adding task

//Deleting task from Database
func Deletetask(c echo.Context) error {
	ID := getid(c.Request().FormValue("id"))
	//log.Println("TASK ID:", ID)

	upsert := true
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}
	update := bson.M{
		"$set": bson.M{"status": true},
	}
	_, err := tasksCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": ID},
		update,
		&opt,
	)
	if err != nil {
		log.Fatal("DeleteOne() ERROR:", err)
	}
	return c.Redirect(http.StatusFound, "/list")
}

//Signup form and Handler that render signup template
func Signup_Form(c echo.Context) error { //Get method
	c = clearcookie(c)
	return c.Render(http.StatusOK, "signup", map[string]interface{}{})
}
func Signup(c echo.Context) error { //Post method
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	var user model.User
	var msg Message
	msg.Errors = make(map[string]string)

	defer cancel()

	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]error{"error": err})
	}
	if *user.First_name == "" {
		msg.Errors = errors(msg.Errors, "First_name", "First name cannot be empty")
	}
	if *user.Last_name == "" {
		msg.Errors = errors(msg.Errors, "Last_name", "Last name cannot be empty")
	}
	if *user.Email == "" {
		msg.Errors = errors(msg.Errors, "Email", "Email cannot be empty")
	}
	confirmation_pass := c.Request().FormValue("confirmation-password")
	if *user.Password == "" && confirmation_pass == "" {
		msg.Errors = errors(msg.Errors, "Password", "Password fields cannot be empty")
	}
	count, err := usersCollection.CountDocuments(ctx, bson.M{"email": user.Email})
	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": err})
	}
	if count > 0 {
		msg.Errors = errors(msg.Errors, "Email", "User with this email alredy exist")
	}
	if confirmation_pass != *user.Password {
		msg.Errors = errors(msg.Errors, "Password", "Passwords did not match")
	}
	password := HashPassword(*user.Password)
	user.Password = &password
	user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	user.ID = primitive.NewObjectID()
	user.User_id = user.ID.Hex()
	token, refreshToken, _ := helper.GenerateAllTokens()
	user.Token = &token
	user.Refresh_token = &refreshToken
	if len(msg.Errors) != 0 {
		//log.Println(msg)
		c.Render(http.StatusBadRequest, "signup", msg)
	}
	_, insertErr := usersCollection.InsertOne(ctx, model.User{
		ID:            user.ID,
		First_name:    user.First_name,
		Last_name:     user.Last_name,
		Password:      user.Password,
		Email:         user.Email,
		Token:         user.Token,
		Refresh_token: user.Refresh_token,
		Created_at:    user.Created_at,
		Updated_at:    user.Updated_at,
		User_id:       user.User_id,
	})
	if insertErr != nil {
		msg := "User was not created"
		return c.JSON(http.StatusNotImplemented, map[string]interface{}{"error": msg})
	}
	defer cancel()
	return c.Redirect(http.StatusFound, "/login")
}

//Ending for signup

//Login form and Handler that render login template
func Login_Form(c echo.Context) error { //Get methon for login
	c = clearcookie(c)
	return c.Render(http.StatusOK, "login", map[string]interface{}{})
}
func Login(c echo.Context) error { //Post method for login
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	var foundUser model.User
	var msg Message

	email := c.Request().FormValue("email")
	pass := c.Request().FormValue("Password")

	msg.Errors = make(map[string]string)
	if email == "" {
		msg.Errors["Email"] = "Please enter an email address"
		if pass == "" {
			msg.Errors["Password"] = "Please enter a password"
			return c.Render(http.StatusBadRequest, "login", msg)
		}
		return c.Render(http.StatusBadRequest, "login", msg)
	} else if pass == "" {
		msg.Errors["Password"] = "Please enter a password"
		return c.Render(http.StatusBadRequest, "login", msg)
	}

	err := usersCollection.FindOne(ctx, bson.M{"email": email}).Decode(&foundUser)
	if err != nil {
		return c.Redirect(http.StatusFound, "/login")
	}
	passwordIsValid := VerifyPassword(pass, *foundUser.Password)
	if !passwordIsValid {
		msg.Errors["Error"] = "Email or password is incorrect"
		return c.Render(http.StatusBadRequest, "login", msg)
	}
	token, refreshToken, _ := helper.GenerateAllTokens()
	helper.UpdateAllTokens(token, refreshToken, foundUser.User_id)

	cookiestoken := &http.Cookie{
		Name:  Token_Cookie_Name,
		Value: token,
	}
	cookiesrefresh := &http.Cookie{
		Name:  Refresh_Token_Cookie_Name,
		Value: refreshToken,
	}

	c.SetCookie(cookiestoken)
	c.SetCookie(cookiesrefresh)
	return c.Redirect(http.StatusFound, "/list")
}

//Ending for login

//Logout handler for exit
func Logout(c echo.Context) error { //Post for logout
	c = clearcookie(c)
	return c.Redirect(http.StatusFound, "/login")
}
func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}

	return string(bytes)
}

func VerifyPassword(userPassword string, providedPassword string) bool {
	fmt.Println()
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true

	if err != nil {
		check = false
	}

	return check
}
func clearcookie(c echo.Context) echo.Context { //Method for clearing user cookie after logout
	c.SetCookie(&http.Cookie{
		Name:  "token",
		Value: "",
	})
	return c
}

func getid(object string) primitive.ObjectID { //Method for parcing mongoDB primitive.ObjectI to string ID
	start := strings.Index(object, "\"")
	id := object[start+1 : len(object)-2]
	idPrimitive, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Fatal("primitive.ObjectIDFromHex ERROR:", err)
	}
	return idPrimitive
}

func errors(msg map[string]string, field, msg_error string) map[string]string { //Method for helping to collect error from login or signup handlerss
	msg[field] = msg_error
	return msg
}
