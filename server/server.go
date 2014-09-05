package server

import (
        "compress/gzip"
        "fmt"
	"encoding/json"
	"time"
	"net/http"
	"os"
        "io/ioutil"
	"os/signal"
	"os/exec"
	"github.com/janelia-flyem/gojsonschema"
	"math/rand"
	"strconv"
	"strings"
	"syscall"
        "sync"
)

const (
        logfile = "transaction.log"
        h5name = "labels.h5"
        h5namez = "labels.h5.gz"
        graphname = "graph.json"
        exeloc = "neuroproof_graph_analyze_gt"
)

type Server struct {
        outputDir string
        progData string
        transactions Transactions
        httpAddress string
        lock      sync.Mutex // for dumping to log and accessing internal DB 
}

func NewServer(outputdir string, progdata string) (*Server) {
        return &Server{outputDir: outputdir, 
                        progData: progdata,
                        transactions: *NewTransactions()} 
}

// TODO

// test on google vm: hand install go, install package, add link to flyem-recon webpage

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

// extractHTML formats information in neuroproof output into html format
func (s *Server) extractHTML(data map[string]interface{}) map[string]interface{} {
        ret_data := make(map[string]interface{})
        html_str := ""
        
        // load status if status exists
        if data2, ok := data["status"]; ok {
                ret_data["status"] = data2
                if data2.(string) == "Finished" {
                    runtime := data["runtime"].(float64)
                    html_str += "Seconds to complete: " + strconv.FormatFloat(runtime, 'f', 1, 64) + "<br><br>" 
                } else if data2.(string) == "Error" {
                    html_str += "Unknown error computing similarity"
                    ret_data["html-data"] = html_str
                    return ret_data
                }
        }
       
        // print initial statistics from NeuroProof
        if data2, ok := data["start"]; ok {
            synapse_stats := data2.(map[string]interface{})["syn"]
            vol_stats := data2.(map[string]interface{})["vol"]
        
            html_str += "<b>Volume Stats</b><ul>"
            
            volover := vol_stats.(map[string]interface{})["overseg-vi"].(float64)
            volunder := vol_stats.(map[string]interface{})["underseg-vi"].(float64)

            html_str += "<li>Over-seg VI: " + strconv.FormatFloat(volover, 'f', 4, 64) + "</li>"
            html_str += "<li>Under-seg VI: " + strconv.FormatFloat(volunder, 'f', 4, 64) + "</li>"
            html_str += "</ul><br>"
           
            synover := synapse_stats.(map[string]interface{})["overseg-vi"].(float64)
            synunder := synapse_stats.(map[string]interface{})["underseg-vi"].(float64)
            orphans := synapse_stats.(map[string]interface{})["orphans"].(float64)
            
            html_str += "<b>Synapse Stats</b><ul>"
            html_str += "<li>Over-seg VI: " + strconv.FormatFloat(synover, 'f', 4, 64) + "</li>"
            html_str += "<li>Under-seg VI: " + strconv.FormatFloat(synunder, 'f', 4, 64) + "</li>"
            html_str += "<li>Orphan synapses: " + strconv.FormatFloat(orphans, 'f', 0, 64) + "</li>"
            html_str += "</ul><br><br>"
        }

        // print proofreading statistics from NeuroProof
        if data2, ok := data["proofreading-work"]; ok {
            html_str += "<b>Automatic Focused Proofreading Performed</b><ul>"
            vol_examined := data2.(map[string]interface{})["vol"].(map[string]interface{})["edges-examined"].(float64)
            html_str += "<li>Number of focused volume decisions: " + strconv.FormatFloat(vol_examined, 'f', 0, 64) + "</li>" 
            syn_examined := data2.(map[string]interface{})["syn"].(map[string]interface{})["edges-examined"].(float64)
            html_str += "<li>Number of focused synapse decisions: " + strconv.FormatFloat(syn_examined, 'f', 0, 64) + "</li>"
            html_str += "</ul><br><br>"
        }

        // print initial statistics from NeuroProof
        if data2, ok := data["final"]; ok {
            synapse_stats := data2.(map[string]interface{})["syn"]
            vol_stats := data2.(map[string]interface{})["vol"]
        
            html_str += "<b>Volume Stats</b><ul>"
            
            volover := vol_stats.(map[string]interface{})["overseg-vi"].(float64)
            volunder := vol_stats.(map[string]interface{})["underseg-vi"].(float64)

            html_str += "<li>Over-seg VI: " + strconv.FormatFloat(volover, 'f', 4, 64) + "</li>"
            html_str += "<li>Under-seg VI: " + strconv.FormatFloat(volunder, 'f', 4, 64) + "</li>"
            html_str += "</ul><br>"
           
            synover := synapse_stats.(map[string]interface{})["overseg-vi"].(float64)
            synunder := synapse_stats.(map[string]interface{})["underseg-vi"].(float64)
            
            html_str += "<b>Synapse Stats</b><ul>"
            html_str += "<li>Over-seg VI: " + strconv.FormatFloat(synover, 'f', 4, 64) + "</li>"
            html_str += "<li>Under-seg VI: " + strconv.FormatFloat(synunder, 'f', 4, 64) + "</li>"
            html_str += "</ul><br><br>"
        }

        ret_data["html-data"] = html_str
        return ret_data
}

// launch is a separate process that call neuroproof and updates log
func (s *Server) launchJob(session_id string, session_dir string, remoteaddr string) {
        // delete entire directory (output is saved to log)
        defer os.RemoveAll(session_dir) 
        
        // build np string
        argument_arr := make([]string, 0)
        argument_arr = append(argument_arr, session_dir + "/" + h5name) 
        argument_arr = append(argument_arr, s.progData + "/groundtruth.h5") 
        argument_arr = append(argument_arr, "--graph-file") 
        argument_arr = append(argument_arr, session_dir + "/" + graphname) 
        argument_arr = append(argument_arr, "--recipe-file") 
        argument_arr = append(argument_arr, s.progData + "/recipe.json") 
        argument_arr = append(argument_arr, "--callback-uri") 
        argument_arr = append(argument_arr, "http://" + s.httpAddress + "/status/" + session_id) 
        argument_arr = append(argument_arr, "--synapse-file") 
        argument_arr = append(argument_arr, s.progData + "/synapses.json") 
        argument_arr = append(argument_arr, "--dump-split-merge-bodies") 
        argument_arr = append(argument_arr, "1") 
 
        // start timer
        startstamp := time.Now()
        //time1 := int(startstamp.Unix())

        output, _ := exec.Command(exeloc, argument_arr...).Output()
       
        endstamp := time.Now()
        //time2 := int(endstamp.Unix())

        // grab results
        //data, found := s.transactions.getTran(session_id)
        //if !found {
        //        panic("Fatal error in transaction dictionary")
        //}

        // update global log -- time and command; dump map entry into log
        s.lock.Lock()
        fout, err := os.OpenFile(s.outputDir + "/" + logfile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0664)

        if err != nil {
                panic("Cannot open log file: " + s.outputDir + "/" + logfile)
        }

        // write time stamp and neuroproof dump
        fout.WriteString("**Started " + session_id + ": " + startstamp.Format(time.ANSIC) + "\n") 
        fout.WriteString(remoteaddr + "\n")  
        fout.WriteString(string(output))
        fout.WriteString("**Finished " + session_id + ": " + endstamp.Format(time.ANSIC) + "\n")
        fout.Close()
        s.lock.Unlock()        

        data, found := s.transactions.getTran(session_id)
        if !found {
            datatemp := make(map[string]interface{})
            datatemp["status"] = "Error"
            s.transactions.updateTran(session_id, datatemp)
        } else {
            if val, found2 := data["status"]; !found2 || val.(string) != "Finished" {
                    data["status"] = "Error"
                    s.transactions.updateTran(session_id, data)
            }
        }

}





// frontHandler handles GET requests to "/"
func (s *Server) frontHandler(w http.ResponseWriter, r *http.Request) {
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
func (s *Server) formHandler(w http.ResponseWriter, r *http.Request) {
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
        session_dir := s.outputDir + "/" + session_id
	err = os.MkdirAll(session_dir, 0755)
        if err != nil {
                badRequest(w, "Cannot create temporary directory")
                return
        }

	h5data, _, err := r.FormFile("h5file")
        if err != nil {
                badRequest(w, "gzip h5 file not provided")
        }
        //bytes, err := ioutil.ReadAll(h5data)
        zreader, err := gzip.NewReader(h5data)
        if err != nil {
                badRequest(w, "Could not read gzip h5 file")
                return
        }
        defer zreader.Close()
        
        //bytes := make([]byte, 0)
        //nread, err := zreader.Read(bytes)
        bytes, err := ioutil.ReadAll(zreader)
        if err != nil {
                badRequest(w, "Could not read gzip h5 file")
                os.RemoveAll(session_dir) 
                return
        }

        ioutil.WriteFile(session_dir + "/" + h5name, bytes, 0644) 

	graphdata, _, err := r.FormFile("graphfile")
        if err != nil {
                badRequest(w, "json file not provided")
                os.RemoveAll(session_dir) 
                return
        }
        bytes2, err := ioutil.ReadAll(graphdata)
        if err != nil {
                badRequest(w, "Could not be read graph json")
                os.RemoveAll(session_dir) 
                return
        }
     
        // convert schema to json data
	var schema_data interface{}
	json.Unmarshal([]byte(graphSchema), &schema_data)

        // convert json bytes to json object	
	var graph_json interface{}
        json.Unmarshal(bytes2, &graph_json)
	
        // validate json schema
	schema, err := gojsonschema.NewJsonSchemaDocument(schema_data)
	validationResult := schema.Validate(graph_json)
	if !validationResult.Valid() {
		badRequest(w, "JSON did not pass validation")
                os.RemoveAll(session_dir) 
		return
	}
        ioutil.WriteFile(session_dir + "/" + graphname, bytes2, 0644) 

        // write initial status
        status := make(map[string]interface{})
        status["status"] = "Started" 
        s.transactions.updateTran(session_id, status)

        // launch job
        go s.launchJob(session_id, session_dir, r.RemoteAddr)

        // dump json callback
	w.Header().Set("Content-Type", "application/json")
	jsondata, _ := json.Marshal(map[string]interface{}{
		"status-callback": "/status/" + session_id,
	})
	fmt.Fprintf(w, string(jsondata))

}
// formHandler handles post request to "/status" from the web interface
func (s *Server) statusHandler(w http.ResponseWriter, r *http.Request) {
	pathlist, requestType, err := parseURI(r, "/status/")
	if err != nil || len(pathlist) != 1 {
		badRequest(w, "Error: incorrectly formatted request")
		return
	}
	if requestType != "post" && requestType != "get" {
		badRequest(w, "only supports get and post")
		return
        }

        if requestType == "post" {
            // read json
            decoder := json.NewDecoder(r.Body)
            var json_data map[string]interface{}
            err = decoder.Decode(&json_data)

            if err != nil {
                    badRequest(w, "unknown put error")
                    return
            }

            // load json
            s.lock.Lock()
            s.transactions.updateTran(pathlist[0], json_data)
            s.lock.Unlock()
        } else if requestType == "get" {
            data, found := s.transactions.getTran(pathlist[0])
            if !found {
                    badRequest(w, "transaction id not found")
            }
            datahtml := s.extractHTML(data)
            w.Header().Set("Content-Type", "application/json")
            jsondata, _ := json.Marshal(datahtml)
            fmt.Fprintf(w, string(jsondata))    
        }
}

// serves files from /static
func (s *Server) staticHandler(w http.ResponseWriter, r *http.Request) {
	pathlist, requestType, err := parseURI(r, "/static/")
	if err != nil || len(pathlist) != 1 {
		badRequest(w, "Error: incorrectly formatted request")
		return
	}
	if requestType != "get" {
		badRequest(w, "only supports get")
		return
        }

        if pathlist[0] == "graph.json" {
            w.Header().Set("Content-Type", "application/json")        
            bytes, _ := ioutil.ReadFile(s.progData + "/graph0.json")
            fmt.Fprintf(w, string(bytes)) 
        }
        if pathlist[0] == "graph.schema.json" {
            w.Header().Set("Content-Type", "application/json")        
            fmt.Fprintf(w, graphSchema) 
        }
        if pathlist[0] == "labels.h5.gz" {
            w.Header().Set("Content-Type", "application/octet-stream")        
            bytes, _ := ioutil.ReadFile(s.progData + "/labels0.h5.gz")
            w.Write(bytes)
        }
        if pathlist[0] == "grayscale.tgz" {
            w.Header().Set("Content-Type", "application/octet-stream")        
            bytes, _ := ioutil.ReadFile(s.progData + "/grayscale_maps.tgz")
            w.Write(bytes)
        }       
}


// Serve is the main server function call that creates http server and handlers
func (s *Server) Serve() {
	//hname, _ := os.Hostname()
	//webAddress := "http://23.251.159.133:80" //+ ":" + strconv.Itoa(80)
	hname, _ := os.Hostname()
	webAddress := hname + ":" + strconv.Itoa(8000)

        s.httpAddress = webAddress
	fmt.Printf("Web server address: %s\n", webAddress)
	fmt.Printf("Running...\n")

	httpserver := &http.Server{Addr: webAddress}

	// front page containing simple form
	http.HandleFunc("/", s.frontHandler)

        // handle gets and puts to status
	http.HandleFunc("/status/", s.statusHandler)

	// handle form inputs
	http.HandleFunc("/formhandler/", s.formHandler)


        // handle static gets
	http.HandleFunc("/static/", s.staticHandler)

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
