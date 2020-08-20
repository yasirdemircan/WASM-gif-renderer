package main

import (
    b64 "encoding/base64"
    "fmt"
    "image"
    "bytes"
"image/png"
"image/gif"
    "image/draw"
    "image/color/palette"
    "strings"
    "syscall/js"
    "sync"
	"strconv"
)

//Initialization function for WASM (runs before main)
func init(){
    fmt.Println("Hello WASM")
}
//Getting variable references from Javascript. (Must be set at Javascript to run the function properly!)
var jsval =js.Global().Get("base64array")
var threadEnd = js.Global().Get("threadEnd")
var arraylength int = 0
var b64slice []string
//Creating a new WaitGroup for async tasks
var waitgroup sync.WaitGroup
//Getting JavaScript values and updating them manually. (This function is connected to sendArray reference at Javascript)
func getJsVal(this js.Value, p []js.Value)interface{} {
    jsval :=js.Global().Get("base64array")
    arraylength = js.Global().Get("base64arraylen").Int()
    for index := 0; index < arraylength; index++ {
        
        b64slice = append(b64slice,jsval.Index(index).String())
    }
    fmt.Println("Array sent!")
    return jsval
}
// Function for starting render. (This function is connected to same name reference at Javascript)
func startRender(this js.Value, p []js.Value) interface{}{
//Creating a new gif with Gif type.
    outGif := &gif.GIF{}
//Creating threads in WorkGroup according to frame count at Javascript
    waitgroup.Add(arraylength)
    fmt.Println(strconv.Itoa(arraylength)+" waitgroup count")
//Starting threads
 for index := 0; index < arraylength; index++ {
 go renderProcess(outGif,index,&waitgroup)
 }
 //Waiting syncronization with threads!
 waitgroup.Wait()
 //Creating a writer for exporting Gif.
 outWriter := new(bytes.Buffer)
 //Encoding and sending gif to writer for export.
 gif.EncodeAll(outWriter, outGif)
 var gifString = outWriter.String()
 //Get render done function reference from Javascript.
 renderDone := js.Global().Get("renderDone")
 //Base64 encoding the created binary gif
 sEnc := b64.StdEncoding.EncodeToString([]byte(gifString))
 //Invoking renderDone function with base64 encoded gif.
 renderDone.Invoke(sEnc)

return gifString
}
//Actual render function runs in threads .(With parameters:Gif object,incoming for index,waitgroup name and type)
func renderProcess(outGif *gif.GIF, index int,waitgroup *sync.WaitGroup){
    var indexString string = strconv.Itoa(index)
    fmt.Println("Thread Number: "+indexString+" started.")
    //Getting PNG images from base64array with Javascript reference.
    base64val,_ := b64.StdEncoding.DecodeString(b64slice[index])//jsval.Index(index).String()
    //Converting base64 to string again??
    base64str := string(base64val)
    //Creating reader from string to feed decoder.
         reader := strings.NewReader(base64str)
         //Decoding reader as png image. (No longer a png)
         inGif, _ := png.Decode(reader)
         //Getting image bounds.
         bounds := inGif.Bounds()
         //Creating new palette for gif requirements.
          pimg := image.NewPaletted(bounds, palette.Plan9[:256])
          //Drawing image according to palette.
          draw.Draw(pimg, pimg.Rect, inGif, bounds.Min, draw.Over)
          //Adding paletted image to gif.
         outGif.Image = append(outGif.Image, pimg)//.(*image.Paletted)
         //Frame delay for each frame. (0 at the moment)
         outGif.Delay = append(outGif.Delay, 0)
         // Telling thread end to Javascript.
         threadEnd.Invoke("Thread Number: "+indexString+" ended.",index)
         //Telling the WaitGroup this thread is done working !IMPORTANT!
         waitgroup.Done()
          
}

func main() {
  //Creating an infinite channel to keep application alive.
     c := make(chan bool)
     //Referencing inner functions to Javascript.
    js.Global().Set("sendArray", js.FuncOf(getJsVal))
    js.Global().Set("startRender", js.FuncOf(startRender))
    //Console logging needed information.
    fmt.Println("Wasm is ready!")
    fmt.Println("Set variable (base64array) and (base64arraylen)")
    fmt.Println("sendArray and startRender functions are defined!")
    fmt.Println("define threadEnd function for notifications!")
 //Ending the application channel.
    <-c

}