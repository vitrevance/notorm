# notorm
 Not ORM for Golang SQL interactions
## Overview
A Go library that provides a Scan function similar to SQL Rows.Scan, but utilizes reflection to automatically map fields into structures based on their names.
## Usage Example
```go
package main

import (
    "database/sql"
    "fmt"

    _ "github.com/lib/pq" // PostgreSQL driver example
    "github.com/vitrevance/notorm"
)

type User struct {
    ID       int    `column:"id"`
    Username string `column:"username"`
}

func main() {
    db, err := sql.Open("postgres", "user=myuser dbname=mydb")
    
    if err != nil {
        fmt.Println(err)
        return
    }
    
    defer db.Close()
    
    rows, err := db.Query("SELECT id, username FROM users")
    
    if err != nil {
        fmt.Println(err)
        return
    }
    
    defer rows.Close()
    
    var users []User

    for rows.Next() {
        var user User
        
        if err := notorm.Scan(&user, rows); err != nil { 
            fmt.Println(err)
            break 
        }
        
        users = append(users, user)
        
        fmt.Printf("User: %+v\n", user)
    }
}
```