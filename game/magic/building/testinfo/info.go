package main

import (
    "log"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/building"
)

func main(){
    cache := lbx.AutoCache()
    err := building.ReadBuildingInfo(cache)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
