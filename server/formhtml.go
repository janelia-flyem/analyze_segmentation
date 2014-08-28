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
<div id="instructions" style="display:none">Here are some instructions</div>

<hr>

<br>
<table border=1>
Submit Segmentation Job<br>
<form id="calclabels" method="post">
<input type="file" name="h5file" id="h5file" accept=".h5"><br>
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
      $('#status').html("");
      $('#results').html("");

      var formData = new FormData();
      var x = document.getElementById("h5file");
      formData.append("h5file", x.files[0]);
      var y = document.getElementById("graphfile");
      formData.append("graphfile", y.files[0]);

       $.ajax({
        type: "POST",
        url: "/formhandler/",
        data: formData,
        contentType: false,
        processData: false,
         success: function(data){
            status_location = data["status-callback"];
            $('#status').html("Starting...");
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
                    $('#status').html(status);
                                   
                    // grab html string
                    results = data["results"]; 
                    $('#results').html(results);
                
                    if (status == "finished") {
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
