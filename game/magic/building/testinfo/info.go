package main

import (
    "log"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/building"
)

func main(){
    cache := lbx.AutoCache()
    infos, err := building.ReadBuildingInfo(cache)
    if err != nil {
        log.Printf("Error: %v", err)
    } else {
        for i, info := range infos {
            log.Printf("Building %v: %+v", i, info)
        }
    }
}
