package routers

import (
	"context"
	"fmt"
	"log"
	"net/http"
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
	Inputed_info map[string]string
	Errors       map[string]string
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

//Helper for getting tasks
func get_All_Tasks(user_id interface{}, key string) ([]*model.Task, error) {
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
	ID, err := get_user_by_id(c.Get("uid"))
	if err != nil {
		log.Fatal("Error with getting user by id")
	}

	upsert := true
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}
	update := bson.M{
		"$set": bson.M{"status": false},
	}
	_, err = tasksCollection.UpdateOne(
		context.Background(),
		bson.M{"user_id": ID},
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

//Helper func for adding to Database
func add_To_Db(task, id string) {
	Id := primitive.NewObjectID()
	task_id := Id.Hex()
	_, err := tasksCollection.InsertOne(context.Background(), model.Task{
		ID:      Id,
		Name:    task,
		User_Id: id,
		Task_id: task_id,
	})
	if err != nil {
		log.Fatal("InsertOne() ERROR:", err)
	}
}

//Deleting task from Database
func Deletetask(c echo.Context) error {

	ID := c.Request().FormValue("task_id")
	//log.Println("TASK ID:", ID)
	err := update_task_by_id(ID, "status", true)
	if err != nil {
		log.Fatal("Update status() ERROR:", err)
	}
	return c.Redirect(http.StatusFound, "/list")
}

//Edit task
func Edit_task(c echo.Context) error {
	ID := c.Request().FormValue("task_id")
	task := c.Request().PostFormValue("task")
	err := update_task_by_id(ID, "name", task)
	if err != nil {
		log.Fatal("Edit task() ERROR:", err)
	}
	return c.Redirect(http.StatusFound, "/list")
}

//Rediraction to account settings page
func Account_info_page(c echo.Context) error {
	return c.Redirect(http.StatusFound, "/account")
}

//Account information page
func Account_info(c echo.Context) error {
	var user model.User
	user, err := get_user_by_id(c.Get("uid"))
	if err != nil {
		log.Fatal("Error with getting user by id")
	}
	return c.Render(http.StatusOK, "account", map[string]interface{}{
		"First_name": user.First_name,
		"Last_name":  user.Last_name,
		"Email":      user.Email,
	})
}

//Updating user info if needed
func Account_update(c echo.Context) error {
	var found_user, form_user model.User
	uid := c.Get("uid")
	found_user, err := get_user_by_id(uid)
	if err != nil {
		log.Fatal("Error with getting user by id")
	}
	_ = c.Bind(&form_user)
	if found_user.First_name != form_user.First_name {
		err := update_user_field_by_uid(uid, "first_name", *form_user.First_name)
		if err != nil {
			return err
		}
	}
	if found_user.Last_name != form_user.Last_name {
		err := update_user_field_by_uid(uid, "last_name", *form_user.Last_name)
		if err != nil {
			return err
		}
	}
	if found_user.Email != form_user.Email {
		err := update_user_field_by_uid(uid, "email", *form_user.Email)
		if err != nil {
			return err
		}
	}

	return c.Redirect(http.StatusFound, "/list")
}

//Signup form and Handler that render signup template
func Signup_Form(c echo.Context) error { //Get method
	c = clearcookie(c)
	return c.Render(http.StatusOK, "signup", map[string]interface{}{})
}

//Registration method
func Signup(c echo.Context) error { //Post method
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	var user model.User
	var msg Message
	msg.Errors = make(map[string]string)
	msg.Inputed_info = make(map[string]string)

	defer cancel()

	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]error{"error": err})
	}
	if *user.First_name == "" {
		msg.Errors = errors(msg.Errors, "First_name", "First name cannot be empty")
	} else {
		msg.Inputed_info["First_Name"] = *user.First_name
	}
	if *user.Last_name == "" {
		msg.Errors = errors(msg.Errors, "Last_name", "Last name cannot be empty")
	} else {
		msg.Inputed_info["Last_Name"] = *user.Last_name
	}
	if !helper.IsEmailValid(*user.Email) {
		msg.Errors = errors(msg.Errors, "Email", "Email is not valid")
	} else {
		msg.Inputed_info["Email"] = *user.Email
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
		msg.Inputed_info["Email"] = ""
	}
	if confirmation_pass != *user.Password {
		msg.Errors = errors(msg.Errors, "Password", "Passwords did not match")
	}
	if len(msg.Errors) != 0 {
		return c.Render(http.StatusOK, "signup", msg)
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

//Login form and Handler that render login template
func Login_Form(c echo.Context) error { //Get methon for login
	c = clearcookie(c)
	return c.Render(http.StatusOK, "login", map[string]interface{}{})
}

//Post method for login
func Login(c echo.Context) error {
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

//Logout handler for exit
func Logout(c echo.Context) error { //Post for logout
	c = clearcookie(c)
	return c.Redirect(http.StatusFound, "/login")
}

//Bcrypting pass for storing in DB
func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}

	return string(bytes)
}

//Compare entered password and password provided by user. Helper for login handler
func VerifyPassword(userPassword string, providedPassword string) bool {
	fmt.Println()
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true

	if err != nil {
		check = false
	}

	return check
}

//Method for clearing user cookie after logout
func clearcookie(c echo.Context) echo.Context {
	c.SetCookie(&http.Cookie{
		Name:  "token",
		Value: "",
	})
	return c
}

//Update Task in DB by the provided field and value
func update_task_by_id(id string, field string, value interface{}) error {
	update := bson.M{}
	if field == "name" {
		update = bson.M{
			"$set": bson.M{field: value.(string)},
		}
	} else if field == "status" {
		update = bson.M{
			"$set": bson.M{field: value.(bool)},
		}
	}
	upsert := true
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}
	//log.Println("ID ", id, "\nTask ", value)
	_, err := tasksCollection.UpdateOne(
		context.Background(),
		bson.M{"task_id": id},
		update,
		&opt,
	)
	if err != nil {
		log.Fatal("UPDATE() ERROR:", err)
	}
	return err
}

//Function for updating any user field by provided uid
func update_user_field_by_uid(uid interface{}, field, value string) error {
	upsert := true
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}
	update := bson.M{
		"$set": bson.M{field: value},
	}
	_, err := usersCollection.UpdateOne(
		context.Background(),
		bson.M{"user_id": uid},
		update,
		&opt,
	)
	if err != nil {
		log.Fatal("Error with updating user info:", err)
	}
	return nil
}

//Get user info by user id
func get_user_by_id(uid interface{}) (model.User, error) {
	var user model.User
	err := usersCollection.FindOne(context.Background(), bson.M{"user_id": uid}).Decode(&user)
	if err != nil {
		return user, err
	}
	return user, nil
}

//Method for helping to collect error from login or signup handlers
func errors(msg map[string]string, field, msg_error string) map[string]string {
	msg[field] = msg_error
	return msg
}
