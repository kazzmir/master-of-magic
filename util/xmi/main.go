package main

import (
    "os"
    "time"
    "strings"
    "fmt"
    "path/filepath"
    "log"

    "github.com/kazzmir/master-of-magic/lib/xmi"

    "gitlab.com/gomidi/midi/v2/smf"
    midiDrivers "gitlab.com/gomidi/midi/v2/drivers"
    // sudo apt install libportmidi-dev
    _ "gitlab.com/gomidi/midi/v2/drivers/portmididrv"
    "gitlab.com/gomidi/midi/v2"
)

/* first run in a terminal
 * $ fluidsynth --audio-driver=pulseaudio /usr/share/sounds/sf2/FluidR3_GM.sf2
 */
func playMidi(smfObject *smf.SMF){
    driver := midiDrivers.Get()
    fmt.Printf("Got driver: %v\n", driver)
    outs, err := driver.Outs()
    if err != nil {
        fmt.Printf("Could not get midi output ports: %v\n", err)
    } else {
        fmt.Printf("Got midi output ports: %v\n", outs)
        if len(outs) > 0 {
            for _, out := range outs {

                if strings.Contains(out.String(), "Through"){
                    continue
                }

                send, err := midi.SendTo(out)
                if err != nil {
                    fmt.Printf("Could not send to midi output port: %v\n", err)
                    return
                }

                defer out.Close()

                for _, event := range smfObject.Tracks[0] {
                    fmt.Printf("Sending event: %v\n", event)
                    err := send(event.Message.Bytes())
                    if err != nil {
                        fmt.Printf("Error: %v\n", err)
                    }

                    // FIXME: use proper delay
                    time.Sleep(time.Millisecond * time.Duration(event.Delta) * 10)
                }

                return
            }

            fmt.Printf("No playable output ports available!\n")

        } else {
            fmt.Printf("No midi output ports available!\n")
        }
    }
}

func dumpMidi(smfObject *smf.SMF){
    for _, event := range smfObject.Tracks[0] {
        fmt.Printf("Sending event: %v\n", event)
    }
}

func replaceExtension(file string, ext string) string {
    return strings.TrimSuffix(file, filepath.Ext(file)) + "." + ext
}

func main(){
    if len(os.Args) < 2 {
        return
    }
    file := os.Args[1]

    data, err := os.Open(file)
    if err != nil {
        fmt.Printf("Error: %s\n", err)
        return
    }
    defer data.Close()

    midi, err := xmi.ConvertToMidi(data)
    
    if err != nil {
        log.Printf("Error: %v", err)
        return
    }

    outputName := replaceExtension(file, "cmid")

    out, err := os.Create(outputName)
    if err != nil {
        log.Printf("Error: %v", err)
        return
    }
    midi.WriteTo(out)
    out.Close()

    log.Printf("Wrote to %v", outputName)

    dumpMidi(midi)

    // playMidi(smfObject)
}
