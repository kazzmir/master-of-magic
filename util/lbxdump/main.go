package main

import (
    "os"
    "fmt"
)

func main(){
    if len(os.Args) < 2 {
        fmt.Println("Give an lbx file, or a zip file and the name of an lbx file inside it")
        return
    }
}
