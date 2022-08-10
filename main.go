package main

import (
	"fmt"
	"html/template"
	"net/http"
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type M map[string]interface{}
var ctx = func() context.Context {
	return context.Background()
}()

type the_task struct {
	ID    primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	Task  string `bson:"task"`
	Assignee string    `bson:"assignee"`
	Deadline string `bson:"deadline"`
	IsDone bool `bson:"isdone"` 
}

func connect() (*mongo.Database, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return nil, err
	}

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	return client.Database("List_tasks"), nil
}

func insert(task string, assignee string, deadline string ) {
	db, err := connect()
	if err != nil {
		log.Fatal(err.Error())
	}

	_, err = db.Collection("tasks").InsertOne(ctx, the_task{primitive.NewObjectID(),task, assignee, deadline,false})
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Println("Insert success!")
}

func find() []the_task{
	db, err := connect()
	if err != nil {
		log.Fatal(err.Error())
	}

	csr, err := db.Collection("tasks").Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err.Error())
	}
	defer csr.Close(ctx)
	
	result := make([]the_task, 0)
	for csr.Next(ctx) {
		var row the_task
		err := csr.Decode(&row)
		if err != nil {
			log.Fatal(err.Error())
		}

		result = append(result, row)
	}
	return result
}

func remove(id string) {
	db, err := connect()
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println(id)
	var selector = primitive.ObjectID{"id": id}
	fmt.Println(selector)
	_, err = db.Collection("tasks").DeleteOne(ctx,selector)
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Println("Remove success!")
}

func main() {
	find()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var list=find()
		var data = M{"name": "Batman","list":list}
		var tmpl = template.Must(template.ParseFiles(
			"views/index.html",
			"views/head.html",
			"views/message.html",
		))

		var err = tmpl.ExecuteTemplate(w, "index", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	
	http.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		var data = M{"name": "Batman"}
		var tmpl = template.Must(template.ParseFiles(
			"views/head.html",
			"views/about.html",
			"views/message.html",
		))

		var err = tmpl.ExecuteTemplate(w, "about", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	

	http.HandleFunc("/form", routeIndexGet)
	http.HandleFunc("/result", submit)
	http.HandleFunc("/delete", routeDelete)
	

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("assets"))))

	fmt.Println("server started at localhost:9000")
	http.ListenAndServe(":9000", nil)
}

func routeIndexGet(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        var tmpl = template.Must(template.New("form").ParseFiles(
			"views/head.html",
			"views/form.html",
		))
        var err = tmpl.Execute(w, nil)

        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
        return
    }

    http.Error(w, "", http.StatusBadRequest)
}


func routeSubmitPost(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var tmpl = template.Must(template.New("result").ParseFiles("views/result.html"))

		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var task = r.FormValue("task")
		var assignee = r.Form.Get("assignee")
		var deadline = r.Form.Get("deadline")

		var data = map[string]string{"task": task, "assignee": assignee, "deadline": deadline}

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	http.Error(w, "", http.StatusBadRequest)
}

func submit(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var task = r.FormValue("task")
		var assignee = r.Form.Get("assignee")
		var deadline = r.Form.Get("deadline")

		insert(task, assignee, deadline)
		http.Redirect(w, r, "/", http.StatusFound)
		return 
	}

	http.Error(w, "", http.StatusBadRequest)
}

func routeDelete(w http.ResponseWriter, r *http.Request){
	id := r.URL.Query().Get("id")
	remove(id)
	http.Redirect(w, r, "/", http.StatusFound)
}