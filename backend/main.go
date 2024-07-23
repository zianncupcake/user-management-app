package main

//Used to interact with databases using SQL queries.
//Used to convert Go data structures to JSON and vice versa.
//Used for logging messages, errors, etc.
//Used to build web servers and handle HTTP requests.

//Used to create more flexible and sophisticated HTTP routers.
//The underscore (_) before the import path indicates that the package is imported solely for its side effects. github.com/lib/pq is a PostgreSQL driver for Go's database/sql package.
import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)
type User struct {
	Id 		int		`json:"id"`	
	Name	string	`json:"name"`
	Email	string	`json:"email"`
}

//main function
func main() {
	//1. connect to database
	//opens a connection to a postgresql database.
	//postgres: specifies database driver
	//os....:fetch database URL from environment variables, which contains connection details
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	//ensures that database connection is closed when the main function exists
	defer db.Close()

	//2. create table if doesnt exists
	//executes sql statement to create a table named users with 3 cols
	//id serial primary key: id is an auto incrementing pri key
	//the rest are text fields
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, name TEXT, email TEXT)")
	if err != nil {
		log.Fatal(err)
	}

	//3. create router
	//creates new router using gorilla mux package
	router := mux.NewRouter()
	//register new route with the router.
	//listen for get requests at the path /api/gp/users
	//getUsers(db) is a handler function that will process requests to this route. db passed inside to allow database interaction within the handler
	router.HandleFunc("/api/go/users", getUsers(db)).Methods("GET")
	router.HandleFunc("/api/go/users", createUser(db)).Methods("POST")
	router.HandleFunc("/api/go/users/{id}", getUser(db)).Methods("GET")
	router.HandleFunc("/api/go/users/{id}", updateUser(db)).Methods("PUT")
	router.HandleFunc("/api/go/users/{id}", deleteUser(db)).Methods("DELETE")

	//wrap the router with the cors and json content type middlewares --> combine multiple middleware functions to create an enhanced router
	enhancedRouter := enableCORS(jsonContentTypeMiddleWare(router))

	//start server
	log.Fatal(http.ListenAndServe(":8000", enhancedRouter))
}

//params: a pointer to an sql.DB instance, representing the connection to the database
//*means a pointer
func getUsers(db *sql.DB) http.HandlerFunc {
	//handles http request to get a alist of users from the database and send it back as a json response
	return func(w http.ResponseWriter, r *http.Request) {
		//execute sql query that is expected to return a single row. typically used for queries that return a single result like retireving a specific row from a table --> return type is *sql.row
		rows, err := db.Query("SELECT * FROM users")
		if err != nil {
			log.Fatal(err)
		}
		//ensure database rows are closed properly after the function completes, preventing resource leaks
		defer rows.Close()

		//initialise empty slice
		users := []User{}
		//iterate through each row
		for rows.Next() {
			var u User
			//scan: a method of the sql.row type. scan cols of curr row into fields of the user struct
			//& used to pass the memory of addresses
			// u need addresses because we are directly modifying the original variables. not copies
			//syntax to make it concise. assign err to rows.scan output. if there is error then log fatal
			if err := rows.Scan(&u.Id, &u.Name, &u.Email); err != nil {
				log.Fatal(err)
			}
			users = append(users, u)
		}
		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}
		//encodes users slice as json and write it to the response. 
		//json encoder: convert go data structures to json. json decoder: convert json data to go data structures
		//newencoder: function from json package that creates a new encoder which writes to w. w is a http.responsewriter, a type of net/http package that allows u to construct a http response
		//a method on the encoder type that encodes users as jsopn and writes it to the output stream aka the http response w
		json.NewEncoder(w).Encode(users)
	}
}

func createUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u User
		//r.body: body of the http request, contians data sent by client
		//&u: decoded data is stored in the address of u
		//&: address operator, used to get memory address of a variable. because u need to provide a pointer to the struct so that the decoder can directly modify the original struct
		json.NewDecoder(r.Body).Decode(&u)

		//insert new row into users table with the specified name and email values.
		//returning id: postresql feature that return the id of the newly inserted row
		//scan: take pointers to variables where the results of the query will be stored. result of the returning id part of the sql query will be stored in u.id, scan writes the value directly into this field
		err := db.QueryRow("INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id", u.Name, u.Email).Scan(&u.Id)
		if err != nil {
			log.Fatal(err)
		}
		json.NewEncoder(w).Encode(u)
	}
}

func getUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//extract request path parameters and return them as map where keys are the name of the url params and values are the corresponding parts of the url
		vars := mux.Vars(r)
		//extract id
		id := vars["id"]

		var u User
		//$ means placeholder. the number 1 means the first placeholder
		err := db.QueryRow("SELECT * FROM USERS WHERE id = $1", id).Scan(&u.Id, &u.Name, &u.Email)
		if err != nil {
			//if user not found, respond with 404 not found status
			w.WriteHeader(http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(u)
	}
}

func updateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u User
		json.NewDecoder(r.Body).Decode(&u)

		//retrieve id
		vars := mux.Vars(r)
		id := vars["id"]

		//execute update query, exec means it does not return any rows. used for queries that modify the database, such as insert/update/delete --> return type is sql.result
		_, err := db.Exec("UPDATE users SET name = $1, email = $2 WHERE id = $3", u.Name, u.Email, id)
		if err != nil {
			log.Fatal(err)
		}

		//retrieve the updated user data from the database
		var updatedUser User
		err = db.QueryRow("SELECT id, name, email FROM users WHERE id = $1", id).Scan(&updatedUser.Id, &updatedUser.Name, &updatedUser.Email)
		if err != nil {
			log.Fatal(err)
		}
		json.NewEncoder(w).Encode(updatedUser)

	}
}

func deleteUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u User

		//retrieve id
		vars := mux.Vars(r)
		id := vars["id"]

		err := db.QueryRow("SELECT * FROM users WHERE id = $1", id).Scan(&u.Id, &u.Name, &u.Email)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			_, err := db.Exec("DELETE FROM users WHERE id = $1", id)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
			}
			json.NewEncoder(w).Encode("User deleted")
		}


	}
}

//explanation on http headers and content-type
//http headers are key value pairs sent between the client and the server with http requests and responses. provide metadata about the request or reponse e.g. content type, length, encoding
//content-type header indicates the media type of the resource being sent to the client (web browser / mobile app...). when client receives response, it looks at the content-type header to determine how to interpret the response body 
//when u set content-type header to application/json, u are telling the client that the reponse body contains json data

//adds headers to the response to enable cors. allows api to be accessed from web pages hosted on different domains, which is essential for modern web applications that interact with apis
//params: next of type http.handler, return value of type http.handler
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//set cors headers --> set http headers for the response
		w.Header().Set("Access-Control-Allow-Origin", "*") //Allow requests from any origin
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS") //Specifies allowed http methods
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type") //specifies allowed headers

		//check if the request is for cors preflight
		//check if http method is options --> determine if actual request is safe to send
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		//pass down the request to the next middleware or final handler
		next.ServeHTTP(w,r)
	})
}

//middleware that ensures the response content type is set to json. wraps around main request handler to perform some pre/post processing on the request amd and the response
//ensure content-type-header is set to application/json --> ensures that clients know the response body is formatted as json
func jsonContentTypeMiddleWare(next http.Handler) http.Handler {
	//create custom handlers from anonymouys functions
	//w: interface used to contruct the http response. use it to write data to the response body, set http status codes, and set headers
	//r: pointer to a http.request object, which represents the incoming http request. contains information like request method, url, headers, body
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//set json content type
		w.Header().Set("Content-Type", "application/json")
		//call next handler in the chain
		//w: response writer
		//r: request
		next.ServeHTTP(w,r)
	})
}

