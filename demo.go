// LPWA Demo application
package main

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/brocaar/lorawan"
	"github.com/marcc-orange/datavenue"
	"github.com/marcc-orange/datavenue/lpwa"
)

// Application flags
var listen = flag.String("listen", ":8080", "address and port to listen")
var aesKeyFlag = flag.String("appSKey", "", "application session `key`, mandatory")
var devAddrFlag = flag.String("devAddr", "", "device `address`, mandatory")
var dvURL = flag.String("dvURL", "https://api.orange.com/datavenue/v1", "Datavenue URL")
var dvOAPIKey = flag.String("dvOAPIKey", "", "Datavenue Orange Partner `key`, mandatory")
var dvISSKey = flag.String("dvISSKey", "", "Datavenue Primary `key`, mandatory")
var dvDatasource = flag.String("dvDatasource", "", "Datavenue Datasource `ID`, mandatory")
var dvCommandStream = flag.String("dvCommandStream", "", "Datavenue Downlink Command Stream `ID`, mandatory")
var dvDownlinkFCntStream = flag.String("dvFDownlinkFCntStream", "", "Datavenue Downlink FCnt Stream `ID`, mandatory")

var aesKey lorawan.AES128Key
var devAddr lorawan.DevAddr
var dvClient *datavenue.Client

// setup aesKey, devAddr and datavenue client
func setup() (err error) {

	if *aesKeyFlag == "" || *devAddrFlag == "" || *dvOAPIKey == "" || *dvISSKey == "" || *dvDatasource == "" || *dvCommandStream == "" || *dvDownlinkFCntStream == "" {
		log.Fatalln("error: missing mandatory parameters, use -h for help")
	}

	if err = aesKey.UnmarshalText([]byte(*aesKeyFlag)); err != nil {
		return err
	}

	var devAddrBin []byte
	if devAddrBin, err = hex.DecodeString(*devAddrFlag); err != nil {
		return err
	}
	devAddr = lorawan.DevAddr{}
	if err = devAddr.UnmarshalBinary(devAddrBin); err != nil {
		return err
	}

	dvClient = &datavenue.Client{
		URL:     *dvURL,
		OAPIKey: *dvOAPIKey,
		ISSKey:  *dvISSKey,
		Client:  &http.Client{},
	}
	return nil
}

// sendCommand will send the command and return the frame counter
func sendCommand(command string) (uint32, error) {

	// Command is a single byte
	var c [1]byte
	switch command {
	case "off":
		c[0] = 0
	case "on":
		c[0] = 1
	case "blink":
		c[0] = 2
	default:
		return 0, errors.New("bad command")
	}

	// Retrieve the current downlink frame counter
	fCnt, err := lpwa.RetrieveDownlinkFCnt(dvClient, *dvDatasource, *dvDownlinkFCntStream)
	if err != nil {
		return 0, err
	}

	// Encrypt command with current frame counter
	encrypted, err := lorawan.EncryptFRMPayload(aesKey, false, devAddr, fCnt, c[:])
	if err != nil {
		return 0, err
	}

	// Send encrypted command
	return fCnt, lpwa.SendDownlinkData(dvClient, *dvDatasource, *dvCommandStream, encrypted, fCnt, 5, true)
}

// handlePush decrypts the received payload and send light and luminosity to all websockets
func handlePush(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Read JSON payload
	buff, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(400)
		log.Println(err)
		return
	}

	// Decode JSON payload
	dvPayload := &datavenue.Values{}
	if err = json.Unmarshal(buff, dvPayload); err != nil {
		w.WriteHeader(400)
		log.Println(err)
		log.Println("json:", string(buff))
		return
	}

	// Extract data and fCnt from payload
	hexData, ok := dvPayload.Values[0].Value.(string)
	if !ok {
		w.WriteHeader(400)
		log.Println("invalid value, ignoring message")
		return
	}
	fCnt := uint32(dvPayload.Values[0].Metadata["fcnt"].(float64))
	log.Printf("raw: %s fCnt: %d", hexData, fCnt)

	// Decode hex data
	encryptedData, err := hex.DecodeString(hexData)
	if err != nil {
		w.WriteHeader(400)
		log.Println(err)
		return
	}

	// Decrypt data and log it
	data, err := lorawan.EncryptFRMPayload(aesKey, true, devAddr, fCnt, encryptedData)
	if err != nil {
		w.WriteHeader(400)
		log.Println(err)
		return
	}
	log.Printf("dec: %x", data)

	// Check data is 5 bytes, ignore otherwise
	if len(data) != 5 {
		w.WriteHeader(400)
		log.Println("data is not 5 bytes long, size:", len(data))
		return
	}

	// Extract light and luminnosity information from uncrypted payload
	lightOn := data[0] != 0
	luminosity := binary.BigEndian.Uint32(data[1:5])
	log.Printf("light: %v, luminosity: %d", lightOn, luminosity)

	// Send as a notification to all opened websockets
	notifyAllWebsockets(NotificationMessage{
		LightOn:    lightOn,
		Luminosity: luminosity,
	})

	w.WriteHeader(204)
}

// panicWrapper wraps a handler to recover from panic
func panicWrapper(handlerFunc func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		defer func() {
			r := recover()
			if r != nil {
				switch t := r.(type) {
				case string:
					err = errors.New(t)
				case error:
					err = t
				default:
					err = errors.New("Unknown error")
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Println("panic: " + err.Error())
			}
		}()
		handlerFunc(w, r)
	})
}

func main() {
	flag.Parse()

	// Retrieve current bin dir (require absolute path)
	wdir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	if err := setup(); err != nil {
		log.Fatal(err)
	}

	// Send all commands comming from websockets to Datavanue
	registerCommandHandler(func(command CommandMessage) {
		sendCommand(command.Light)
	})

	log.Printf("AppSKey: %x", [16]byte(aesKey))
	log.Printf("DevAddr: %x, %d", [4]byte(devAddr), len(devAddr))

	// Start HTTP service
	log.Println("Starting listening on:", *listen)
	http.Handle("/demo/push", panicWrapper(handlePush))
	http.Handle("/demo/ws", panicWrapper(wsHandler))
	http.Handle("/", http.FileServer(http.Dir(wdir+"/static")))
	log.Fatal(http.ListenAndServe(*listen, nil))
}
