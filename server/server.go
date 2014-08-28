package server

import (
        "fmt"
	"net/http"
	"os"
        "io/ioutil"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
        "sync"
)

const (
        logfile = "transaction.log"
        h5prefix = "labels-"
        graphprefix = "graph-"
)

type Server struct {
        outputDir string
        exePath   string
        transactions map[string]map[string]interface{}
        lock      sync.Mutex // for dumping to log and accessing internal DB 
}

func NewServer(outputdir string, exepath string) (*Server) {


}


// TODO
// mechanism for random assignments
// callback address for status: get returns json with status and html

// go routine will access locked map that holds info; stats and other stuff should be written to a log -- need lock; go routine should also have IP address
// go thread that deletes temporary files after finishing

// callback address for post: take status -- which could be error, put log in list, add other stuff in a formatted way
// add neuroproof html status option -- write status, load info as computed in specialized fields -- create for best

// add basic metric info in html
// add instructions including sample files and json schema
// add schema for graph to provide immediate server error message

// (optional): show results against current (or best of the day)
// (optional): pretty format of data (or just format each item) -- might need something if there are any charts like histogram


// badRequest is a halper for printing an http error message
func badRequest(w http.ResponseWriter, msg string) {
	fmt.Println(msg)
	http.Error(w, msg, http.StatusBadRequest)
}

// parseURI is a utility function for retrieving parts of the URI
func parseURI(r *http.Request, prefix string) ([]string, string, error) {
	requestType := strings.ToLower(r.Method)
	prefix = strings.Trim(prefix, "/")
	path := strings.Trim(r.URL.Path, "/")
	prefix_list := strings.Split(prefix, "/")
	url_list := strings.Split(path, "/")
	var path_list []string

	if len(prefix_list) > len(url_list) {
		return path_list, requestType, fmt.Errorf("Incorrectly formatted URI")
	}

	for i, val := range prefix_list {
		if val != url_list[i] {
			return path_list, requestType, fmt.Errorf("Incorrectly formatted URI")
		}
	}

	if len(prefix_list) < len(url_list) {
		path_list = url_list[len(prefix_list):]
	}

	return path_list, requestType, nil
}

// randomHex computes a random hash for storing service results
func randomHex() (randomStr string) {
    randomStr = ""
    for i := 0; i < 8; i++ {
        val := rand.Intn(16)
        randomStr += strconv.FormatInt(int64(val), 16)
    }
    return
}


// frontHandler handles GET requests to "/"
func frontHandler(w http.ResponseWriter, r *http.Request) {
	pathlist, requestType, err := parseURI(r, "/")
	if err != nil || len(pathlist) != 0 {
		badRequest(w, "Error: incorrectly formatted request")
		return
	}
	if requestType != "get" {
		badRequest(w, "only supports gets")
		return
	}
	w.Header().Set("Content-Type", "text/html")

	formHTMLsub := strings.Replace(formHTML, "DEFAULT", "", 1)
	fmt.Fprintf(w, formHTMLsub)
}


// formHandler handles post request to "/formhandler" from the web interface
func formHandler(w http.ResponseWriter, r *http.Request) {
	pathlist, requestType, err := parseURI(r, "/formhandler/")
	if err != nil || len(pathlist) != 0 {
		badRequest(w, "Error: incorrectly formatted request")
		return
	}
	if requestType != "post" {
		badRequest(w, "only supports posts")
		return
	}

        session_id := randomHex()
        tstamp := int(time.Now().Unix())
        session_id = session_id + "-" + strconv.Itoa(tstamp)


	h5data, _, _ := r.FormFile("h5file")
        bytes, err := ioutil.ReadAll(h5data)
        if err != nil {
                badRequest(w, "Could not be read")
                return
        }
        ioutil.WriteFile("temp.h5", bytes, 0644) 

	graphdata, _, _ := r.FormFile("graphfile")
        bytes2, err := ioutil.ReadAll(graphdata)
        if err != nil {
                badRequest(w, "Could not be read")
                return
        }
        ioutil.WriteFile("temp.json", bytes2, 0644) 
}


// Serve is the main server function call that creates http server and handlers
func (*Server) Serve() {
	//hname, _ := os.Hostname()
	//webAddress := "http://23.251.159.133:80" //+ ":" + strconv.Itoa(80)
	hname, _ := os.Hostname()
	webAddress := hname + ":" + strconv.Itoa(8000)


	fmt.Printf("Web server address!!!: %s\n", webAddress)
	fmt.Printf("Running...\n")

	httpserver := &http.Server{Addr: webAddress}

	// front page containing simple form
	http.HandleFunc("/", frontHandler)

	// handle form inputs
	http.HandleFunc("/formhandler/", formHandler)

	// exit server if user presses Ctrl-C
	go func() {
		sigch := make(chan os.Signal)
		signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)
		<-sigch
		fmt.Println("Exiting...")
		os.Exit(0)
	}()

	httpserver.ListenAndServe()
}
