package server

// String representing html for simple service access
const formHTML=`
<html>                                                                                                   
<head>
<script src="http://code.jquery.com/jquery-1.11.0.min.js"></script>                                      
<script type="text/javascript" src="//www.google.com/jsapi"></script>

<script>
var status_location = "";
</script>
</head>

<table>
<tr>
<td><H2>Evaluate EM Segmentation</H2></td>
<td><a href="http://janelia.org"><img src="https://raw.github.com/janelia-flyem/janelia-flyem.github.com/master/images/gray_janelia_logo.png"/></a>
</tr>
</table>

<a href="http://www.janelia.org/team-project/fly-em">Fly EM Reconstruction Team</a>
<br>

metrics<span id="metricsheader" style="cursor:pointer;">+</span>
<div id="metrics" style="display:none">Here are some metrics</div>
<br>

instructions<span id="instructionsheader" style="cursor:pointer;">+</span>
<div id="instructions" style="display:none">
<a href="/static/graph.json">Sample graph file</a>
<a href="/static/labels.h5.gz">Sample zipped h5 file</a>
<a href="/static/grayscale.tgz">Directory of grayscale image stack</a>
</div>

<hr>

<br>
<table border=1>
Submit Segmentation Job<br>
<form id="calclabels" method="post">
<input type="file" name="h5file" id="h5file" accept=".gz"><br>
<input type="file" name="graphfile" id="graphfile" accept=".json"><br>
<input type="submit" id="submitbut" value="Submit"/><br>
</form>
</table>

<hr>
<hr>

<br>
<div id="status"></div><br>
<div id="results"></div><br>
</div>

<script>
    setInterval(loadUpdate, 1000);
    $("#metricsheader").click(function() {
        $("#metrics").slideToggle();
        if ($("#metricsheader").text() == "+") {
            $("#metricsheader").html("&#8722")
        } else {
            $("#metricsheader").html("+")
        }
    });

    $("#instructionsheader").click(function() {
        $("#instructions").slideToggle();
        if ($("#instructionsheader").text() == "+") {
            $("#instructionsheader").html("&#8722")
        } else {
            $("#instructionsheader").html("+")
        }
    });

    $("#calclabels").submit(function(event) {                                                           
      event.preventDefault();

      var formData = new FormData();
      
      // load h5
      var x = document.getElementById("h5file");
      if (x.files[0] === undefined) {
            alert("Must provide h5 file");
            return;
      }    
      if (x.files[0].size > 2000000) {
            alert("H5 file is too big");
            return;
      }
      formData.append("h5file", x.files[0]);
      
      // load graph
      var y = document.getElementById("graphfile");
      if (y.files[0] === undefined) {
            alert("Must provide graph file");
            return;
      }    
      if (y.files[0].size > 4000000) {
            alert("Graph file is too big");
            return;
      }
      formData.append("graphfile", y.files[0]);
      
      $('#status').html("Uploading...");
      $('#results').html("");

       $.ajax({
        type: "POST",
        url: "/formhandler/",
        data: formData,
        contentType: false,
        processData: false,
         success: function(data){
            status_location = data["status-callback"];
            $('#status').html("Processing...");
            document.getElementById("submitbut").disabled = true;
        },
        error: function(msg) {
                $('#status').html("Error Processing Results: " + msg.responseText);
          }
        });
    });
    function loadUpdate() {
        if (status_location != "") {
            $.ajax({
                type: "GET",
                url: status_location,
                success: function(data){
                    status = data["status"];
                    $('#status').html("Status: <b>" + status + "</b>");
                                   
                    // grab html string
                    results = data["html-data"];
                    $('#results').html(results);
                
                    if (status == "Finished") {
                        status_location = "";
                        document.getElementById("submitbut").disabled = false;
                    }
                },
                error: function(msg) {
                    $('#results').html("Error accessing server");
                }
            });
        }      
    }                                                                                                
</script>                                                                                                
</html>                                  
`
